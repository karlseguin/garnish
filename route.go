package garnish

import (
	"time"
)

type Route struct {
	Name        string
	Method      string
	Upstream    Upstream
	Stats       *RouteStats
	Cache       *RouteCache
	StopHandler Handler
	FlowHandler Middleware
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
