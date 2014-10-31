package garnish

import (
	"github.com/karlseguin/bytepool"
	"github.com/karlseguin/dnscache"
	"github.com/karlseguin/garnish/configurations"
	"github.com/karlseguin/garnish/gc"
	"github.com/karlseguin/garnish/middlewares"
	"github.com/karlseguin/router"
	"time"
)

// Configuration
type Configuration struct {
	address   string
	stats     *configurations.Stats
	router    *configurations.Router
	upstreams *configurations.Upstreams
	cache     *configurations.Cache
	bytePool  poolConfiguration
	dnsTTL    time.Duration
}

type poolConfiguration struct {
	capacity int
	size     int
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
// [65536, 64]
func (c *Configuration) BytePool(capacity, size uint32) *Configuration {
	c.bytePool.capacity, c.bytePool.size = int(capacity), int(size)
	return c
}

// The default time to cache DNS lookups. Overwritable on a per-upstream basis
// Even a short value (1s) can help under heavy load
// [1 minute]
func (c *Configuration) DnsTTL(ttl time.Duration) *Configuration {
	c.dnsTTL = ttl
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

// Enable and configure the cace middleware
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
	ok := true
	logger := gc.Log
	if c.upstreams == nil {
		logger.Error("Atleast one upstream must be configured")
		return nil
	}

	runtime := &gc.Runtime{
		Resolver: dnscache.New(c.dnsTTL),
		Router:   router.New(router.Configure()),
		Executor: gc.WrapMiddleware("upstream", middlewares.Upstream, nil),
	}

	if c.upstreams.Build(runtime) == false {
		ok = false
	}

	if c.router == nil {
		logger.Error("Atleast one route must be configured")
		return nil
	}

	if c.router.Build(runtime) == false {
		ok = false
	}

	if c.cache != nil {
		if c.cache.Build(runtime) == false {
			ok = false
		} else {
			runtime.Executor = gc.WrapMiddleware("cache", middlewares.Cache, runtime.Executor)
		}
	}

	if c.stats != nil {
		if c.stats.Build(runtime) == false {
			ok = false
		} else {
			runtime.Executor = gc.WrapMiddleware("stats", middlewares.Stats, runtime.Executor)
		}
	}

	if ok == false {
		return nil
	}

	runtime.BytePool = bytepool.New(c.bytePool.capacity, c.bytePool.size)
	runtime.RegisterStats("bytepool", runtime.BytePool.Stats)
	return runtime
}
