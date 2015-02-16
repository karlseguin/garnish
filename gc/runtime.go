package gc

import (
	"gopkg.in/karlseguin/bytepool.v3"
	"gopkg.in/karlseguin/dnscache.v1"
	"gopkg.in/karlseguin/router.v1"
)

var UnauthorizedResponse = Empty(401)

// Authorization / authentication handler
type AuthHandler func(req *Request) Response

// All the data needed to serve requests
// Built automatically when the garnish.Start() is called
type Runtime struct {
	Executor      Middleware
	Upstreams     map[string]*Upstream
	Routes        map[string]*Route
	Router        *router.Router
	BytePool      *bytepool.Pool
	StatsWorker   *StatsWorker
	Cache         *Cache
	Resolver      *dnscache.Resolver
	HydrateLoader HydrateLoader
}

func (r *Runtime) RegisterStats(name string, reporter Reporter) {
	if r.StatsWorker != nil {
		r.StatsWorker.register(name, reporter)
	}
}
