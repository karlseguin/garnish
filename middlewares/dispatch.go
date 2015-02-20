package middlewares

import (
	"github.com/karlseguin/garnish/gc"
)

func Dispatch(req *gc.Request, next gc.Middleware) gc.Response {
	if h := req.Route.Handler; h != nil {
		return h(req, next)
	}
	return next(req)
}
