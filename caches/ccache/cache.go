package ccache

import (
	"container/list"
	"github.com/karlseguin/garnish/caches"
	"hash/fnv"
	"runtime"
	"time"
)

type Cache struct {
	*Configuration
	list        *list.List
	buckets     []*Bucket
	bucketCount uint32
	deletables  chan *Item
	promotables chan *Item
}

func New(config *Configuration) *Cache {
	c := &Cache{
		list:          list.New(),
		Configuration: config,
		bucketCount:   uint32(config.buckets),
		buckets:       make([]*Bucket, config.buckets),
		deletables:    make(chan *Item, config.deleteBuffer),
		promotables:   make(chan *Item, config.promoteBuffer),
	}
	for i := 0; i < config.buckets; i++ {
		c.buckets[i] = &Bucket{
			lookup: make(map[string]*Vary),
		}
	}
	go c.worker()
	return c
}

func (c *Cache) Get(key, vary string) *caches.CachedResponse {
	bucket := c.bucket(key)
	item := bucket.get(key, vary)
	if item == nil {
		return nil
	}
	if item.value.Expires.After(time.Now()) {
		c.conditionalPromote(item)
	}
	return item.value
}

func (c *Cache) Set(key, vary string, duration time.Duration, value *caches.CachedResponse) {
	item, new := c.bucket(key).set(key, vary, value, duration)
	if new {
		c.promote(item)
	} else {
		c.conditionalPromote(item)
	}
}

func (c *Cache) Delete(key string) bool {
	vary := c.bucket(key).getAndDelete(key)
	if vary == nil {
		return false
	}

	for _, item := range vary.lookup {
		c.deletables <- item
	}
	return true
}

func (c *Cache) DeleteVary(key, vary string) bool {
	item := c.bucket(key).getAndDeleteVary(key, vary)
	if item == nil {
		return false
	}
	c.deletables <- item
	return true
}

//this isn't thread safe. It's meant to be called from non-concurrent tests
func (c *Cache) Clear() {
	for _, bucket := range c.buckets {
		bucket.clear()
	}
	c.list = list.New()
}

func (c *Cache) deleteItem(bucket *Bucket, item *Item) {
	bucket.delete(item.key) //stop othe GETs from getting it
	c.deletables <- item
}

func (c *Cache) bucket(key string) *Bucket {
	h := fnv.New32a()
	h.Write([]byte(key))
	index := h.Sum32() % c.bucketCount
	return c.buckets[index]
}

func (c *Cache) conditionalPromote(item *Item) {
	if item.shouldPromote(c.getsPerPromote) == false {
		return
	}
	c.promote(item)
}

func (c *Cache) promote(item *Item) {
	c.promotables <- item
}

func (c *Cache) worker() {
	ms := new(runtime.MemStats)
	for {
		select {
		case item := <-c.promotables:
			wasNew := c.doPromote(item)
			if wasNew == false {
				continue
			}
			runtime.ReadMemStats(ms)
			if ms.HeapAlloc > c.size {
				c.gc()
			}
		case item := <-c.deletables:
			c.list.Remove(item.element)
		}
	}
}

func (c *Cache) doPromote(item *Item) bool {
	item.Lock()
	defer item.Unlock()
	item.promotions = 0
	if item.element != nil { //not a new item
		c.list.MoveToFront(item.element)
		return false
	}
	item.element = c.list.PushFront(item)
	return true
}

func (c *Cache) gc() {
	for i := 0; i < c.itemsToPrune; i++ {
		element := c.list.Back()
		if element == nil {
			return
		}
		item := element.Value.(*Item)
		c.bucket(item.key).deleteVary(item.key, item.vary)
		c.list.Remove(element)
	}
}
