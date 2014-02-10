package garnish

import (
	"github.com/karlseguin/garnish/middleware/stats"
	// "github.com/karlseguin/garnish/middleware/caching"
	"github.com/karlseguin/garnish/middleware/upstream"
	"github.com/karlseguin/garnish/core"
)

var (
	Stats = stats.Configure()
	Upstream = upstream.Configure()
	// Caching = caching.Configure()

	Unauthorized = core.Unauthorized
	NotFound = core.NotFound
	InternalError = core.InternalError
	Json = core.Json
	Respond = core.Respond
	RespondH = core.RespondH
	Fatal = core.Fatal
	Redirect = core.Redirect
)