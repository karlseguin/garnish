package gc

import (
	"github.com/karlseguin/ccache"
	"time"
	"strings"
	"sync"
	"strconv"
)

type CacheKeyLookup func(req *Request) (string, string)

func DefaultCacheKeyLookup(req *Request) (string, string) {
	return req.URL.Path, req.URL.RawQuery
}

type Cache struct {
	graceLock sync.Mutex
	*ccache.LayeredCache
	Saint bool
	GraceTTL time.Duration
	downloads map[string]time.Time
}

func NewCache(count int) *Cache {
	return &Cache{
		LayeredCache: ccache.Layered(ccache.Configure().MaxItems(uint64(count))),
		downloads: make(map[string]time.Time),
	}
}


func (c *Cache) Set(primary string, secondary string, config *RouteCache, res Response) {
	ttl := c.ttl(config, res)
	if ttl == 0 {
		return
	}
	cacheable := res
	if dt, ok := res.(Detachable); ok {
		cacheable = dt.Detach()
	}
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
		if strings.Contains(value, "private"){
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

func (c *Cache) Grace(primary string, secondary string, req *Request, next Middleware) {
	key := primary + secondary
	if c.reserveDownload(key) == false {
		return
	}

	defer func() {
		c.graceLock.Lock()
		delete(c.downloads, key)
		c.graceLock.Unlock()
	}()

	if res := next(req); res.Status() >= 500 {
		Log.Error("grace error for %q: %s", req.URL.String(), string(res.Body()))
	} else {
		c.Set(primary, secondary, req.Route.Cache, res)
	}
}

func (c *Cache) reserveDownload(key string) bool {
	now := time.Now()
	c.graceLock.Lock()
	defer c.graceLock.Unlock()
	expires, exists := c.downloads[key]
	if exists && expires.After(now) {
		return false
	}
	c.downloads[key] = now.Add(time.Second * 30)
	return true
}
