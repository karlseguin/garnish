package caches

import (
	"github.com/karlseguin/garnish"
	"time"
)

type CachedResponse struct {
	Expires time.Time
	garnish.Response
}
