package gc

import (
	"github.com/karlseguin/bytepool"
	"github.com/karlseguin/router"
	"github.com/karlseguin/dnscache"
)

// All the data needed to serve requests
// Built automatically when the garnish.Start() is called
type Runtime struct {
	Executor    Middleware
	Upstreams   map[string]*Upstream
	Routes      map[string]*Route
	Router      *router.Router
	BytePool    *bytepool.Pool
	StatsWorker *StatsWorker
	Cache       *Cache
	Resolver    *dnscache.Resolver
}

func (r *Runtime) RegisterStats(name string, reporter Reporter) {
	if r.StatsWorker != nil {
		r.StatsWorker.register(name, reporter)
	}
}
