package gc

import (
	"github.com/karlseguin/bytepool"
	"github.com/karlseguin/router"
)

type Runtime struct {
	Executor    Middleware
	Upstreams   map[string]*Upstream
	Routes      map[string]*Route
	Router      *router.Router
	BytePool    *bytepool.Pool
	StatsWorker *StatsWorker
}

func (r *Runtime) RegisterStats(name string, reporter Reporter) {
	if r.StatsWorker != nil {
		r.StatsWorker.register(name, reporter)
	}
}
