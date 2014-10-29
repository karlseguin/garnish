package configurations

import (
	"github.com/karlseguin/garnish/gc"
	"time"
)

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

func (s *Stats) FileName(fileName string) *Stats {
	s.fileName = fileName
	return s
}

func (s *Stats) Slow(slow time.Duration) *Stats {
	s.slow = slow
	return s
}

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
