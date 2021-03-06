package gc

import (
	"gopkg.in/karlseguin/garnish.v1"
	"gopkg.in/karlseguin/garnish.v1/cache"
	"time"
)

type Cache struct {
	maxSize      int
	grace        time.Duration
	saint        bool
	lookup       garnish.CacheKeyLookup
	purgeHandler garnish.PurgeHandler
}

func NewCache() *Cache {
	return &Cache{
		maxSize: 104857600,
		grace:   time.Minute,
		lookup:  garnish.DefaultCacheKeyLookup,
		saint:   true,
	}
}

// The maximum size, in bytes, to cache
// [104857600] (100MB)
func (c *Cache) MaxSize(size int) *Cache {
	c.maxSize = size
	return c
}

// If a request is expired but within the grace window, the expired version
// will be returned. In a background job, the cache will be refreshed.
// Grace is effective at eliminating the thundering heard problem
// [1 minute]
func (c *Cache) Grace(window time.Duration) *Cache {
	c.grace = window
	return c
}

// Disable saint mode
// With saint mode, if the upstream returns a 5xx error and a cached
// response is available, the cached response will be returned
// regardless of how far expired it is.
// [saint is enabled by default]
func (c *Cache) NoSaint() *Cache {
	c.saint = false
	return c
}

// The function used to generate the primary and secondary cache keys
// This defaults use the URL for the primary key and the QueryString
// for the secondary key
// Having a separate primary and secondary cache key allows us to purge
// a group of values. For example:
// primary: /v1/users/32
// secondary: "ext=json"   and   "ext=xml"
// We can purge all variations associated with /v1/users/32
func (c *Cache) KeyLookup(lookup garnish.CacheKeyLookup) *Cache {
	c.lookup = lookup
	return c
}

// The function which will handle purge requests
// No default is provided since some level of custom authorization should be done
// If the handler returns a response, the middleware chain is stopped and the
// specified response is returned.
// If the handler does not return a response, the chain continues.
// This makes it possible to purge the garnish cache while allowing the purge
// request to be sent to the upstream
func (c *Cache) PurgeHandler(handler garnish.PurgeHandler) *Cache {
	c.purgeHandler = handler
	return c
}

func (c *Cache) Build(runtime *garnish.Runtime) error {
	runtime.Cache = garnish.NewCache()
	runtime.Cache.Saint = c.saint
	runtime.Cache.GraceTTL = c.grace
	runtime.Cache.Storage = cache.New(c.maxSize)

	if c.purgeHandler != nil {
		runtime.Cache.PurgeHandler = c.purgeHandler
		runtime.Router.AddNamed("_gc_purge_all", "PURGE", "/*", nil)
		runtime.Routes["_gc_purge_all"] = &garnish.Route{
			Stats: garnish.NewRouteStats(time.Millisecond * 500),
			Cache: garnish.NewRouteCache(-1, c.lookup),
		}
	}

	for _, route := range runtime.Routes {
		if route.Cache != nil && route.Cache.KeyLookup == nil {
			route.Cache.KeyLookup = c.lookup
		}
	}
	return nil
}
