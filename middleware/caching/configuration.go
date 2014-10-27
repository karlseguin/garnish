package caching

import (
	"errors"
	"github.com/karlseguin/garnish/gc"
	"github.com/karlseguin/ccache"
	"strings"
	"time"
)

// Authorize a purge request
type AuthorizePurge func(context gc.Context) bool

// Generate a cache key and the vary parameters
type KeyGenerator func(context gc.Context) (string, string)

// Callback to call at runtime which can decide to skip checking the cache
type RuntimeSkip func(context gc.Context) bool

// Configuration for the Caching middleware
type Configuration struct {
	overriding     *RouteConfig
	cache          *ccache.LayeredCache
	grace          time.Duration
	saint          time.Duration
	ttl            time.Duration
	runtimeSkip    RuntimeSkip
	keyGenerator   KeyGenerator
	routeConfigs   map[string]*RouteConfig
	authorizePurge AuthorizePurge
	error          error
}

func Configure() *Configuration {
	return &Configuration{
		grace:        time.Second * 10,
		saint:        time.Hour * 4,
		ttl:          time.Minute,
		keyGenerator: UrlKeyGenerator,
		routeConfigs: make(map[string]*RouteConfig),
	}
}

// Create the middleware from the configuration
func (c *Configuration) Create(config gc.Configuration) (gc.Middleware, error) {
	if c.cache == nil {
		return nil, errors.New("Cannot using caching middleware without specifying a Cache")
	}
	if c.error != nil {
		return nil, c.error
	}
	for name, _ := range config.Router().Routes() {
		if _, ok := c.routeConfigs[name]; ok == false {
			c.routeConfigs[name] = newRouteConfig(c)
		}
	}
	return &Caching{
		cache:        c.cache,
		routeConfigs: c.routeConfigs,
		runtimeSkip:  c.runtimeSkip,
		downloading:  make(map[string]time.Time),
	}, nil
}

// Serve a request even if it has expired within the specified
// duration. A background worker will refresh the data.
// This prevents the thundering herd problem that can happen when
// a resource suddenly expires.
//
// Can be set globally or on a per-route basis
//
// May not be available in all storage engines.
//
// [10 seconds]
func (c *Configuration) Grace(duration time.Duration) *Configuration {
	if c.overriding != nil {
		c.overriding.grace = duration
	} else {
		c.grace = duration
	}
	return c
}

// Serve a request regardless of how stale it is should the upstream
// return a 500+ status code. The duration specifies how long to
// extend the cache for before trying again.
//
// Can be set globally or on a per-route basis
//
// May not be available in all storage engines.
// [4 hours]
func (c *Configuration) Saint(duration time.Duration) *Configuration {
	if c.overriding != nil {
		c.overriding.saint = duration
	} else {
		c.saint = duration
	}
	return c
}

// A callback method to authorize a PURGE request. If this method returns false
// a 401 will be returned. Else it'll return 200 on a successful purge (the
// purged item was in the cache) or a 204 (the purged item was not in the cache)
//
// Can be set globally or on a per-route basis
//
// [nil - purging is not enabled]
func (c *Configuration) AuthorizePurge(authorizePurge AuthorizePurge) *Configuration {
	if c.overriding != nil {
		c.overriding.authorizePurge = authorizePurge
	} else {
		c.authorizePurge = authorizePurge
	}
	return c
}

// The callback to use to generate a primary and secondary cache key for a given
// request
//
// Can be set globally or on a per-route basis
//
// [UrlKeyGenerator]
func (c *Configuration) KeyGenerator(keyGenerator KeyGenerator) *Configuration {
	if c.overriding != nil {
		c.overriding.keyGenerator = keyGenerator
	} else {
		c.keyGenerator = keyGenerator
	}
	return c
}

// The length of time to cache a request
//
// Can be set globally or on a per-route basis
//
// [1 minute]
func (c *Configuration) TTL(ttl time.Duration) *Configuration {
	if c.overriding != nil {
		c.overriding.ttl = ttl
	} else {
		c.ttl = ttl
	}
	return c
}

func (c *Configuration) Never() *Configuration {
	return c.TTL(0)
}

// A callback to call on each request which can be used to skip checking the
// cache. A common use case for this might be to skip the cache for certain
// user roles or if a nocache parameter is present in the query string.
//
// Can be set globally
//
// [nil]
func (c *Configuration) RuntimeSkip(callback RuntimeSkip) *Configuration {
	if c.overriding != nil {
		c.error = errors.New("runtimeSkip cannot be specified on a per-route basis")
	}
	c.runtimeSkip = callback
	return c
}

func (c *Configuration) OverrideFor(route *gc.Route) {
	routeConfig := newRouteConfig(c)
	c.routeConfigs[route.Name] = routeConfig
	c.overriding = routeConfig
}

type RouteConfig struct {
	keyGenerator   KeyGenerator
	grace          time.Duration
	saint          time.Duration
	ttl            time.Duration
	authorizePurge AuthorizePurge
}

func newRouteConfig(c *Configuration) *RouteConfig {
	return &RouteConfig{
		keyGenerator:   c.keyGenerator,
		grace:          c.grace,
		saint:          c.saint,
		ttl:            c.ttl,
		authorizePurge: c.authorizePurge,
	}
}

func UrlKeyGenerator(context gc.Context) (string, string) {
	url := context.Request().URL
	path := url.Path
	secondary := url.RawQuery
	if extensionIndex := strings.LastIndex(path, "."); extensionIndex != -1 {
		secondary += path[extensionIndex:]
		path = path[0:extensionIndex]
	}
	return url.Host + path, secondary
}
