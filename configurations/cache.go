package configurations

import (
	"github.com/karlseguin/garnish/cache"
	"github.com/karlseguin/garnish/gc"
	"time"
)

type Cache struct {
	count        int
	grace        time.Duration
	saint        bool
	lookup       gc.CacheKeyLookup
	purgeHandler gc.PurgeHandler
}

func NewCache() *Cache {
	return &Cache{
		count:  5000,
		grace:  time.Minute,
		lookup: gc.DefaultCacheKeyLookup,
		saint:  true,
	}
}

// The maximum number of items to cache
// [5000]
func (c *Cache) Count(count int) *Cache {
	c.count = count
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
func (c *Cache) KeyLookup(lookup gc.CacheKeyLookup) *Cache {
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
func (c *Cache) PurgeHandler(handler gc.PurgeHandler) *Cache {
	c.purgeHandler = handler
	return c
}

func (c *Cache) Build(runtime *gc.Runtime) error {
	runtime.Cache = gc.NewCache()
	runtime.Cache.Saint = c.saint
	runtime.Cache.GraceTTL = c.grace
	runtime.Cache.Storage = cache.New(c.count)

	if c.purgeHandler != nil {
		runtime.Cache.PurgeHandler = c.purgeHandler
		runtime.Router.AddNamed("_gc_purge_all", "PURGE", "/*", nil)
		runtime.Routes["_gc_purge_all"] = &gc.Route{
			Stats: gc.NewRouteStats(time.Millisecond * 500),
			Cache: gc.NewRouteCache(-1, c.lookup),
		}
	}

	for _, route := range runtime.Routes {
		if route.Cache != nil && route.Cache.KeyLookup == nil {
			route.Cache.KeyLookup = c.lookup
		}
	}
	return nil
}
