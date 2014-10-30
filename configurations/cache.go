package configurations

import (
	"github.com/karlseguin/garnish/gc"
	"time"
)

type Cache struct {
	count  int
	grace  time.Duration
	saint  bool
	lookup gc.CacheKeyLookup
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
// This defaults used the URL for the primary key and the QueryString
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

func (c *Cache) Build(runtime *gc.Runtime) bool {
	runtime.Cache = gc.NewCache(c.count)
	runtime.Cache.Saint = c.saint
	runtime.Cache.GraceTTL = c.grace
	for _, route := range runtime.Routes {
		if route.Cache.KeyLookup == nil {
			route.Cache.KeyLookup = c.lookup
		}
	}
	return true
}
