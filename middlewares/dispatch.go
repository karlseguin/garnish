package middlewares

import (
	"gopkg.in/karlseguin/garnish.v1"
)

func Dispatch(req *garnish.Request, next garnish.Handler) garnish.Response {
	if h := req.Route.StopHandler; h != nil {
		return h(req)
	}
	if h := req.Route.FlowHandler; h != nil {
		return h(req, next)
	}
	return next(req)
}
