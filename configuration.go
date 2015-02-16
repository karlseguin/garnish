package garnish

import (
	"github.com/karlseguin/garnish/configurations"
	"github.com/karlseguin/garnish/gc"
	"github.com/karlseguin/garnish/middlewares"
	"gopkg.in/karlseguin/bytepool.v3"
	"gopkg.in/karlseguin/dnscache.v1"
	"gopkg.in/karlseguin/router.v1"
	"time"
)

// Configuration
type Configuration struct {
	address   string
	stats     *configurations.Stats
	router    *configurations.Router
	upstreams *configurations.Upstreams
	cache     *configurations.Cache
	hydrate   *configurations.Hydrate
	auth      *configurations.Auth
	bytePool  poolConfiguration
	dnsTTL    time.Duration
}

type poolConfiguration struct {
	capacity int
	count    int
}

// Create a configuration
func Configure() *Configuration {
	return &Configuration{
		address:  ":8080",
		dnsTTL:   time.Minute,
		bytePool: poolConfiguration{65536, 64},
	}
}

// The address to listen on
// [:8080]
func (c *Configuration) Address(address string) *Configuration {
	c.address = address
	return c
}

// Enable debug-level logging
func (c *Configuration) Debug() *Configuration {
	gc.Log.Verbose()
	return c
}

// Specify a custom logger to use
func (c *Configuration) Logger(logger gc.Logs) *Configuration {
	gc.Log = logger
	return c
}

// The size of each buffer and the number of buffers to keep in a the pool
// Upstream replies with a content length which fit within the specified
// capacity wil perform better.
// The pool is also used for incoming requests with bodies.
// [65536, 64]
func (c *Configuration) BytePool(capacity, count uint32) *Configuration {
	c.bytePool.capacity, c.bytePool.count = int(capacity), int(count)
	return c
}

// The default time to cache DNS lookups. Overwritable on a per-upstream basis
// Even a short value (1s) can help under heavy load
// [1 minute]
func (c *Configuration) DnsTTL(ttl time.Duration) *Configuration {
	c.dnsTTL = ttl
	return c
}

// Enable and configure the auth middleware
func (c *Configuration) Auth(handler gc.AuthHandler) *Configuration {
	c.auth = configurations.NewAuth(handler)
	return c
}

// Enable and configure the stats middleware
func (c *Configuration) Stats() *configurations.Stats {
	if c.router == nil {
		c.stats = configurations.NewStats()
	}
	return c.stats
}

// Configure upstreams
func (c *Configuration) Upstream(name string) *configurations.Upstream {
	if c.upstreams == nil {
		c.upstreams = configurations.NewUpstreams()
	}
	return c.upstreams.Add(name)
}

// Enable and configure the hydration middleware
func (c *Configuration) Hydrate(loader gc.HydrateLoader) *configurations.Hydrate {
	if c.hydrate == nil {
		c.hydrate = configurations.NewHydrate(loader)
	}
	return c.hydrate
}

// Enable and configure the cache middleware
func (c *Configuration) Cache() *configurations.Cache {
	if c.cache == nil {
		c.cache = configurations.NewCache()
	}
	return c.cache
}

// Configure your routes
func (c *Configuration) Route(name string) *configurations.Route {
	if c.router == nil {
		c.router = configurations.NewRouter()
	}
	return c.router.Add(name)
}

// In normal usage, there's no need to call this method.
// Builds a *gc.Runtime based on the configuration. Returns nil on error
// and prints those errors to *gc.Logger
func (c *Configuration) Build() *gc.Runtime {
	logger := gc.Log
	if c.upstreams == nil {
		logger.Error("Atleast one upstream must be configured")
		return nil
	}

	runtime := &gc.Runtime{
		Resolver: dnscache.New(c.dnsTTL),
		Router:   router.New(router.Configure()),
		Executor: gc.WrapMiddleware("upst", middlewares.Upstream, nil),
	}

	if c.upstreams.Build(runtime) == false {
		return nil
	}

	if c.router == nil {
		logger.Error("Atleast one route must be configured")
		return nil
	}

	if c.router.Build(runtime) == false {
		return nil
	}

	if c.hydrate != nil {
		if m := c.hydrate.Build(runtime); m != nil {
			runtime.Executor = gc.WrapMiddleware("hdrt", m.Handler, runtime.Executor)
		} else {
			return nil
		}
	}

	if c.cache != nil {
		if c.cache.Build(runtime) == false {
			return nil
		}
		runtime.Executor = gc.WrapMiddleware("cach", middlewares.Cache, runtime.Executor)
	}

	if c.auth != nil {
		if m := c.auth.Build(runtime); m != nil {
			runtime.Executor = gc.WrapMiddleware("auth", m.Handle, runtime.Executor)
		} else {
			return nil
		}
	}

	if c.stats != nil {
		if c.stats.Build(runtime) == false {
			return nil
		}
		runtime.Executor = gc.WrapMiddleware("stat", middlewares.Stats, runtime.Executor)
	}

	runtime.BytePool = bytepool.New(c.bytePool.capacity, c.bytePool.count)
	runtime.RegisterStats("bytepool", runtime.BytePool.Stats)
	return runtime
}
