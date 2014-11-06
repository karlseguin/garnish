package configurations

import (
	"github.com/karlseguin/garnish/gc"
	"github.com/karlseguin/garnish/middlewares"
)

type Auth struct {
	handler gc.AuthHandler
}

func NewAuth(handler gc.AuthHandler) *Auth {
	return &Auth{
		handler: handler,
	}
}

func (a *Auth) Build(runtime *gc.Runtime) *middlewares.Auth {
	return &middlewares.Auth{
		Handler: a.handler,
	}
}
