package hydrate

import (
	"github.com/karlseguin/bytepool"
	"github.com/karlseguin/garnish/gc"
)

type Hydrator func(id []byte) []byte

// Configuration for the Hydrate Parser middleware
type Configuration struct {
	hydrator     Hydrator
	pool         *bytepool.Pool
	overriding   *RouteConfig
	routeConfigs map[string]*RouteConfig
}

func Configure() *Configuration {
	return &Configuration{
		pool:         bytepool.New(128, 65536),
		routeConfigs: make(map[string]*RouteConfig),
	}
}

// Create the middleware from the configuration
func (c *Configuration) Create(config gc.Configuration) (gc.Middleware, error) {
	for name, _ := range config.Router().Routes() {
		if _, ok := c.routeConfigs[name]; ok == false {
			c.routeConfigs[name] = newRouteConfig(c)
		}
	}
	return &Hydrate{
		pool:         c.pool,
		routeConfigs: c.routeConfigs,
	}, nil
}

// The pool to use for a response
func (c *Configuration) ResponsePool(count, size int) *Configuration {
	c.pool = bytepool.New(count, size)
	return c
}

func (c *Configuration) For(routeName string, hydrator Hydrator) {
	c.routeConfigs[routeName].hydrator = hydrator
}

func (c *Configuration) OverrideFor(route *gc.Route) {
	routeConfig := newRouteConfig(c)
	c.routeConfigs[route.Name] = routeConfig
	c.overriding = routeConfig
}

type RouteConfig struct {
	hydrator Hydrator
}

func newRouteConfig(c *Configuration) *RouteConfig {
	return &RouteConfig{
		hydrator: c.hydrator,
	}
}
