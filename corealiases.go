package garnish

import (
	"github.com/karlseguin/garnish/gc"
)

type Response gc.Response
type Context gc.Context

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
