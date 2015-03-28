package gc

import (
	"fmt"
	"gopkg.in/karlseguin/garnish.v1"
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
		garnish.Log.Warnf("Route %q already defined. Overwriting.", name)
	}
	route := &Route{name: name, slow: -1}
	r.routes[name] = route
	return route
}

func (r *Router) Build(runtime *garnish.Runtime) error {
	routes := make(map[string]*garnish.Route, len(r.routes))
	for name, route := range r.routes {
		if len(route.cacheKeyLookupRef) > 0 {
			ref, ok := r.routes[route.cacheKeyLookupRef]
			if ok == false {
				return fmt.Errorf("route %s's keylookup referenced %s, a non-existent route", name, route.cacheKeyLookupRef)
			}
			if ref.cacheKeyLookup == nil {
				return fmt.Errorf("route %s's keylookup referenced %s, which does not have a cacheKeyLookup", name, route.cacheKeyLookupRef)
			}
			route.cacheKeyLookup = ref.cacheKeyLookup
		}
		r, err := route.Build(runtime)
		if err != nil {
			return err
		}
		routes[name] = r
	}
	runtime.Routes = routes
	return nil
}

type Route struct {
	name              string
	path              string
	method            string
	upstream          string
	handler           garnish.Handler
	slow              time.Duration
	cacheTTL          time.Duration
	cacheKeyLookup    garnish.CacheKeyLookup
	cacheKeyLookupRef string
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

// Register a route for the requested method + path
func (r *Route) Method(method string, path string) *Route {
	r.method, r.path = method, path
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
		garnish.Log.Warnf("CacheTTL can only be specified for a GET", r.name)
	}
	return r
}

// The function used to get the cache key for this route.
// (overwrites the global Cache's lookup)
func (r *Route) CacheKeyLookup(lookup garnish.CacheKeyLookup) *Route {
	if r.method == "GET" || r.method == "ALL" {
		r.cacheKeyLookup = lookup
	} else {
		garnish.Log.Warnf("CacheKeyLookup can only be specified for a GET", r.name)
	}
	return r
}

// Allows the route to reference another route's CacheKeyLookup (based on that
// route's name). Used by the file-based configuration; when configured
// programmatically, CacheKeyLookup should be used.
func (r *Route) CacheKeyLookupRef(route string) *Route {
	r.cacheKeyLookupRef = route
	return r
}

// Specify the handler function
func (r *Route) Handler(handler garnish.Handler) *Route {
	r.handler = handler
	return r
}

func (r *Route) Build(runtime *garnish.Runtime) (*garnish.Route, error) {
	route := &garnish.Route{
		Name:    r.name,
		Handler: r.handler,
	}

	if r.slow > -1 {
		route.Stats = garnish.NewRouteStats(r.slow)
	}

	if len(r.method) == 0 {
		return nil, fmt.Errorf("Route %q doesn't have a method+path", r.name)
	}

	if r.method == "GET" || r.method == "ALL" {
		route.Cache = garnish.NewRouteCache(r.cacheTTL, r.cacheKeyLookup)
	}

	if len(r.upstream) > 0 {
		upstream, exists := runtime.Upstreams[r.upstream]
		if exists == false {
			return nil, fmt.Errorf("Route %q has an unknown upstream %q", r.name, r.upstream)
		}
		route.Upstream = upstream
	}
	runtime.Router.AddNamed(r.name, r.method, r.path, nil)
	return route, nil
}
