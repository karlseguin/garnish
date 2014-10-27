package garnish

import (
	"github.com/karlseguin/garnish/configurations"
	"github.com/karlseguin/garnish/gc"
	"github.com/karlseguin/garnish/middlewares/upstream"
	"github.com/karlseguin/router"
	"github.com/op/go-logging"
)

type Configuration struct {
	address   string
	upstreams *configurations.Upstreams
	router    *configurations.Router
	e         gc.MiddlewareExecutor
}

func Configure() *Configuration {
	return &Configuration{
		address: ":8080",
	}
}

// The address to listen on
func (c *Configuration) Address(address string) *Configuration {
	c.address = address
	return c
}

// Enable debug-level logging
func (c *Configuration) Debug() *Configuration {
	logging.SetLevel(logging.DEBUG, "garnish")
	return c
}

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
	logger := gc.Logger
	if c.upstreams == nil {
		logger.Error("Atleast one upstream must be configured")
		return nil
	}

	runtime := &gc.Runtime{
		Router:   router.New(router.Configure()),
		Executor: gc.WrapMiddleware(upstream.Handler, nil),
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

	if ok == false {
		return nil
	}

	return runtime
}
