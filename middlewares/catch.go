package middlewares

import (
	"github.com/karlseguin/garnish/gc"
)

func Catch(req *gc.Request, next gc.Middleware) gc.Response {
	return gc.NotFoundResponse
}
