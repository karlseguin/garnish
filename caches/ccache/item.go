package ccache

import (
	"container/list"
	"github.com/karlseguin/garnish/caches"
	"sync"
	"sync/atomic"
	"time"
)

type Item struct {
	key  string
	vary string
	sync.RWMutex
	promotions int32
	value      *caches.CachedResponse
	element    *list.Element
}

func newItem(key, vary string, value *caches.CachedResponse, expires time.Time) *Item {
	return &Item{
		key:        key,
		vary:       vary,
		value:      value,
		promotions: -1,
	}
}

func (i *Item) shouldPromote(getsPerPromote int32) bool {
	return atomic.AddInt32(&i.promotions, 1) == getsPerPromote
}
