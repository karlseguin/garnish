// Middleware package that implements caching.
package caching

import (
	"github.com/karlseguin/garnish/gc"
	"github.com/karlseguin/ccache"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Caching struct {
	routeConfigs map[string]*RouteConfig
	cache        *ccache.LayeredCache
	lock         sync.RWMutex
	runtimeSkip  RuntimeSkip
	downloading  map[string]time.Time
}

func (c *Caching) Name() string {
	return "caching"
}

var (
	grace = func(c *Caching, key, vary string, context gc.Context, config *RouteConfig, next gc.Next) {
		go c.grace(key, vary, context, config, next)
	}
	purgeHitResponse  = gc.Respond([]byte("")).Status(200)
	purgeMissResponse = gc.Respond([]byte("")).Status(204)
)

func (c *Caching) Run(context gc.Context, next gc.Next) gc.Response {
	config := c.routeConfigs[context.Route().Name]
	if config.ttl == 0 {
		context.Info("not cacheable")
		return next(context)
	}

	if c.runtimeSkip != nil && c.runtimeSkip(context) {
		context.Info("runtime skip")
		return next(context)
	}

	request := context.Request()
	switch request.Method {
	case "GET":
		return c.get(context, config, request, next)
	case "PURGE":
		return c.purge(context, config, request)
	default:
		context.Info("not cacheable")
		return next(context)
	}
}

func (c *Caching) get(context gc.Context, config *RouteConfig, request *http.Request, next gc.Next) gc.Response {
	key, vary := config.keyGenerator(context)
	cached := c.cache.Get(key, vary)
	if cached != nil {
		now, expires := time.Now(), cached.Expires()
		if expires.After(now) {
			context.Info("hit")
			return cached.Value().(gc.Response)
		}
		if expires.Add(config.grace).After(now) {
			context.Info("grace")
			grace(c, key, vary, context, config, next)
			return cached.Value().(gc.Response)
		}
	}

	context.Info("miss")
	response := next(context)
	if response.GetStatus() >= 500 && config.saint.Nanoseconds() > 0 {
		if cached == nil {
			return response
		}
		context.Errorf("%q %d %v", context.Request().URL, response.GetStatus(), string(response.GetBody()))
		context.Info("saint")
		cached.Extend(config.saint)
		return cached.Value().(gc.Response)
	}
	c.set(key, vary, context, config, response)
	response.GetHeader().Set("X-Cache", "miss")
	return response
}

func (c *Caching) set(key, vary string, context gc.Context, config *RouteConfig, response gc.Response) {
	ttl, ok := ttl(config, response)
	if ok == false {
		context.Error("configured to cache but no expiry was given")
		return
	}
	detached := response.Detach()
	detached.GetHeader().Set("X-Cache", "hit")
	c.cache.Set(key, vary, detached, ttl)
}

func ttl(config *RouteConfig, response gc.Response) (time.Duration, bool) {
	status := response.GetStatus()
	if status >= 200 && status <= 400 && config.ttl.Nanoseconds() > 0 {
		return config.ttl, true
	}

	for _, value := range response.GetHeader()["Cache-Control"] {
		if index := strings.Index(value, "max-age="); index > -1 {
			seconds, err := strconv.Atoi(value[index+8:])
			if err != nil {
				return time.Minute, false
			}
			return time.Second * time.Duration(seconds), true
		}
	}

	if status == 404 {
		return time.Second * 10, true
	}
	return time.Second, false
}

func (c *Caching) grace(key, vary string, context gc.Context, config *RouteConfig, next gc.Next) {
	downloadKey := key + vary
	if c.isDownloading(downloadKey) {
		return
	}
	if c.lockDownload(downloadKey) == false {
		return
	}
	defer c.unlockDownload(downloadKey)
	context.Info("grace start download")
	response := next(context)
	if response.GetStatus() < 500 {
		context.Infof("grace %d", response.GetStatus())
		c.set(key, vary, context, config, response)
	} else {
		context.Errorf("%q %d %v", context.Request().URL, response.GetStatus(), string(response.GetBody()))
	}
}

func (c *Caching) isDownloading(key string) bool {
	c.lock.RLock()
	expires, exists := c.downloading[key]
	c.lock.RUnlock()
	return exists && time.Now().Before(expires)
}

func (c *Caching) lockDownload(key string) bool {
	c.lock.Lock()
	defer c.lock.Unlock()
	if _, exists := c.downloading[key]; exists {
		return false
	}
	c.downloading[key] = time.Now().Add(time.Second * 30)
	return true
}

func (c *Caching) unlockDownload(key string) {
	c.lock.Lock()
	delete(c.downloading, key)
	c.lock.Unlock()
}

func (c *Caching) purge(context gc.Context, config *RouteConfig, request *http.Request) gc.Response {
	if config.authorizePurge == nil || config.authorizePurge(context) == false {
		context.Info("unauthorized purge")
		return gc.Unauthorized
	}
	key, _ := config.keyGenerator(context)
	if c.cache.DeleteAll(key) {
		context.Info("purge hit")
		return purgeHitResponse
	}
	context.Info("purge miss")
	return purgeMissResponse
}
