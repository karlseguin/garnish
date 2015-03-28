package cache

import (
	"gopkg.in/karlseguin/garnish.v1"
	"hash/fnv"
	"math/rand"
	"time"
)

const (
	BUCKETS     = 16
	BUCKET_MASK = BUCKETS - 1
)

type Entry struct {
	garnish.CachedResponse
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
	persist     chan persist
	deletables  chan *Entry
	promotables chan *Entry
	newSize     chan int
	stop        chan struct{}
}

func New(maxSize int) *Cache {
	c := &Cache{
		maxSize:     maxSize,
		list:        NewList(),
		buckets:     make([]*bucket, BUCKETS),
		persist:     make(chan persist),
		deletables:  make(chan *Entry, 1024),
		promotables: make(chan *Entry, 1024),
		newSize:     make(chan int),
		stop:        make(chan struct{}),
	}
	for i := 0; i < BUCKETS; i++ {
		c.buckets[i] = &bucket{lookup: make(map[string]map[string]*Entry)}
	}
	go c.worker()
	return c
}

func (c *Cache) Get(primary, secondary string) garnish.CachedResponse {
	bucket := c.bucket(primary)
	response := bucket.get(primary, secondary)
	if response == nil {
		return nil
	}
	c.promotables <- response
	return response
}

func (c *Cache) Set(primary string, secondary string, response garnish.CachedResponse) {
	entry := &Entry{
		Primary:        primary,
		Secondary:      secondary,
		CachedResponse: response,
		size:           response.Size(),
	}
	c.set(entry)
}

func (c *Cache) set(entry *Entry) {
	existing := c.bucket(entry.Primary).set(entry.Primary, entry.Secondary, entry)
	if existing != nil {
		c.deletables <- existing
	}
	c.promotables <- entry
}

func (c *Cache) Delete(primary string, secondary string) bool {
	return c.bucket(primary).delete(primary, secondary, c.deletables)
}

func (c *Cache) DeleteAll(primary string) bool {
	return c.bucket(primary).deleteAll(primary, c.deletables)
}

func (c *Cache) Save(path string, count int, cutoff time.Duration) error {
	p := persist{
		path:   path,
		count:  count,
		cutoff: cutoff,
		done:   make(chan error),
	}
	c.persist <- p
	return <-p.done
}

func (c *Cache) Load(path string) error {
	entries, err := loadFromFile(path)
	if err != nil {
		return err
	}
	expires := time.Now().Add(time.Second * 60)
	for _, entry := range entries {
		entry.Expire(expires.Add(time.Duration(rand.Intn(120)) * time.Second))
		c.set(entry)
	}
	return nil
}

func (c *Cache) Stop() {
	c.stop <- struct{}{}
}

func (c *Cache) SetSize(s int) {
	c.newSize <- s
}

func (c *Cache) GetSize() int {
	return c.maxSize
}

func (c *Cache) bucket(key string) *bucket {
	h := fnv.New32a()
	h.Write([]byte(key))
	return c.buckets[h.Sum32()&BUCKET_MASK]
}

func (c *Cache) worker() {
	for {
		select {
		case <-c.stop:
			return
		case s := <-c.newSize:
			c.maxSize = s
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
		case p := <-c.persist:
			i := 0
			cutoff := time.Now().Add(p.cutoff)
			entries := make([]*Entry, p.count)
			entry := c.list.head.next
			for i < p.count && entry != c.list.tail {
				if entry.Expires().After(cutoff) {
					entries[i] = entry
					i++
				}
				entry = entry.next
			}
			entries = entries[:i]
			go p.persist(entries)
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
		c.bucket(primary).delete(primary, entry.Secondary, nil)
		c.list.Remove(entry)
		c.size -= entry.size
	}
}
