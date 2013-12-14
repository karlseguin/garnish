package dispatcher

import (
	"github.com/karlseguin/garnish"
)

type Dispatch func(action interface{}, context garnish.Context) garnish.Response

// Configuration for upstreams middleware
type Configuration struct {
	logger   garnish.Logger
	actions  map[string]interface{}
	dispatch Dispatch
}

func Configure(base *garnish.Configuration) *Configuration {
	return &Configuration{
		logger:  base.Logger,
		actions: make(map[string]interface{}),
	}
}

// Create the middleware from the configuration
func (c *Configuration) Create() (garnish.Middleware, error) {
	return &Dispatcher{c}, nil
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
