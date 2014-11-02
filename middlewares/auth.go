package middlewares

import (
	"github.com/karlseguin/garnish/gc"
)

func Auth(req *gc.Request, next gc.Middleware) gc.Response {
	if res := req.Runtime.AuthHandler(req); res != nil {
		return res
	}
	return next(req)
}
