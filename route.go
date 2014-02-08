package garnish

import (
	"time"
)

// Route parameters
type Params map[string]string

// Map an http.Request to a Route
type Router interface {
	Route(context Context) (*Route, Params, Response)
	RouteNames() []string
}

// Generate a cache key and the vary parameters
type CacheKeyGenerator func(context Context) (string, string)

// Route information
type Route struct {
	Name    string
	Caching *Caching
}

type Caching struct {
	TTL          time.Duration
	KeyGenerator CacheKeyGenerator
}

func NewRoute(name string) *Route {
	return &Route{
		Name: name,
	}
}

func (r *Route) Cache(keyGenerator CacheKeyGenerator) *Route {
	r.Caching = &Caching{KeyGenerator: keyGenerator}
	return r
}

func (r *Route) TTL(duration time.Duration) *Route {
	r.Caching.TTL = duration
	return r
}

func (r *Route) String() string {
	return r.Name
}
