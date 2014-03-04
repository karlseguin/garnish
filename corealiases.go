package garnish

import (
	"github.com/karlseguin/garnish/gc"
	"github.com/karlseguin/garnish/middleware/access"
	"github.com/karlseguin/garnish/middleware/caching"
	"github.com/karlseguin/garnish/middleware/dispatch"
	"github.com/karlseguin/garnish/middleware/hydrate"
	"github.com/karlseguin/garnish/middleware/stats"
	"github.com/karlseguin/garnish/middleware/upstream"
)

var (
	Stats    = stats.Configure()
	Upstream = upstream.Configure()
	Caching  = caching.Configure()
	Dispatch = dispatch.Configure()
	Hydrate  = hydrate.Configure()
	Access   = access.Configure()

	Unauthorized  = gc.Unauthorized
	NotFound      = gc.NotFound
	InternalError = gc.InternalError
	Json          = gc.Json
	Respond       = gc.Respond
	RespondH      = gc.RespondH
	Fatal         = gc.Fatal
	Redirect      = gc.Redirect
)
