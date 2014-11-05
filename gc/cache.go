package gc

import (
	"github.com/karlseguin/ccache"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	PurgeHitResponse  = Empty(200)
	PurgeMissResponse = Empty(204)
)

// A function that generates cache keys from a request
type CacheKeyLookup func(req *Request) (string, string)

// A function that purges the cache
// Returning a nil response means that the request will be forward onwards
type PurgeHandler func(req *Request, lookup CacheKeyLookup, cache Purgeable) Response

// An interface for deleting items
type Purgeable interface {
	Delete(primary, secondary string) bool
	DeleteAll(primary string) bool
}

func DefaultCacheKeyLookup(req *Request) (string, string) {
	return req.URL.Path, req.URL.RawQuery
}

type Cache struct {
	graceLock sync.RWMutex
	*ccache.LayeredCache
	downloads    map[string]time.Time
	Saint        bool
	GraceTTL     time.Duration
	PurgeHandler PurgeHandler
}

func NewCache(count int) *Cache {
	return &Cache{
		LayeredCache: ccache.Layered(ccache.Configure().MaxItems(uint64(count))),
		downloads:    make(map[string]time.Time),
	}
}

func (c *Cache) Set(primary string, secondary string, config *RouteCache, res Response) {
	ttl := c.ttl(config, res)
	if ttl == 0 {
		return
	}

	cacheable := res.ToCacheable()
	cacheable.Header().Set("X-Cache", "hit")
	c.LayeredCache.Set(primary, secondary, cacheable, ttl)
}

func (c *Cache) ttl(config *RouteCache, res Response) time.Duration {
	status := res.Status()
	if status >= 200 && status <= 400 && config.TTL > 0 {
		return config.TTL
	}

	cc := res.Header()["Cache-Control"]
	if len(cc) == 0 {
		return 0
	}

	for _, value := range cc {
		if strings.Contains(value, "private") {
			break
		}
		if index := strings.Index(value, "max-age="); index > -1 {
			if seconds, err := strconv.Atoi(value[index+8:]); err == nil {
				return time.Second * time.Duration(seconds)
			} else {
				Log.Warn("invalid cache control header %q", value)
				break
			}
		}
	}
	return 0
}

// A clone is critical since the original request is likely to be closed
// before we're finishing with Grace and we might end up with a request
// that contains data from multiple sources.
func (c *Cache) Grace(primary string, secondary string, req *Request, next Middleware) {
	key := primary + secondary
	if c.reserveDownload(key) == false {
		return
	}
	go c.grace(key, primary, secondary, req.Clone(), next)
}

func (c *Cache) grace(key string, primary string, secondary string, req *Request, next Middleware) {
	defer func() {
		c.graceLock.Lock()
		delete(c.downloads, key)
		c.graceLock.Unlock()
	}()

	res := next(req)
	if res == nil {
		Log.Error("grace nil response for %q", req.URL.String())
		return
	}
	defer res.Close()
	if res.Status() >= 500 {
		Log.Error("grace error for %q", req.URL.String())
	} else {
		c.Set(primary, secondary, req.Route.Cache, res)
	}
}

func (c *Cache) reserveDownload(key string) bool {
	now := time.Now()
	c.graceLock.RLock()
	expires, exists := c.downloads[key]
	c.graceLock.RUnlock()
	if exists && expires.After(now) {
		return false
	}

	c.graceLock.Lock()
	defer c.graceLock.Unlock()
	expires, exists = c.downloads[key]
	if exists && expires.After(now) {
		return false
	}

	c.downloads[key] = now.Add(time.Second * 30)
	return true
}
