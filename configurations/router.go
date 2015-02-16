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
		gc.Log.Warnf("Route %q already defined. Overwriting.", name)
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
	name           string
	path           string
	method         string
	upstream       string
	slow           time.Duration
	cacheTTL       time.Duration
	cacheKeyLookup gc.CacheKeyLookup
}

// Specify the name of the upstream.
func (r *Route) Upstream(upstream string) *Route {
	r.upstream = upstream
	return r
}

// Register a route for requests issued with a GET
func (r *Route) Get(path string) *Route {
	r.method, r.path = "GET", path
	return r
}

// Register a route for requests issued with a POST
func (r *Route) Post(path string) *Route {
	r.method, r.path = "POST", path
	return r
}

// Register a route for requests issued with a PUT
func (r *Route) Put(path string) *Route {
	r.method, r.path = "PUT", path
	return r
}

// Register a route for requests issued with a DELETE
func (r *Route) Delete(path string) *Route {
	r.method, r.path = "DELETE", path
	return r
}

// Register a route for requests issued with a PURGE
func (r *Route) Purge(path string) *Route {
	r.method, r.path = "PURGE", path
	return r
}

// Register a route for requests issued with a PATCH
func (r *Route) Patch(path string) *Route {
	r.method, r.path = "PATCH", path
	return r
}

// Register a route for requests issued with a HEAD
func (r *Route) Head(path string) *Route {
	r.method, r.path = "HEAD", path
	return r
}

// Register a route for requests issued with an OPTIONS
func (r *Route) Options(path string) *Route {
	r.method, r.path = "OPTIONS", path
	return r
}

// Register a route for all methods. Can be overwritten on a per-method basis
// by registering the method-specific route BEFORE specifying the All variant.
func (r *Route) All(path string) *Route {
	r.method, r.path = "ALL", path
	return r
}

// The amount of time before this request is logged as a slow request
// (overwrites the global Stat's slow value)
func (r *Route) Slow(max time.Duration) *Route {
	r.slow = max
	return r
}

// The amount of time to cachet his request. If not specified, the
// Cache-Control header will be used (including not caching private).
// A value < 0 disables the cache for this route
func (r *Route) CacheTTL(ttl time.Duration) *Route {
	if r.method == "GET" || r.method == "ALL" {
		r.cacheTTL = ttl
	} else {
		gc.Log.Warnf("CacheTTL can only be specified for a GET", r.name)
	}
	return r
}

// The function used to get the cache key for this route.
// (overwrites the global Cache's lookup)
func (r *Route) CacheKeyLookup(lookup gc.CacheKeyLookup) *Route {
	if r.method == "GET" || r.method == "ALL" {
		r.cacheKeyLookup = lookup
	} else {
		gc.Log.Warnf("CacheKeyLookup can only be specified for a GET", r.name)
	}
	return r
}

func (r *Route) Build(runtime *gc.Runtime) *gc.Route {
	ok := true

	route := &gc.Route{
		Name: r.name,
	}

	if r.slow > -1 {
		route.Stats = gc.NewRouteStats(r.slow)
	}

	if len(r.method) == 0 {
		gc.Log.Errorf("Route %q doesn't have a method+path", r.name)
		ok = false
	}

	if r.method == "GET" || r.method == "ALL" {
		route.Cache = gc.NewRouteCache(r.cacheTTL, r.cacheKeyLookup)
	}

	if upstream, exists := runtime.Upstreams[r.upstream]; exists == false {
		gc.Log.Errorf("Route %q has an unknown upstream %q", r.name, r.upstream)
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
