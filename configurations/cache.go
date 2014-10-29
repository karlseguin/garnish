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
		saint: true,
	}
}

func (c *Cache) Count(count int) *Cache {
	c.count = count
	return c
}

func (c *Cache) Grace(window time.Duration) *Cache {
	c.grace = window
	return c
}

func (c *Cache) NoSaint() *Cache {
	c.saint = false
	return c
}

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
