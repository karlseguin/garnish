package stats

import (
	"github.com/karlseguin/garnish/gc"
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
	treshhold   time.Duration
	*Configuration
	hits        int64
	cache       int64
	oks         int64
	errors      int64
	failures    int64
	slow        int64
	samples     []int
	sampleCount int64
	scratch     []int
	snapshot    Snapshot
	percentiles map[string]float64
	max         int64
}

func newStat(c *Configuration) *Stat {
	return &Stat{
		snapshot:    make(Snapshot),
		samples:     make([]int, c.sampleSize),
		scratch:     make([]int, c.sampleSize),
		sampleSize:  c.sampleSize,
		sampleSizeF: c.sampleSizeF,
		treshhold:   c.treshhold,
		percentiles: c.percentiles,
	}
}

func (s *Stat) hit(response gc.Response, t time.Duration) {
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

	if isCacheHit(response) {
		atomic.AddInt64(&s.cache, 1)
	} else {
		if t > s.treshhold {
			atomic.AddInt64(&s.slow, 1)
		}
		s.sample(hits, t)
	}
}

func (s *Stat) sample(hits int64, t time.Duration) {
	index := -1
	sampleCount := atomic.LoadInt64(&s.sampleCount)
	if sampleCount < s.sampleSize {
		index = int(sampleCount)
		atomic.AddInt64(&s.sampleCount, 1)
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
	sampleCount := s.sampleCount
	s.snapshot["2xx"] = s.oks
	s.snapshot["4xx"] = s.errors
	s.snapshot["5xx"] = s.failures
	s.snapshot["slow"] = s.slow
	s.snapshot["hits"] = hits
	s.snapshot["cache"] = s.cache
	s.oks, s.errors, s.failures, s.hits, s.slow, s.cache, s.sampleCount = 0, 0, 0, 0, 0, 0, 0
	s.samples, s.scratch = s.scratch, s.samples
	s.snapLock.Unlock()

	samples := s.scratch[:sampleCount]
	sort.Ints(samples)
	for key, value := range s.percentiles {
		s.snapshot[key] = percentile(samples, value, int(sampleCount))
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

func isCacheHit(response gc.Response) bool {
	// _, ok := response.(*gc.CachedResponse)
	// return ok
	// TODO: fix
	return false
}
