package garnish

import (
	"github.com/karlseguin/bytepool"
	"github.com/karlseguin/garnish/configurations"
	"github.com/karlseguin/garnish/gc"
	"github.com/karlseguin/garnish/middlewares"
	"github.com/karlseguin/router"
)

type Configuration struct {
	address   string
	upstreams *configurations.Upstreams
	router    *configurations.Router
	bytePool  PoolConfiguration
}

type PoolConfiguration struct {
	capacity int
	size     int
}

func Configure() *Configuration {
	return &Configuration{
		address:  ":8080",
		bytePool: PoolConfiguration{65536, 64},
	}
}

// The address to listen on
func (c *Configuration) Address(address string) *Configuration {
	c.address = address
	return c
}

// Enable debug-level logging
func (c *Configuration) Debug() *Configuration {
	gc.Log.Verbose()
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

// Specify a custom logger to use
func (c *Configuration) Logger(logger gc.Logs) *Configuration {
	gc.Log = logger
	return c
}

// Define an upstream
func (c *Configuration) Upstream(name string) *configurations.Upstream {
	if c.upstreams == nil {
		c.upstreams = configurations.NewUpstreams()
	}
	return c.upstreams.Add(name)
}

func (c *Configuration) Route(name string) *configurations.Route {
	if c.router == nil {
		c.router = configurations.NewRouter()
	}
	return c.router.Add(name)
}

func (c *Configuration) Build() *gc.Runtime {
	ok := true
	logger := gc.Log
	if c.upstreams == nil {
		logger.Error("Atleast one upstream must be configured")
		return nil
	}

	catch := gc.WrapMiddleware("catch", middlewares.Catch, nil)
	runtime := &gc.Runtime{
		Router:   router.New(router.Configure()),
		Executor: gc.WrapMiddleware("upstream", middlewares.Upstream, catch),
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

	gc.BytePool = bytepool.New(c.bytePool.capacity, c.bytePool.size)
	gc.BytePoolItemSize = c.bytePool.size

	if ok == false {
		return nil
	}

	return runtime
}
