package middlewares

import (
	"gopkg.in/karlseguin/garnish.v1/gc"
)

func Dispatch(req *gc.Request, next gc.Middleware) gc.Response {
	if h := req.Route.Handler; h != nil {
		return h(req, next)
	}
	return next(req)
}
