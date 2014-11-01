package gc

import (
	"time"
)

type Route struct {
	Name     string
	Method   string
	Upstream *Upstream
	Stats    *RouteStats
	Cache    *RouteCache
}

type RouteCache struct {
	KeyLookup CacheKeyLookup
	TTL       time.Duration
}

func NewRouteCache(ttl time.Duration, keyLookup CacheKeyLookup) *RouteCache {
	return &RouteCache{
		TTL:       ttl,
		KeyLookup: keyLookup,
	}
}
