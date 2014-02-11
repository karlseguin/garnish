package dispatcher

import (
	"github.com/karlseguin/garnish/core"
)

type Dispatcher struct {
	dispatch Dispatch
	logger   core.Logger
	actions  map[string]interface{}
}

func (d *Dispatcher) Name() string {
	return "dispatcher"
}

func (d *Dispatcher) Run(context core.Context, next core.Next) core.Response {
	route := context.Route()
	action, exists := d.actions[route.Name]
	if exists == false {
		d.logger.Info(context, "404")
		return next(context)
	}
	d.logger.Info(context, "+ ", route.Name)
	res := d.dispatch(action, context)
	d.logger.Info(context, "- ", route.Name)
	if res != nil {
		return res
	}
	return next(context)
}
