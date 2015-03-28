package middlewares

import (
	"gopkg.in/karlseguin/garnish.v1"
)

var Catch = func(req *garnish.Request) garnish.Response {
	return req.Runtime.NotFoundResponse
}
