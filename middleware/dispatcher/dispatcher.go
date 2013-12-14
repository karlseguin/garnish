package dispatcher

import (
	"github.com/karlseguin/garnish"
)

type Dispatcher struct {
	*Configuration
}

func (d *Dispatcher) Name() string {
	return "dispatcher"
}

func (d *Dispatcher) Run(context garnish.Context, next garnish.Next) garnish.Response {
	route := context.Route()
	action, exists := d.actions[route.Upstream]
	if exists == false {
		d.logger.Info(context, "404")
		return next(context)
	}
	d.logger.Info(context, "+ ", route.Upstream)
	res := d.dispatch(action, context)
	d.logger.Info(context, "- ", route.Upstream)
	if res != nil {
		return res
	}
	return next(context)
}
