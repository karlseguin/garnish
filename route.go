package garnish

import (
	"time"
)

// Map an http.Request to a Route
type Router func(context Context) (*Route, Response)

// Generate a cache key and the vary parameters
type CacheKeyGenerator func(context Context) (string, string)

// Route information
type Route struct {
	Name     string
	Upstream string
	Caching  *Caching
}

type Caching struct {
	TTL          time.Duration
	KeyGenerator CacheKeyGenerator
}

func NewRoute(name, upstream string) *Route {
	return &Route{
		Name:     name,
		Upstream: upstream,
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
