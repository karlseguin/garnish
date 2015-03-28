package middlewares

import (
	"gopkg.in/karlseguin/garnish.v1"
)

func Dispatch(req *garnish.Request, next garnish.Middleware) garnish.Response {
	if h := req.Route.Handler; h != nil {
		return h(req, next)
	}
	return next(req)
}
