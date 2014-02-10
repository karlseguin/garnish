package caches

import (
	"github.com/karlseguin/garnish/core"
	"time"
)

type CachedResponse struct {
	Expires time.Time
	core.Response
}
