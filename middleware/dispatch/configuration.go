package dispatch

import (
	"github.com/karlseguin/garnish/core"
)

type Dispatcher func(action interface{}, context core.Context) core.Response

// Configuration for upstreams middleware
type Configuration struct {
	overriding string
	actions    map[string]interface{}
	dispatcher Dispatcher
}

func Configure() *Configuration {
	return &Configuration{
		actions: make(map[string]interface{}),
	}
}

// Create the middleware from the configuration
func (c *Configuration) Create(config core.Configuration) (core.Middleware, error) {
	return &Dispatch{
		actions:    c.actions,
		dispatcher: c.dispatcher,
	}, nil
}

// Create the middleware from the configuration
func (c *Configuration) Action(name string, action interface{}) *Configuration {
	c.actions[name] = action
	return c
}

func (c *Configuration) Dispatch(dispatcher Dispatcher) *Configuration {
	c.dispatcher = dispatcher
	return c
}

func (c *Configuration) OverrideFor(route *core.Route) {
	c.overriding = route.Name
}

func (c *Configuration) To(action interface{}) {
	c.actions[c.overriding] = action
}
