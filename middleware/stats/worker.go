package stats

import (
	"github.com/karlseguin/garnish/core"
	"sync"
	"time"
)

type Snapshot map[string]int64

type Worker struct {
	sync.Mutex
	running    bool
	window     time.Duration
	logger     core.Logger
	routeStats map[string]*Stat
	persister  Persister
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

	snapshots := make(map[string]Snapshot)
	for key, stat := range w.routeStats {
		ss := stat.Snapshot()
		if ss["hits"] > 0 {
			snapshots[key] = ss
		}
	}
	if len(snapshots) > 0 {
		if err := w.persister.Persist(snapshots); err != nil {
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
