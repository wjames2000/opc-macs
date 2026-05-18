package media

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sort"
	"sync"
	"time"
)

const (
	// DefaultCacheTTL is the default time-to-live for cached entries (1 hour).
	DefaultCacheTTL = 1 * time.Hour
	// DefaultCacheMaxEntries is the maximum number of entries in the cache.
	DefaultCacheMaxEntries = 1000
)

// CacheEntry holds a cached image generation result.
type CacheEntry struct {
	Data      []byte    `json:"data"`
	Prompt    string    `json:"prompt"`
	OptsHash  string    `json:"opts_hash"`
	CreatedAt time.Time `json:"created_at"`
	HitCount  int       `json:"hit_count"`
}

// ImageCache provides caching and deduplication for image generation requests.
// Identical prompts + options combinations reuse cached results to avoid
// redundant API calls and costs.
type ImageCache struct {
	mu         sync.RWMutex
	entries    map[string]*CacheEntry
	ttl        time.Duration
	maxEntries int
}

// NewImageCache creates a new image cache with default settings.
func NewImageCache() *ImageCache {
	return &ImageCache{
		entries:    make(map[string]*CacheEntry),
		ttl:        DefaultCacheTTL,
		maxEntries: DefaultCacheMaxEntries,
	}
}

// NewImageCacheWithConfig creates a cache with custom TTL and max entries.
func NewImageCacheWithConfig(ttl time.Duration, maxEntries int) *ImageCache {
	return &ImageCache{
		entries:    make(map[string]*CacheEntry),
		ttl:        ttl,
		maxEntries: maxEntries,
	}
}

// cacheKey generates a deterministic hash from prompt and options.
func (c *ImageCache) cacheKey(prompt string, opts map[string]interface{}) string {
	h := sha256.New()
	h.Write([]byte(prompt))
	if opts != nil {
		keys := make([]string, 0, len(opts))
		for k := range opts {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			val, _ := json.Marshal(opts[k])
			h.Write([]byte(k))
			h.Write(val)
		}
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}

// Get retrieves a cached result. Returns nil if not found or expired.
func (c *ImageCache) Get(prompt string, opts map[string]interface{}) []byte {
	key := c.cacheKey(prompt, opts)

	c.mu.RLock()
	entry, ok := c.entries[key]
	if !ok {
		c.mu.RUnlock()
		return nil
	}
	expired := time.Since(entry.CreatedAt) > c.ttl
	if expired {
		c.mu.RUnlock()
		c.mu.Lock()
		delete(c.entries, key)
		c.mu.Unlock()
		return nil
	}
	entry.HitCount++
	data := entry.Data
	c.mu.RUnlock()
	return data
}

// Set stores a result in the cache.
func (c *ImageCache) Set(prompt string, opts map[string]interface{}, data []byte) {
	key := c.cacheKey(prompt, opts)

	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.entries) >= c.maxEntries {
		c.evictOldest()
	}

	c.entries[key] = &CacheEntry{
		Data:      data,
		Prompt:    prompt,
		OptsHash:  key,
		CreatedAt: time.Now(),
		HitCount:  0,
	}
}

// GetOrGenerate checks the cache first; if miss, calls generator and caches result.
func (c *ImageCache) GetOrGenerate(ctx context.Context, prompt string, opts map[string]interface{}, generator ImageGenerator) ([]byte, error) {
	if cached := c.Get(prompt, opts); cached != nil {
		return cached, nil
	}

	data, err := generator.Generate(ctx, prompt, opts)
	if err != nil {
		return nil, err
	}

	c.Set(prompt, opts, data)
	return data, nil
}

// Stats returns cache usage statistics.
type CacheStats struct {
	Entries    int `json:"entries"`
	MaxEntries int `json:"max_entries"`
	TotalHits  int `json:"total_hits"`
}

// Stats returns current cache statistics.
func (c *ImageCache) Stats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	totalHits := 0
	for _, entry := range c.entries {
		totalHits += entry.HitCount
	}
	return CacheStats{
		Entries:    len(c.entries),
		MaxEntries: c.maxEntries,
		TotalHits:  totalHits,
	}
}

// Clear removes all cached entries.
func (c *ImageCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries = make(map[string]*CacheEntry)
}

// evictOldest removes the oldest 10% of entries when cache is full.
func (c *ImageCache) evictOldest() {
	if len(c.entries) < c.maxEntries {
		return
	}

	type kv struct {
		key string
		t   time.Time
	}
	var sorted []kv
	for k, v := range c.entries {
		sorted = append(sorted, kv{k, v.CreatedAt})
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].t.Before(sorted[j].t)
	})

	evictCount := len(sorted) / 10
	if evictCount < 1 {
		evictCount = 1
	}
	for i := 0; i < evictCount && i < len(sorted); i++ {
		delete(c.entries, sorted[i].key)
	}
}
