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
	saint  bool
}

func Configure(base *garnish.Configuration, cache caches.Cache) *Configuration {
	return &Configuration{
		logger: base.Logger,
		grace:  time.Second * 10,
		saint:  true,
		cache:  cache,
	}
}

// Create the middleware from the configuration
func (c *Configuration) Create() (garnish.Middleware, error) {
	return &Caching{c}, nil
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
// return a 500+ status code.
//
// May not be available in all storage engines.
//
// [enabled]
func (c *Configuration) Saintless() *Configuration {
	c.saint = false
	return c
}
