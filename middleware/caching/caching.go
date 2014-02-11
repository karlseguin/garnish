// Middleware package that implements caching.
package caching

import (
	"github.com/karlseguin/garnish/caches"
	"github.com/karlseguin/garnish/core"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Caching struct {
	routeConfigs map[string]*RouteConfig
	logger       core.Logger
	cache        caches.Cache
	lock         sync.RWMutex
	downloading  map[string]time.Time
}

func (c *Caching) Name() string {
	return "caching"
}

var (
	grace = func(c *Caching, key, vary string, context core.Context, config *RouteConfig, next core.Next) {
		go c.grace(key, vary, context, config, next)
	}
	purgeHitResponse  = core.Respond([]byte("")).Status(200)
	purgeMissResponse = core.Respond([]byte("")).Status(204)
)

func (c *Caching) Run(context core.Context, next core.Next) core.Response {
	config := c.routeConfigs[context.Route().Name]
	if config == nil {
		c.logger.Info(context, "not cacheable")
		return next(context)
	}

	request := context.Request()
	switch request.Method {
	case "GET":
		return c.get(context, config, request, next)
	case "PURGE":
		return c.purge(context, config, request)
	default:
		c.logger.Info(context, "not cacheable")
		return next(context)
	}
}

func (c *Caching) get(context core.Context, config *RouteConfig, request *http.Request, next core.Next) core.Response {
	key, vary := config.keyGenerator(context)
	cached := c.cache.Get(key, vary)
	if cached != nil {
		now := time.Now()
		if cached.Expires.After(now) {
			c.logger.Info(context, "hit")
			return cached
		}
		if cached.Expires.Add(config.grace).After(now) {
			c.logger.Info(context, "grace")
			grace(c, key, vary, context, config, next)
			return cached
		}
	}

	c.logger.Info(context, "miss")
	response := next(context)
	if response.GetStatus() >= 500 && config.saint.Nanoseconds() > 0 {
		if cached == nil {
			return response
		}
		c.logger.Errorf(context, "%q %d %v", context.Request().URL, response.GetStatus(), string(response.GetBody()))
		c.logger.Info(context, "saint")
		cached.Expires = time.Now().Add(config.saint)
		return cached
	}
	return c.set(key, vary, context, config, response)
}

func (c *Caching) set(key, vary string, context core.Context, config *RouteConfig, response core.Response) core.Response {
	ttl, ok := ttl(config, response)
	if ok == false {
		c.logger.Error(context, "configured to cache but no expiry was given")
		return response
	}
	cr := &caches.CachedResponse{
		Expires:  time.Now().Add(ttl),
		Response: response.Detach(),
	}
	c.cache.Set(key, vary, cr)
	return cr
}

func ttl(config *RouteConfig, response core.Response) (time.Duration, bool) {
	status := response.GetStatus()
	if status >= 200 && status <= 400 && config.ttl.Nanoseconds() > 0 {
		return config.ttl, true
	}

	for _, value := range response.GetHeader()["Cache-Control"] {
		if index := strings.Index(value, "max-age="); index > -1 {
			seconds, err := strconv.Atoi(value[index+8:])
			if err != nil {
				return time.Second, false
			}
			return time.Second * time.Duration(seconds), true
		}
	}
	return time.Second, false
}

func (c *Caching) grace(key, vary string, context core.Context, config *RouteConfig, next core.Next) {
	downloadKey := key + vary
	if c.isDownloading(downloadKey) {
		return
	}
	if c.lockDownload(downloadKey) == false {
		return
	}
	defer c.unlockDownload(downloadKey)
	c.logger.Info(context, "grace start download")
	response := next(context)
	if response.GetStatus() < 500 {
		c.logger.Infof(context, "grace %d", response.GetStatus())
		c.set(key, vary, context, config, response)
	} else {
		c.logger.Errorf(context, "%q %d %v", context.Request().URL, response.GetStatus(), string(response.GetBody()))
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

func (c *Caching) purge(context core.Context, config *RouteConfig, request *http.Request) core.Response {
	if config.authorizePurge == nil || config.authorizePurge(context) == false {
		c.logger.Info(context, "unauthorized purge")
		return core.Unauthorized
	}
	key, _ := config.keyGenerator(context)
	if c.cache.Delete(key) {
		c.logger.Info(context, "purge hit")
		return purgeHitResponse
	}
	c.logger.Info(context, "purge miss")
	return purgeMissResponse
}
