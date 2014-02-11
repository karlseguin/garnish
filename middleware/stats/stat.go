package stats

import (
	"github.com/karlseguin/garnish/core"
	"math"
	"math/rand"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

type Stat struct {
	snapLock    sync.RWMutex
	sampleLock  sync.Mutex
	sampleSize  int64
	sampleSizeF float64
	*Configuration
	hits        int64
	oks         int64
	errors      int64
	failures    int64
	samples     []int
	scratch     []int
	snapshot    Snapshot
	percentiles map[string]float64
}

func newStat(c *Configuration) *Stat {
	return &Stat{
		snapshot:    make(Snapshot),
		samples:     make([]int, c.sampleSize),
		scratch:     make([]int, c.sampleSize),
		percentiles: c.percentiles,
	}
}

func (s *Stat) hit(response core.Response, t time.Duration) {
	s.snapLock.RLock()
	defer s.snapLock.RUnlock()
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
		s.sampleLock.Lock()
		s.samples[index] = int(t / 1000)
		s.sampleLock.Unlock()
	}
}

func (s *Stat) Snapshot() Snapshot {
	s.snapLock.Lock()
	hits := s.hits
	s.snapshot["2xx"] = s.oks
	s.snapshot["4xx"] = s.errors
	s.snapshot["5xx"] = s.failures
	s.snapshot["hits"] = hits
	s.oks, s.errors, s.failures, s.hits = 0, 0, 0, 0
	s.samples, s.scratch = s.scratch, s.samples
	s.snapLock.Unlock()

	samples := s.samples[:hits]
	sort.Ints(samples)
	for key, value := range s.percentiles {
		s.snapshot[key] = percentile(samples, value, int(hits))
	}
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
