package cache

import (
	"sync"
)

type bucket struct {
	sync.RWMutex
	lookup map[string]map[string]*Entry
}

func (b *bucket) get(primary string, secondary string) *Entry {
	defer b.RUnlock()
	b.RLock()
	group, ok := b.lookup[primary]
	if ok == false {
		return nil
	}
	return group[secondary]
}

func (b *bucket) set(primary string, secondary string, entry *Entry) *Entry {
	defer b.Unlock()
	b.Lock()

	if group, ok := b.lookup[primary]; ok {
		existing := group[secondary]
		group[secondary] = entry
		return existing
	}

	b.lookup[primary] = map[string]*Entry{secondary: entry}
	return nil
}

func (b *bucket) delete(primary string, secondary string) bool {
	defer b.Unlock()
	b.Lock()
	if group, ok := b.lookup[primary]; ok {
		if _, ok := group[secondary]; ok {
			delete(group, secondary)
			return true
		}
	}
	return false
}

func (b *bucket) deleteAll(primary string) bool {
	defer b.Unlock()
	b.Lock()
	if _, ok := b.lookup[primary]; ok {
		delete(b.lookup, primary)
		return true
	}
	return false
}
