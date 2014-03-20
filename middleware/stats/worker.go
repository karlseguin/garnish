package stats

import (
	"github.com/karlseguin/garnish/gc"
	"sync"
	"time"
)

type Snapshot map[string]int64

type Worker struct {
	sync.Mutex
	running    bool
	window     time.Duration
	logger     gc.Logger
	routeStats map[string]*Stat
	persister  Persister
	reporters  map[string]Reporter
}

func (w *Worker) start() {
	w.Lock()
	w.running = true
	w.Unlock()

	for {
		time.Sleep(w.window)
		if w.snapshot() == false {
			return
		}
	}
}

func (w *Worker) snapshot() bool {
	w.Lock()
	defer w.Unlock()
	if w.running == false {
		return false
	}

	routes := make(map[string]Snapshot)
	for key, stat := range w.routeStats {
		ss := stat.Snapshot()
		if ss["hits"] > 0 {
			routes[key] = ss
		}
	}

	other := make(map[string]Snapshot)
	for name, reporter := range w.reporters {
		other[name] = Snapshot(reporter())
	}

	if len(routes) > 0 || len(other) > 0 || w.persister.LogEmpty() {
		if err := w.persister.Persist(routes, other); err != nil {
			w.logger.Errorf("Failed to save stats: %v", err)
		}
	}
	return true
}

func (w *Worker) stop() {
	w.Lock()
	w.running = false
	w.Unlock()
}
