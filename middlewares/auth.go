package middlewares

import (
	"github.com/karlseguin/garnish/gc"
)

type Auth struct {
	Handler gc.AuthHandler
}

func (a *Auth) Handle(req *gc.Request, next gc.Middleware) gc.Response {
	if res := a.Handler(req); res != nil {
		return res
	}
	return next(req)
}
