// Middleware package that implements caching.
package access

import (
	"github.com/karlseguin/garnish/gc"
)

type Access struct {
	routeConfigs map[string]*RouteConfig
}

func (a *Access) Name() string {
	return "access"
}

func (c *Access) Run(context gc.Context, next gc.Next) gc.Response {
	config := c.routeConfigs[context.Route().Name]
	if config.authenticator == nil {
		return next(context)
	}
	user := config.authenticator(context)
	if user == nil {
		return gc.Unauthorized
	}
	context.SetUser(user)
	if len(config.permission) > 0 && user.Permissions()[config.permission] == false {
		return gc.Unauthorized
	}
	if config.authorizer != nil && config.authorizer(context) == false {
		return gc.Unauthorized
	}
	return next(context)
}
