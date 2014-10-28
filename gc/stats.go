package gc

import (
	"encoding/json"
	"os"
	"sync/atomic"
	"time"
)

type Snapshot map[string]int64

type Stats struct {
	Treshold time.Duration
	snapshot Snapshot

	hits     int64
	oks      int64
	errors   int64
	failures int64
	slow     int64
}

func NewStats(treshold time.Duration) *Stats {
	return &Stats{
		Treshold: treshold,
		snapshot: make(Snapshot, 5),
	}
}

func (s *Stats) Hit(response Response, t time.Duration) {
	atomic.AddInt64(&s.hits, 1)
	status := response.Status()
	if status > 499 {
		atomic.AddInt64(&s.failures, 1)
	} else if status > 399 {
		atomic.AddInt64(&s.errors, 1)
	} else {
		atomic.AddInt64(&s.oks, 1)
	}
	if t > s.Treshold {
		atomic.AddInt64(&s.slow, 1)
	}
}

// func (s *Stat) sample(hits int64, t time.Duration) {
// 	index := -1
// 	sampleCount := atomic.LoadInt64(&s.sampleCount)
// 	if sampleCount < s.sampleSize {
// 		index = int(sampleCount)
// 		atomic.AddInt64(&s.sampleCount, 1)
// 	} else if s.sampleSizeF/float64(hits) > rand.Float64() {
// 		index = int(rand.Int31n(int32(s.sampleSize)))
// 	}
// 	if index != -1 {
// 		s.sampleLock.Lock()
// 		s.samples[index] = int(t / 1000)
// 		s.sampleLock.Unlock()
// 	}
// }

func (s *Stats) Snapshot() Snapshot {
	hits := atomic.SwapInt64(&s.hits, 0)
	s.snapshot["2xx"] = atomic.SwapInt64(&s.oks, 0)
	s.snapshot["4xx"] = atomic.SwapInt64(&s.errors, 0)
	s.snapshot["5xx"] = atomic.SwapInt64(&s.failures, 0)
	s.snapshot["slow"] = atomic.SwapInt64(&s.slow, 0)
	s.snapshot["hits"] = hits
	return s.snapshot
}

// func percentile(values []int, p float64, size int) int64 {
// 	if size == 0 {
// 		return -1
// 	}
// 	findex := p * float64(size+1)
// 	index := int(findex)
// 	if index < 1 {
// 		return int64(values[0])
// 	}
// 	if index >= size {
// 		return int64(values[size-1])
// 	}
// 	s1 := float64(size) - 1
// 	k := int(math.Floor(p*s1+1) - 1)
// 	valueK := float64(values[k])
// 	_, f := math.Modf(p*s1 + 1)
// 	return int64(math.Ceil(valueK + (f * (float64(values[k+1]) - valueK))))
// }

// func isCacheHit(response gc.Response) bool {
// 	_, ok := response.(*caches.CachedResponse)
// 	return ok
// }

func StatsWorker(r *Runtime) {
	for {
		time.Sleep(time.Minute)
		dumpStats(r)
	}
}

func dumpStats(r *Runtime) {
	Log.Info("stats dump")
	routes := make(map[string]Snapshot)
	for name, route := range r.Routes {
		snapshot := route.Stats.Snapshot()
		if snapshot["hits"] > 0 {
			routes[name] = snapshot
		}
	}

	if len(routes) == 0 {
		Log.Info("stats none")
		return
	}

	m := map[string]interface{}{
		"time":   time.Now(),
		"routes": routes,
	}

	bytes, err := json.Marshal(m)
	if err != nil {
		Log.Error("stats serialize %v", err)
		return
	}

	file, err := os.OpenFile(r.StatsFileName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		Log.Error("stats save %v", err)
		return
	}
	defer file.Close()

	if _, err := file.Write(bytes); err != nil {
		Log.Error("stats write %v", err)
	}
}
