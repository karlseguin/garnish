package configurations

import (
	"github.com/karlseguin/garnish/gc"
	"time"
)

type Router struct {
	routes map[string]*Route
}

func NewRouter() *Router {
	return &Router{
		routes: make(map[string]*Route),
	}
}

func (r *Router) Add(name string) *Route {
	if _, exists := r.routes[name]; exists {
		gc.Log.Warn("Route %q already defined. Overwriting.", name)
	}
	route := &Route{name: name, slow: -1}
	r.routes[name] = route
	return route
}

func (r *Router) Build(runtime *gc.Runtime) bool {
	ok := true

	routes := make(map[string]*gc.Route, len(r.routes))
	for name, route := range r.routes {
		if r := route.Build(runtime); r != nil {
			routes[name] = r
		} else {
			ok = false
		}
	}
	runtime.Routes = routes
	return ok
}

type Route struct {
	name     string
	path     string
	method   string
	upstream string
	slow     time.Duration
	cacheTTL time.Duration
	cacheKeyLookup gc.CacheKeyLookup
}

func (r *Route) Upstream(upstream string) *Route {
	r.upstream = upstream
	return r
}

func (r *Route) Get(path string) *Route {
	r.method, r.path = "GET", path
	return r
}

func (r *Route) Post(path string) *Route {
	r.method, r.path = "POST", path
	return r
}

func (r *Route) Put(path string) *Route {
	r.method, r.path = "PUT", path
	return r
}

func (r *Route) Delete(path string) *Route {
	r.method, r.path = "DELETE", path
	return r
}

func (r *Route) Purge(path string) *Route {
	r.method, r.path = "PURGE", path
	return r
}

func (r *Route) Patch(path string) *Route {
	r.method, r.path = "PATCH", path
	return r
}

func (r *Route) Head(path string) *Route {
	r.method, r.path = "HEAD", path
	return r
}

func (r *Route) Options(path string) *Route {
	r.method, r.path = "OPTIONS", path
	return r
}

func (r *Route) All(path string) *Route {
	r.method, r.path = "ALL", path
	return r
}

func (r *Route) Slow(max time.Duration) *Route {
	r.slow = max
	return r
}

func (r *Route) CacheTTL(ttl time.Duration) *Route {
	r.cacheTTL = ttl
	return r
}

func (r *Route) CacheKeyLookup(lookup gc.CacheKeyLookup) *Route {
	r.cacheKeyLookup = lookup
	return r
}

func (r *Route) Build(runtime *gc.Runtime) *gc.Route {
	ok := true

	route := &gc.Route{
		Name:  r.name,
		Stats: gc.NewRouteStats(r.slow),
		Cache: gc.NewRouteCache(r.cacheTTL, r.cacheKeyLookup),
	}

	if len(r.method) == 0 {
		gc.Log.Error("Route %q doesn't have a method+path", r.name)
		ok = false
	}

	if upstream, exists := runtime.Upstreams[r.upstream]; exists == false {
		gc.Log.Error("Route %q has an unknown upstream %q", r.name, r.upstream)
		ok = false
	} else {
		route.Upstream = upstream
	}

	if ok == false {
		return nil
	}

	runtime.Router.AddNamed(r.name, r.method, r.path, nil)
	return route
}
