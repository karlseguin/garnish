package caching

import (
	"github.com/karlseguin/garnish"
	"time"
)

// Configuration for the Caching middleware
type Configuration struct {
	logger  garnish.Logger
	storage Storage
	grace   time.Duration
	saint   bool
}

func Configure(base *garnish.Configuration) *Configuration {
	return &Configuration{
		logger: base.Logger,
		grace:  time.Second * 10,
		saint:  true,
	}
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
