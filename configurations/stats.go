package configurations

import (
	"github.com/karlseguin/garnish/gc"
	"time"
)

// Configuration for the stats middleware
type Stats struct {
	fileName string
	slow     time.Duration
}

func NewStats() *Stats {
	return &Stats{
		fileName: "stats.json",
		slow:     time.Millisecond * 250,
	}
}

// The file to save the statistics to.
// The file is overwritten on each write.
// ["stats.json"]
func (s *Stats) FileName(fileName string) *Stats {
	s.fileName = fileName
	return s
}

// The default threshold to consider flag a request as being slow
// This can be overwritten on a per-route basis
// [250ms]
func (s *Stats) Slow(slow time.Duration) *Stats {
	s.slow = slow
	return s
}

// In normal usage, there's no need to call this method.
// Builds the stats middleware into the runtime
func (s *Stats) Build(runtime *gc.Runtime) bool {
	for _, route := range runtime.Routes {
		if route.Stats.Treshold == -1 {
			route.Stats.Treshold = s.slow
		}
	}
	sw := gc.NewStatsWorker(runtime, s.fileName)
	runtime.StatsWorker = sw
	go sw.Run()
	return true
}
