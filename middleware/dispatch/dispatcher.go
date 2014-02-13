package dispatch

import (
	"github.com/karlseguin/garnish/core"
)

type Dispatch struct {
	dispatcher Dispatcher
	actions    map[string]interface{}
}

func (d *Dispatch) Name() string {
	return "dispatch"
}

func (d *Dispatch) Run(context core.Context, next core.Next) core.Response {
	route := context.Route()
	action, exists := d.actions[route.Name]
	if exists == false {
		context.Info("404")
		return next(context)
	}
	res := d.dispatcher(action, context)
	if res != nil {
		return res
	}
	return next(context)
}
