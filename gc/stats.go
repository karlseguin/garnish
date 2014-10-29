package gc

import (
	"encoding/json"
	"math"
	"math/rand"
	"os"
	"runtime"
	"runtime/debug"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

var STATS_SAMPLE_SIZE int64 = 1000
var STATS_SAMPLE_SIZE_F = float64(STATS_SAMPLE_SIZE)
var STATS_PERCENTILES = map[string]float64{"75p": 0.75, "95p": 0.95}

type Snapshot map[string]int64
type Reporter func() map[string]int64

type RouteStats struct {
	Treshold time.Duration
	snapshot Snapshot

	sampleLock  sync.Mutex
	sampleCount int64
	samplesA    []int
	samplesB    []int

	hits     int64
	oks      int64
	errors   int64
	failures int64
	slow     int64
}

func NewRouteStats(treshold time.Duration) *RouteStats {
	return &RouteStats{
		Treshold: treshold,
		snapshot: make(Snapshot, 5+len(STATS_PERCENTILES)),
		samplesA: make([]int, STATS_SAMPLE_SIZE),
		samplesB: make([]int, STATS_SAMPLE_SIZE),
	}
}

func (s *RouteStats) Hit(response Response, t time.Duration) {
	hits := atomic.AddInt64(&s.hits, 1)
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
	s.sample(hits, t)
}

func (s *RouteStats) sample(hits int64, t time.Duration) {
	index := -1
	sampleCount := atomic.LoadInt64(&s.sampleCount)
	if sampleCount < STATS_SAMPLE_SIZE {
		index = int(sampleCount)
		atomic.AddInt64(&s.sampleCount, 1)
	} else if STATS_SAMPLE_SIZE_F/float64(hits) > rand.Float64() {
		index = int(rand.Int63n(STATS_SAMPLE_SIZE))
	}
	if index != -1 {
		s.sampleLock.Lock()
		s.samplesA[index] = int(t / 1000)
		s.sampleLock.Unlock()
	}
}

func (s *RouteStats) Snapshot() Snapshot {
	hits := atomic.SwapInt64(&s.hits, 0)
	s.snapshot["2xx"] = atomic.SwapInt64(&s.oks, 0)
	s.snapshot["4xx"] = atomic.SwapInt64(&s.errors, 0)
	s.snapshot["5xx"] = atomic.SwapInt64(&s.failures, 0)
	s.snapshot["slow"] = atomic.SwapInt64(&s.slow, 0)
	s.snapshot["hits"] = hits

	s.sampleLock.Lock()
	sampleCount := int(s.sampleCount)
	s.sampleCount = 0
	s.samplesA, s.samplesB = s.samplesB, s.samplesA
	s.sampleLock.Unlock()

	if sampleCount > 0 {
		samples := s.samplesB[:sampleCount]
		sort.Ints(samples)
		for key, value := range STATS_PERCENTILES {
			s.snapshot[key] = percentile(samples, value, sampleCount)
		}
	} else {
		for key, _ := range STATS_PERCENTILES {
			s.snapshot[key] = 0
		}
	}
	return s.snapshot
}

func percentile(values []int, p float64, size int) int64 {
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

type StatsWorker struct {
	fileName  string
	routes    map[string]*Route
	gcstats   *debug.GCStats
	rt        map[string]int64
	stats     map[string]interface{}
	reporters map[string]Reporter
}

func NewStatsWorker(runtime *Runtime, fileName string) *StatsWorker {
	rt := map[string]int64{"gc": 0, "go": 0}
	return &StatsWorker{
		rt:       rt,
		routes:   runtime.Routes,
		gcstats:  new(debug.GCStats),
		fileName: fileName,
		stats: map[string]interface{}{
			"time":    time.Now(),
			"routes":  nil,
			"runtime": rt,
			"other":   nil,
		},
		reporters: make(map[string]Reporter),
	}
}

func (w *StatsWorker) Run() {
	for {
		time.Sleep(time.Minute)
		w.work()
	}
}

func (w *StatsWorker) register(name string, reporter Reporter) {
	if _, exists := w.reporters[name]; exists {
		Log.Warn("reporter with name %q was already registered.", name)
		return
	}
	w.reporters[name] = reporter
}

func (w *StatsWorker) work() {
	w.stats["time"] = time.Now()
	w.stats["routes"] = w.collectRouteStats()
	w.stats["other"] = w.collectReporters()
	debug.ReadGCStats(w.gcstats)
	w.rt["gc"] = w.gcstats.NumGC
	w.rt["go"] = int64(runtime.NumGoroutine())
	w.save()
}

func (w *StatsWorker) collectRouteStats() map[string]Snapshot {
	routes := make(map[string]Snapshot)
	for name, route := range w.routes {
		snapshot := route.Stats.Snapshot()
		if snapshot["hits"] > 0 {
			routes[name] = snapshot
		}
	}
	return routes
}

func (w *StatsWorker) collectReporters() map[string]Snapshot {
	reporters := make(map[string]Snapshot)
	for name, reporter := range w.reporters {
		reporters[name] = reporter()
	}
	return reporters
}

func (w *StatsWorker) save() {
	bytes, err := json.Marshal(w.stats)
	if err != nil {
		Log.Error("stats serialize %v", err)
		return
	}

	file, err := os.OpenFile(w.fileName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		Log.Error("stats save %v", err)
		return
	}
	defer file.Close()

	if _, err := file.Write(bytes); err != nil {
		Log.Error("stats write %v", err)
	}
}
