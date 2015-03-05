package garnish

import (
	"errors"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/karlseguin/garnish/configurations"
	"github.com/karlseguin/garnish/gc"
	"github.com/karlseguin/garnish/middlewares"
	"gopkg.in/karlseguin/bytepool.v3"
	"gopkg.in/karlseguin/dnscache.v1"
	"gopkg.in/karlseguin/router.v1"
	"gopkg.in/karlseguin/typed.v1"
	"time"
)

type MiddlewarePosition int

const (
	BEFORE_STATS MiddlewarePosition = iota
	BEFORE_CACHE
	BEFORE_HYDRATE
	BEFORE_DISPATCH
)

// Configuration
type Configuration struct {
	address   string
	stats     *configurations.Stats
	router    *configurations.Router
	upstreams *configurations.Upstreams
	cache     *configurations.Cache
	hydrate   *configurations.Hydrate
	bytePool  poolConfiguration
	dnsTTL    time.Duration
	before    map[MiddlewarePosition]struct {
		name    string
		handler gc.Handler
	}
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
		before: make(map[MiddlewarePosition]struct {
			name    string
			handler gc.Handler
		}),
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

func (c *Configuration) Insert(position MiddlewarePosition, name string, handler gc.Handler) *Configuration {
	c.before[position] = struct {
		name    string
		handler gc.Handler
	}{
		name,
		handler,
	}
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

// Build a runtime object from the configuration, which can then be
// used to start garnish
func (c *Configuration) Build() (*gc.Runtime, error) {
	runtime := &gc.Runtime{
		Address:  c.address,
		Resolver: dnscache.New(c.dnsTTL),
		Router:   router.New(router.Configure()),
		Executor: gc.WrapMiddleware("upst", middlewares.Upstream, nil),
	}

	if err := c.upstreams.Build(runtime); err != nil {
		return nil, err
	}

	if c.router == nil {
		return nil, errors.New("Atleast one route must be configured")
	}

	if err := c.router.Build(runtime); err != nil {
		return nil, err
	}

	runtime.Executor = gc.WrapMiddleware("dspt", middlewares.Dispatch, runtime.Executor)
	if h, ok := c.before[BEFORE_DISPATCH]; ok {
		runtime.Executor = gc.WrapMiddleware(h.name, h.handler, runtime.Executor)
	}

	if c.hydrate != nil {
		m, err := c.hydrate.Build(runtime)
		if err != nil {
			return nil, err
		}
		runtime.Executor = gc.WrapMiddleware("hdrt", m.Handle, runtime.Executor)
	}
	if h, ok := c.before[BEFORE_HYDRATE]; ok {
		runtime.Executor = gc.WrapMiddleware(h.name, h.handler, runtime.Executor)
	}

	if c.cache != nil {
		if err := c.cache.Build(runtime); err != nil {
			return nil, err
		}
		runtime.Executor = gc.WrapMiddleware("cach", middlewares.Cache, runtime.Executor)
	}
	if h, ok := c.before[BEFORE_CACHE]; ok {
		runtime.Executor = gc.WrapMiddleware(h.name, h.handler, runtime.Executor)
	}

	if c.stats != nil {
		if err := c.stats.Build(runtime); err != nil {
			return nil, err
		}
		runtime.Executor = gc.WrapMiddleware("stat", middlewares.Stats, runtime.Executor)
	}
	if h, ok := c.before[BEFORE_STATS]; ok {
		runtime.Executor = gc.WrapMiddleware(h.name, h.handler, runtime.Executor)
	}

	runtime.BytePool = bytepool.New(c.bytePool.capacity, c.bytePool.count)
	runtime.RegisterStats("bytepool", runtime.BytePool.Stats)
	return runtime, nil
}

// Loads configuration from a toml file
func LoadConfig(path string) (*Configuration, error) {
	var t typed.Typed
	if _, err := toml.DecodeFile(path, &t); err != nil {
		return nil, err
	}

	config := Configure()
	if a, ok := t.StringIf("address"); ok {
		config.Address(a)
	}
	if t.BoolOr("debug", false) {
		config.Debug()
	}

	for _, ut := range t.Objects("upstreams") {
		upstream := config.Upstream(ut.String("name"))
		if n, ok := ut.IntIf("dns"); ok {
			upstream.DnsCache(time.Second * time.Duration(n))
		}
		if h, ok := ut.StringsIf("headers"); ok {
			upstream.Headers(h...)
		}
		for _, tt := range ut.Objects("transports") {
			transport := upstream.Address(tt.String("address"))
			if n, ok := tt.IntIf("keepalive"); ok {
				transport.KeepAlive(uint32(n))
			}
		}
	}

	fmt.Println(t)

	return config, nil
}
