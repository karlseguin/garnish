package hydrate

import (
	"github.com/karlseguin/bytepool"
	"github.com/karlseguin/garnish/gc"
)

// Configuration for the Hydrate Parser middleware
type Configuration struct {
	pool *bytepool.Pool
}

func Configure() *Configuration {
	return &Configuration{
		pool: bytepool.New(125, 65536),
	}
}

// Create the middleware from the configuration
func (c *Configuration) Create(config gc.Configuration) (gc.Middleware, error) {
	return &Hydrate{
		pool: c.pool,
	}, nil
}

// The pool to use for a response
func (c *Configuration) ResponsePool(count, size int) *Configuration {
	c.pool = bytepool.New(count, size)
	return c
}

func (c *Configuration) OverrideFor(route *gc.Route) {

}
