package dispatch

import (
	"github.com/karlseguin/garnish/core"
)

type Dispatch struct {
	dispatcher Dispatcher
	logger     core.Logger
	actions    map[string]interface{}
}

func (d *Dispatch) Name() string {
	return "dispatch"
}

func (d *Dispatch) Run(context core.Context, next core.Next) core.Response {
	route := context.Route()
	action, exists := d.actions[route.Name]
	if exists == false {
		d.logger.Info(context, "404")
		return next(context)
	}
	res := d.dispatcher(action, context)
	if res != nil {
		return res
	}
	return next(context)
}
