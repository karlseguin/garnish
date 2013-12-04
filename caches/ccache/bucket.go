package ccache

import (
	"github.com/karlseguin/garnish/caches"
	"sync"
)

type Bucket struct {
	sync.RWMutex
	lookup map[string]*Vary
}

type Vary struct {
	sync.RWMutex
	lookup map[string]*Item
}

func (b *Bucket) get(key, vary string) *Item {
	v, ok := b.getVary(key)
	if ok == false {
		return nil
	}
	v.RLock()
	defer v.RUnlock()
	return v.lookup[vary]
}

func (b *Bucket) set(key, vary string, value *caches.CachedResponse) (*Item, bool) {
	v, ok := b.getVary(key)
	if ok == false {
		b.Lock()
		if v, ok = b.lookup[key]; ok == false {
			v = &Vary{lookup: make(map[string]*Item)}
			b.lookup[key] = v
		}
		b.Unlock()
	}
	v.Lock()
	defer v.Unlock()
	if existing, exists := v.lookup[vary]; exists {
		existing.Lock()
		existing.value = value
		existing.Unlock()
		return existing, false
	}
	item := newItem(key, vary, value)
	v.lookup[vary] = item
	return item, true
}

func (b *Bucket) delete(key string) {
	b.Lock()
	defer b.Unlock()
	delete(b.lookup, key)
}

func (b *Bucket) deleteVary(key, vary string) {
	v, ok := b.getVary(key)
	if ok == false {
		return
	}
	v.Lock()
	defer v.Unlock()
	delete(v.lookup, vary)
}

func (b *Bucket) getAndDelete(key string) *Vary {
	b.Lock()
	defer b.Unlock()
	vary := b.lookup[key]
	delete(b.lookup, key)
	return vary
}

func (b *Bucket) getAndDeleteVary(key, vary string) *Item {
	v, ok := b.getVary(key)
	if ok == false {
		return nil
	}
	v.Lock()
	defer v.Unlock()
	item := v.lookup[vary]
	delete(v.lookup, vary)
	return item
}

func (b *Bucket) clear() {
	b.Lock()
	defer b.Unlock()
	b.lookup = make(map[string]*Vary)
}

func (b *Bucket) getVary(key string) (*Vary, bool) {
	b.RLock()
	defer b.RUnlock()
	v, exists := b.lookup[key]
	return v, exists
}
