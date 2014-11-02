package configurations

import (
	"github.com/karlseguin/garnish/gc"
)

type Auth struct {
	handler gc.AuthHandler
}

func NewAuth(handler gc.AuthHandler) *Auth {
	return &Auth{
		handler: handler,
	}
}

func (a *Auth) Build(runtime *gc.Runtime) bool {
	runtime.AuthHandler = a.handler
	return true
}
