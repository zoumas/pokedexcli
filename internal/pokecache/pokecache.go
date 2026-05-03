// Package pokecache provides a generic in-memory cache with automatic expiration of entries.
// It is designed to be used in the Pokedex application to store API responses and reduce the number of requests made to the PokeAPI.
// The cache uses a simple map to store entries, and a background goroutine to periodically check for and remove expired entries.
// Each entry is timestamped when it is added, and the expiration time is determined by the reap interval specified when creating the cache.
package pokecache

import (
	"log"
	"sync"
	"time"
)

type Cache[T any] struct {
	m  map[string]Entry[T]
	mu sync.RWMutex
}

type Entry[T any] struct {
	value     T
	createdAt time.Time
}

func NewCache[T any](reapInterval time.Duration) *Cache[T] {
	cache := &Cache[T]{
		m: make(map[string]Entry[T]),
	}
	go cache.reapLoop(reapInterval)
	return cache
}

func (c *Cache[T]) Add(key string, value T) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.m[key] = Entry[T]{
		value:     value,
		createdAt: time.Now(),
	}
}

func (c *Cache[T]) Get(key string) (T, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, ok := c.m[key]
	return entry.value, ok
}

func (c *Cache[T]) reapLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()

		for key, entry := range c.m {
			if time.Since(entry.createdAt) > interval {
				log.Printf("Reaping cache entry for key: %s", key)
				delete(c.m, key)
			}
		}

		c.mu.Unlock()
	}
}
