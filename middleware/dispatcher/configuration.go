package dispatcher

import (
	"github.com/karlseguin/garnish/core"
)

type Dispatch func(action interface{}, context core.Context) core.Response

// Configuration for upstreams middleware
type Configuration struct {
	overriding string
	actions    map[string]interface{}
	dispatch   Dispatch
}

func Configure() *Configuration {
	return &Configuration{
		actions: make(map[string]interface{}),
	}
}

// Create the middleware from the configuration
func (c *Configuration) Create(config core.Configuration) (core.Middleware, error) {
	return &Dispatcher{
		actions:  c.actions,
		dispatch: c.dispatch,
		logger:   config.Logger(),
	}, nil
}

// Create the middleware from the configuration
func (c *Configuration) Action(name string, action interface{}) *Configuration {
	c.actions[name] = action
	return c
}

func (c *Configuration) Dispatch(dispatch Dispatch) *Configuration {
	c.dispatch = dispatch
	return c
}

func (c *Configuration) OverrideFor(route *core.Route) {
	c.overriding = route.Name
}

func (c *Configuration) To(action interface{}) {
	c.actions[c.overriding] = action
}
