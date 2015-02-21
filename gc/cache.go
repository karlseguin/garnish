package gc

import (
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	PurgeHitResponse    = Empty(200)
	PurgeMissResponse   = Empty(204)
	NotModifiedResponse = Empty(304)

	hitHeaderValue = []string{"hit"}
	zero           time.Time
)

type CacheStorage interface {
	Get(primary, secondary string) CachedResponse
	Set(primary string, secondary string, response CachedResponse)
	Delete(primary, secondary string) bool
	DeleteAll(primary string) bool
}

type CachedResponse interface {
	Response
	Size() int
	Expire(at time.Time)
	Expires() time.Time
}

// A function that generates cache keys from a request
type CacheKeyLookup func(req *Request) (string, string)

// A function that purges the cache
// Returning a nil response means that the request will be forward onwards
type PurgeHandler func(req *Request, lookup CacheKeyLookup, cache CacheStorage) Response

func DefaultCacheKeyLookup(req *Request) (string, string) {
	return req.URL.Path, req.URL.RawQuery
}

type Cache struct {
	sync.Mutex
	downloads    map[string]time.Time
	Storage      CacheStorage
	Saint        bool
	GraceTTL     time.Duration
	PurgeHandler PurgeHandler
}

func NewCache() *Cache {
	return &Cache{
		downloads: make(map[string]time.Time),
	}
}

func (c *Cache) Set(primary string, secondary string, config *RouteCache, res Response) {
	ttl := c.ttl(config, res)
	if ttl == 0 {
		return
	}

	cacheable := res.ToCacheable(time.Now().Add(ttl))
	c.Storage.Set(primary, secondary, cacheable)
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
				Log.Warnf("invalid cache control header %q", value)
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
		req.Close()
		c.Lock()
		delete(c.downloads, key)
		c.Unlock()
	}()

	res := next(req)
	if res == nil {
		Log.Errorf("grace nil response for %q", req.URL)
		return
	}
	defer res.Close()
	if res.Status() >= 500 {
		Log.Errorf("grace error for %q", req.URL)
	} else {
		c.Set(primary, secondary, req.Route.Cache, res)
	}
}

func (c *Cache) reserveDownload(key string) bool {
	now := time.Now()
	c.Lock()
	defer c.Unlock()
	if expires, exists := c.downloads[key]; exists && expires.After(now) {
		return false
	}
	c.downloads[key] = now.Add(time.Second * 30)
	return true
}
