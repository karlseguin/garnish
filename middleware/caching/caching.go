// Middleware package that implements caching.
package caching

import (
	"github.com/karlseguin/garnish/core"
	"github.com/karlseguin/garnish/caches"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Caching struct {
	*Configuration
	lock        sync.RWMutex
	downloading map[string]time.Time
}

func (c *Caching) Name() string {
	return "caching"
}

var (
	grace = func(c *Caching, key, vary string, context core.Context, next core.Next) {
		go c.grace(key, vary, context, next)
	}
	purgeHitResponse  = core.Respond([]byte("")).Status(200)
	purgeMissResponse = core.Respond([]byte("")).Status(204)
)

func (c *Caching) Run(context core.Context, next core.Next) core.Response {
	caching := context.Route().Caching
	if caching == nil {
		c.logger.Info(context, "not cacheable")
		return next(context)
	}

	request := context.RequestIn()
	switch request.Method {
	case "GET":
		return c.get(context, caching, request, next)
	case "PURGE":
		return c.purge(context, caching, request)
	default:
		c.logger.Info(context, "not cacheable")
		return next(context)
	}
}

func (c *Caching) get(context core.Context, caching *core.Caching, request *http.Request, next core.Next) core.Response {
	key, vary := caching.KeyGenerator(context)
	cached := c.cache.Get(key, vary)
	if cached != nil {
		now := time.Now()
		if cached.Expires.After(now) {
			c.logger.Info(context, "hit")
			return cached
		}
		if cached.Expires.Add(c.Configuration.grace).After(now) {
			c.logger.Info(context, "grace")
			grace(c, key, vary, context, next)
			return cached
		}
	}

	c.logger.Info(context, "miss")
	response := next(context)
	if response.GetStatus() >= 500 && c.saint.Nanoseconds() > 0 {
		// log this here since the final handler will never see this 500 error
		core.LogError(c.logger, context, response.GetStatus(), response.GetBody())
		c.logger.Info(context, "saint")
		cached.Expires = time.Now().Add(c.saint)
		return cached
	}
	return c.set(key, vary, context, response)
}

func (c *Caching) set(key, vary string, context core.Context, response core.Response) core.Response {
	ttl, ok := ttl(context.Route().Caching, response)
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

func ttl(caching *core.Caching, response core.Response) (time.Duration, bool) {
	status := response.GetStatus()
	if status >= 200 && status <= 400 && caching.TTL.Nanoseconds() > 0 {
		return caching.TTL, true
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

func (c *Caching) grace(key, vary string, context core.Context, next core.Next) {
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
		c.set(key, vary, context, response)
	} else {
		core.LogError(c.logger, context, response.GetStatus(), response.GetBody())
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

func (c *Caching) purge(context core.Context, caching *core.Caching, request *http.Request) core.Response {
	if c.authorizePurge == nil || c.authorizePurge(context) == false {
		c.logger.Info(context, "unauthorized purge")
		return core.Unauthorized
	}
	key, _ := caching.KeyGenerator(context)
	if c.cache.Delete(key) {
		c.logger.Info(context, "purge hit")
		return purgeHitResponse
	}
	c.logger.Info(context, "purge miss")
	return purgeMissResponse
}
