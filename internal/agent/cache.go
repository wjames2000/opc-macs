package agent

import (
	"crypto/sha256"
	"fmt"
	"sync"
	"time"

	"github.com/wjames2000/opc-macs/internal/runtime"
)

type CacheEntry struct {
	Result    *runtime.ExecutionResult
	CreatedAt time.Time
	HitCount  int
}

type ResultCache struct {
	mu      sync.RWMutex
	entries map[string]*CacheEntry
	maxSize int
	ttl     time.Duration
}

func NewResultCache(maxSize int, ttl time.Duration) *ResultCache {
	return &ResultCache{
		entries: make(map[string]*CacheEntry),
		maxSize: maxSize,
		ttl:     ttl,
	}
}

func (c *ResultCache) key(agentName, input string) string {
	h := sha256.Sum256([]byte(agentName + "|" + input))
	return fmt.Sprintf("%x", h[:16])
}

func (c *ResultCache) Get(agentName, input string) *runtime.ExecutionResult {
	c.mu.RLock()
	defer c.mu.RUnlock()
	entry, ok := c.entries[c.key(agentName, input)]
	if !ok {
		return nil
	}
	if time.Since(entry.CreatedAt) > c.ttl {
		return nil
	}
	entry.HitCount++
	return entry.Result
}

func (c *ResultCache) Set(agentName, input string, result *runtime.ExecutionResult) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.entries) >= c.maxSize {
		var oldestKey string
		var oldestTime time.Time
		for k, v := range c.entries {
			if oldestKey == "" || v.CreatedAt.Before(oldestTime) {
				oldestKey = k
				oldestTime = v.CreatedAt
			}
		}
		delete(c.entries, oldestKey)
	}
	c.entries[c.key(agentName, input)] = &CacheEntry{
		Result: result, CreatedAt: time.Now(), HitCount: 0,
	}
}

func (c *ResultCache) Stats() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	var totalHits int
	for _, e := range c.entries {
		totalHits += e.HitCount
	}
	return fmt.Sprintf("缓存：%d 条条目，%d 次命中，TTL=%v", len(c.entries), totalHits, c.ttl)
}

func (c *ResultCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries = make(map[string]*CacheEntry)
}
