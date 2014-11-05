package middlewares

import (
	"github.com/karlseguin/garnish/gc"
)

func HydrateLoader(req *gc.Request, next gc.Middleware) gc.Response {
	res := next(req)
	return res
}

func HydrateParser(req *gc.Request, next gc.Middleware) gc.Response {
	res := next(req)
	return res
}
