package gc

import (
	"github.com/karlseguin/router"
)

type Runtime struct {
	Executor  MiddlewareExecutor
	Upstreams map[string]*Upstream
	Routes    map[string]*Route
	Router    *router.Router
}
