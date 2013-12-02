// Middleware package that implements caching.
package caching

import (
	"github.com/karlseguin/garnish"
	"github.com/karlseguin/garnish/caches"
	"strconv"
	"strings"
	"time"
)

type Caching struct {
	*Configuration
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
		if now.Add(c.grace).After(cached.Expires) {
			c.logger.Info(context, "grace")
			//todo grace fetch
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
	return c.set(caching, response, context, key, vary)
}

func (c *Caching) set(caching *garnish.Caching, response garnish.Response, context garnish.Context, key, vary string) garnish.Response {
	ttl, ok := ttl(caching, response)
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
