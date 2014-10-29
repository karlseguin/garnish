package middlewares

import (
	"github.com/karlseguin/garnish/gc"
)

var Catch = func(req *gc.Request) gc.Response {
	return gc.NotFoundResponse
}
