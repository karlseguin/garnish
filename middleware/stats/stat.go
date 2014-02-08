package stats

import (
	"github.com/karlseguin/garnish"
	"math"
	"math/rand"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

type Stat struct {
	sync.Mutex
	*Configuration
	hits     int64
	oks      int64
	errors   int64
	failures int64
	samples  []int
	snapshot Snapshot
}

func (s *Stat) hit(response garnish.Response, t time.Duration) {
	hits := atomic.AddInt64(&s.hits, 1)
	status := response.GetStatus()
	if status > 499 {
		atomic.AddInt64(&s.failures, 1)
	} else if status > 399 {
		atomic.AddInt64(&s.errors, 1)
	} else {
		atomic.AddInt64(&s.oks, 1)
	}
	s.sample(hits, t)
}

func (s *Stat) sample(hits int64, t time.Duration) {
	index := -1
	if hits < s.sampleSize {
		index = int(hits)
	} else if s.sampleSizeF/float64(hits) > rand.Float64() {
		index = int(rand.Int31n(int32(s.sampleSize)))
	}
	if index != -1 {
		s.Lock()
		s.samples[index] = int(t)
		s.Unlock()
	}
}

// To avoid needing a lock on each hit, it's posisble for
// s.oks + s.errors + s.failure to be greater than s.hits (a snapshot could
// easily happen when some of these have been incremented but others haven't).
// (s.hits is still used for our reservoir sampling)
func (s *Stat) Snapshot() Snapshot {
	s.snapshot["2xx"] = atomic.SwapInt64(&s.oks, 0)
	s.snapshot["4xx"] = atomic.SwapInt64(&s.errors, 0)
	s.snapshot["5xx"] = atomic.SwapInt64(&s.failures, 0)
	s.snapshot["hits"] = s.snapshot["2xx"] + s.snapshot["4xx"] + s.snapshot["5xx"]

	hits := int(atomic.LoadInt64(&s.hits))
	sample := s.samples[:hits]
	sort.Ints(sample)
	for key, value := range s.percentiles {
		s.snapshot[key] = percentile(sample, value, hits)
	}
	atomic.StoreInt64(&s.hits, 0)
	return s.snapshot
}

func percentile(values []int, p float64, size int) int64 {
	if size == 0 {
		return -1
	}
	findex := p * float64(size+1)
	index := int(findex)
	if index < 1 {
		return int64(values[0])
	}
	if index >= size {
		return int64(values[size-1])
	}
	s1 := float64(size) - 1
	k := int(math.Floor(p*s1+1) - 1)
	valueK := float64(values[k])
	_, f := math.Modf(p*s1 + 1)
	return int64(math.Ceil(valueK + (f * (float64(values[k+1]) - valueK))))
}
