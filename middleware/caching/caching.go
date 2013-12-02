// Middleware package that implements caching.
package caching

import (
	"sync"
	"github.com/karlseguin/garnish"
	"github.com/karlseguin/garnish/caches"
	"strconv"
	"strings"
	"time"
)

type Caching struct {
	*Configuration
	lock sync.RWMutex
	downloading map[string]time.Time
}

func (c *Caching) Name() string {
	return "caching"
}

func (c *Caching) Run(context garnish.Context, next garnish.Next) garnish.Response {
	request := context.RequestIn()
	caching := context.Route().Caching
	if request.Method != "GET" || caching == nil {
		c.logger.Info(context, "not cacheable")
		return next(context)
	}

	key, vary := caching.KeyGenerator(context)
	cached := c.cache.Get(key, vary)
	if cached != nil {
		now := time.Now()
		if cached.Expires.After(now) {
			c.logger.Info(context, "hit")
			return cached
		}
		if now.Add(c.Configuration.grace).After(cached.Expires) {
			c.logger.Info(context, "grace")
			go c.grace(key, vary, context, next)
			return cached
		}
	}

	c.logger.Info(context, "miss")
	response := next(context)
	if response.GetStatus() >= 500 && c.saint {
		// log this here since the final handler will never see this 500 error
		garnish.LogError(c.logger, context, response.GetStatus(), response.GetBody())
		c.logger.Info(context, "saint")
		return cached
	}
	return c.set(key, vary, context, response)
}

func (c *Caching) set(key, vary string, context garnish.Context, response garnish.Response) garnish.Response {
	ttl, ok := ttl(context.Route().Caching, response)
	if ok == false {
		c.logger.Error(context, "configured to cache but no expiry was given")
		return response
	}
	cr := &caches.CachedResponse{
		Expires:  time.Now().Add(ttl),
		Response: response.Detach(),
	}
	c.cache.Set(key, vary, ttl, cr)
	return cr
}

func ttl(caching *garnish.Caching, response garnish.Response) (time.Duration, bool) {
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

func (c *Caching) grace(key, vary string, context garnish.Context, next garnish.Next) {
	downloadKey := key + vary
	if c.isDownloading(downloadKey) { return }
	if c.lockDownload(downloadKey) == false { return }
	defer c.unlockDownload(downloadKey)
	c.logger.Info(context, "grace start download")
	response := next(context)
	if response.GetStatus() < 500 {
		c.logger.Infof(context, "grace %d", response.GetStatus())
		c.set(key, vary, context, response)
	} else {
		garnish.LogError(c.logger, context, response.GetStatus(), response.GetBody())
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
