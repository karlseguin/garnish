package caching

import (
	"github.com/karlseguin/garnish"
	"github.com/karlseguin/garnish/caches"
	"time"
)

// Configuration for the Caching middleware
type Configuration struct {
	logger garnish.Logger
	cache  caches.Cache
	grace  time.Duration
	saint  time.Duration
}

func Configure(base *garnish.Configuration, cache caches.Cache) *Configuration {
	return &Configuration{
		logger: base.Logger,
		grace:  time.Second * 10,
		saint:  time.Second * 10,
		cache:  cache,
	}
}

// Create the middleware from the configuration
func (c *Configuration) Create() (garnish.Middleware, error) {
	return &Caching{Configuration: c, downloading: make(map[string]time.Time)}, nil
}

// Serve a request even if it has expired within the specified
// duration. A background worker will refresh the data.
// This prevents the thundering herd problem that can happen when
// a resource suddenly expires.
//
// May not be available in all storage engines.
//
// [10 seconds]
func (c *Configuration) Grace(duration time.Duration) *Configuration {
	c.grace = duration
	return c
}

// Serve a request regardless of how stale it is should the upstream
// return a 500+ status code. The duration specifies how long to
// extend the cache for before trying again.
//
// May not be available in all storage engines.
//
// [10 second]
func (c *Configuration) Saint(duration time.Duration) *Configuration {
	c.saint = duration
	return c
}
