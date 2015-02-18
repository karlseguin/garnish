package cache

import (
	"github.com/karlseguin/garnish/gc"
	"hash/fnv"
)

const (
	BUCKETS     = 16
	BUCKET_MASK = BUCKETS - 1
)

type Entry struct {
	gc.CachedResponse
	Primary   string
	Secondary string
	next      *Entry
	prev      *Entry
	size      int
}

type Cache struct {
	list        *List
	maxSize     int
	size        int
	buckets     []*bucket
	deletables  chan *Entry
	promotables chan *Entry
}

func New(maxSize int) *Cache {
	c := &Cache{
		maxSize:     maxSize,
		list:        NewList(),
		buckets:     make([]*bucket, BUCKETS),
		deletables:  make(chan *Entry, 1024),
		promotables: make(chan *Entry, 1024),
	}
	for i := 0; i < BUCKETS; i++ {
		c.buckets[i] = &bucket{lookup: make(map[string]map[string]*Entry)}
	}
	go c.worker()
	return c
}

func (c *Cache) Get(primary, secondary string) gc.CachedResponse {
	bucket := c.bucket(primary)
	response := bucket.get(primary, secondary)
	if response == nil {
		return nil
	}
	c.promotables <- response
	return response
}

func (c *Cache) Set(primary string, secondary string, response gc.CachedResponse) {
	entry := &Entry{
		Primary:        primary,
		Secondary:      secondary,
		CachedResponse: response,
		size:           response.Size(),
	}
	existing := c.bucket(primary).set(primary, secondary, entry)
	if existing != nil {
		c.deletables <- existing
	}
	c.promotables <- entry
}

func (c *Cache) Delete(primary string, secondary string) bool {
	return c.bucket(primary).delete(primary, secondary)
}

func (c *Cache) DeleteAll(primary string) bool {
	return c.bucket(primary).deleteAll(primary)
}

func (c *Cache) bucket(key string) *bucket {
	h := fnv.New32a()
	h.Write([]byte(key))
	return c.buckets[h.Sum32()&BUCKET_MASK]
}

func (c *Cache) worker() {
	for {
		select {
		case entry := <-c.promotables:
			if entry.prev == nil { //new item
				c.size += entry.size
				if c.size > c.maxSize {
					c.gc()
				}
			}
			c.list.PushToFront(entry)
		case entry := <-c.deletables:
			if entry.prev != nil {
				c.list.Remove(entry)
				c.size -= entry.size
			}
		}
	}
}

func (c *Cache) gc() {
	for i := 0; i < 1000; i++ {
		entry := c.list.tail.prev
		if entry == nil {
			return
		}
		primary := entry.Primary
		if len(primary) == 0 {
			return
		}
		c.bucket(primary).delete(primary, entry.Secondary)
		c.list.Remove(entry)
		c.size -= entry.size
	}
}
