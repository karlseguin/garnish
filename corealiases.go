package garnish

import (
	"github.com/karlseguin/garnish/gc"
	"github.com/karlseguin/garnish/middleware/stats"
)

type Response gc.Response
type Context gc.Context

func RegisterReporter(name string, reporter stats.Reporter) {
	stats.RegisterReporter(name, reporter)
}

var (
	Unauthorized  = gc.Unauthorized
	NotFound      = gc.NotFound
	InternalError = gc.InternalError
	Json          = gc.Json
	Respond       = gc.Respond
	RespondH      = gc.RespondH
	Fatal         = gc.Fatal
	Redirect      = gc.Redirect
)
