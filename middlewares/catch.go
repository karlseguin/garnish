package middlewares

import (
	"gopkg.in/karlseguin/garnish.v1/gc"
)

var Catch = func(req *gc.Request) gc.Response {
	return req.Runtime.NotFoundResponse
}
