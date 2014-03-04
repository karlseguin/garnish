package access

import (
	"github.com/karlseguin/garnish/gc"
)

// 401 wil be returned when the Authenticator returns nil
type Authenticator func(context gc.Context) gc.User

// 401 wil be returned when the Authorizer returns nil
type Authorizer func(context gc.Context) bool

// Configuration for the Access middleware
type Configuration struct {
	authenticator Authenticator
	authorizer    Authorizer
	overriding    *RouteConfig
	routeConfigs  map[string]*RouteConfig
	error         error
	permission    string
}

func Configure() *Configuration {
	return &Configuration{
		routeConfigs: make(map[string]*RouteConfig),
	}
}

func (c *Configuration) Name() string {
	return "access"
}

// Create the middleware from the configuration
func (c *Configuration) Create(config gc.Configuration) (gc.Middleware, error) {
	if c.error != nil {
		return nil, c.error
	}
	println(c.authenticator)
	for name, _ := range config.Router().Routes() {
		if _, ok := c.routeConfigs[name]; ok == false {
			c.routeConfigs[name] = newRouteConfig(c)
		}
	}
	return &Access{
		routeConfigs: c.routeConfigs,
	}, nil
}

// The callback which will authenticate the user
//
// Can be set globally or on a per-route basis
//
// [nil]
func (c *Configuration) Authenticator(a Authenticator) *Configuration {
	if c.overriding != nil {
		c.overriding.authenticator = a
	} else {
		c.authenticator = a
	}
	return c
}

// The callback which will authorize the user
//
// Can be set globally or on a per-route basis
//
// [nil]
func (c *Configuration) Authorizer(a Authorizer) *Configuration {
	if c.overriding != nil {
		c.overriding.authorizer = a
	} else {
		c.authorizer = a
	}
	return c
}

// Permissions is a built-in authorization scheme which is meant to accomodate simple
// permission-based access to resources. The Authorizer callback can be used for
// more advanced scenarios
//
//
// [nil]
func (c *Configuration) Permission(permission string) *Configuration {
	if c.overriding != nil {
		c.overriding.permission = permission
		c.overriding.authorizer = nil
	} else {
		c.permission = permission
	}
	return c
}

func (c *Configuration) OverrideFor(route *gc.Route) {
	routeConfig := newRouteConfig(c)
	c.routeConfigs[route.Name] = routeConfig
	c.overriding = routeConfig
}

type RouteConfig struct {
	authenticator Authenticator
	authorizer    Authorizer
	permission    string
}

func newRouteConfig(c *Configuration) *RouteConfig {
	return &RouteConfig{
		authenticator: c.authenticator,
		authorizer:    c.authorizer,
		permission:    c.permission,
	}
}
