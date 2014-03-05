package dispatch

import (
	"github.com/karlseguin/garnish/gc"
)

type Dispatch struct {
	dispatcher Dispatcher
	actions    map[string]interface{}
}

func (d *Dispatch) Name() string {
	return "dispatch"
}

func (d *Dispatch) Run(context gc.Context, next gc.Next) gc.Response {
	route := context.Route()
	action, exists := d.actions[route.Name]
	if exists == false {
		return next(context)
	}
	res := d.dispatcher(action, context)
	if res != nil {
		return res
	}
	return next(context)
}
