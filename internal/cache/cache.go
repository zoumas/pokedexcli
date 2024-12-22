package cache

import (
	"sync"
	"time"
)

type Cache struct {
	m  map[string]entry
	mu sync.RWMutex
}

type entry struct {
	creation time.Time
	value    []byte
}

func NewCache(interval time.Duration) *Cache {
	c := &Cache{
		m: make(map[string]entry),
	}
	go func(c *Cache) {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			c.reap(interval)
		}
	}(c)
	return c
}

func (c *Cache) Get(key string) (value []byte, ok bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	entry, ok := c.m[key]
	return entry.value, ok
}

func (c *Cache) Add(key string, value []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.m[key] = entry{
		creation: time.Now(),
		value:    value,
	}
}

func (c *Cache) reap(interval time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for k, v := range c.m {
		if time.Since(v.creation) > interval {
			delete(c.m, k)
		}
	}
}
