package caching

import (
	"github.com/karlseguin/garnish/caches"
	"github.com/karlseguin/garnish/core"
	"strings"
	"time"
)

type AuthorizePurge func(context core.Context) bool

// Generate a cache key and the vary parameters
type KeyGenerator func(context core.Context) (string, string)

// Configuration for the Caching middleware
type Configuration struct {
	overriding     *RouteConfig
	logger         core.Logger
	cache          caches.Cache
	grace          time.Duration
	saint          time.Duration
	ttl            time.Duration
	keyGenerator   KeyGenerator
	routeConfigs   map[string]*RouteConfig
	authorizePurge AuthorizePurge
}

func Configure(cache caches.Cache) *Configuration {
	return &Configuration{
		grace:        time.Second * 10,
		saint:        time.Hour * 4,
		cache:        cache,
		ttl:          time.Minute,
		keyGenerator: UrlKeyGenerator,
		routeConfigs: make(map[string]*RouteConfig),
	}
}

// Create the middleware from the configuration
func (c *Configuration) Create(config core.Configuration) (core.Middleware, error) {
	c.logger = config.Logger()
	for name, _ := range config.Router().Routes() {
		if _, ok := c.routeConfigs[name]; ok == false {
			c.routeConfigs[name] = newRouteConfig(c)
		}
	}
	return &Caching{
		cache:        c.cache,
		logger:       c.logger,
		routeConfigs: c.routeConfigs,
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

func (c *Configuration) OverrideFor(route *core.Route) {
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

func UrlKeyGenerator(context core.Context) (string, string) {
	url := context.Request().URL
	path := url.Path
	secondary := url.RawQuery
	if extensionIndex := strings.LastIndex(path, "."); extensionIndex != -1 {
		secondary += path[extensionIndex:]
		path = path[0:extensionIndex]
	}
	return url.Host + path, secondary
}
