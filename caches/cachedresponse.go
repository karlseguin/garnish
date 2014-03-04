package caches

import (
	"github.com/karlseguin/garnish/gc"
	"time"
)

type CachedResponse struct {
	Expires time.Time
	gc.Response
}
