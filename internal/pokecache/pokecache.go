// Package pokecache provides caching for the PokeAPI responses.
package pokecache

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// Cache is a concurrency-safe, in-memory store of PokeAPI response bodies
// keyed by request URL. Entries older than the interval given to New are
// removed by a background goroutine.
type Cache struct {
	m      map[string]cacheEntry
	mu     sync.RWMutex
	logger *slog.Logger
}

type cacheEntry struct {
	createdAt time.Time
	val       []byte
}

// New returns an empty Cache and starts a goroutine that removes entries older
// than interval, every interval. The goroutine runs until ctx is cancelled, so
// callers must cancel ctx when they are done with the returned Cache.
func New(ctx context.Context, logger *slog.Logger, interval time.Duration) *Cache {
	c := &Cache{
		m:      make(map[string]cacheEntry),
		logger: logger,
	}
	go c.reapLoop(ctx, interval)
	return c
}

// Add stores val under key, replacing any existing entry and resetting its age.
// The Cache keeps val rather than copying it, so callers must not modify val
// after the call.
func (c *Cache) Add(key string, val []byte) {
	c.mu.Lock()
	c.m[key] = cacheEntry{
		createdAt: time.Now(),
		val:       val,
	}
	c.mu.Unlock()

	c.logger.Debug("cache add", slog.String("key", key))
}

// Get returns the value stored under key and whether it was present. The
// returned slice is the stored one, not a copy, so callers must not modify it.
func (c *Cache) Get(key string) ([]byte, bool) {
	c.mu.RLock()
	entry, ok := c.m[key]
	c.mu.RUnlock()

	if !ok {
		c.logger.Debug("cache miss", slog.String("key", key))
		return nil, false
	}
	c.logger.Debug("cache hit", slog.String("key", key))
	return entry.val, true
}

// reapLoop deletes entries older than interval, once per interval, until ctx is
// cancelled.
func (c *Cache) reapLoop(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.reap(interval)
		}
	}
}

// reap deletes every entry older than interval.
func (c *Cache) reap(interval time.Duration) {
	var reaped []string

	c.mu.Lock()
	for key, entry := range c.m {
		if time.Since(entry.createdAt) > interval {
			delete(c.m, key)
			reaped = append(reaped, key)
		}
	}
	c.mu.Unlock()

	for _, key := range reaped {
		c.logger.Debug("cache reap", slog.String("key", key))
	}
}
