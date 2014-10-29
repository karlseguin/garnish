package gc

import (
	"github.com/karlseguin/bytepool"
	"github.com/karlseguin/router"
)

type Runtime struct {
	Executor      Middleware
	Upstreams     map[string]*Upstream
	Routes        map[string]*Route
	Router        *router.Router
	BytePool      *bytepool.Pool
	StatsFileName string
}

// initialize any jobs we need to run
func (r *Runtime) Start() {
	if len(r.StatsFileName) > 0 {
		go NewStatsWorker(r).Run()
	}
}
