package apimock

import (
	"hash/fnv"
	"io"
	"os"
	"sync"
	"time"
)

// CacheEntry represents a cached parsed APIMock file.
type CacheEntry struct {
	File     *APIMockFile
	ModTime  time.Time
	FileHash uint64
	CachedAt time.Time
}

// ParserCache provides thread-safe caching of parsed APIMock files.
// It improves performance by avoiding re-parsing of unchanged files.
type ParserCache struct {
	mu      sync.RWMutex
	entries map[string]*CacheEntry
	config  *CacheConfig
	stats   *CacheStats
}

// CacheConfig configures the parser cache behavior.
type CacheConfig struct {
	// MaxSize is the maximum number of entries to cache (0 = unlimited)
	MaxSize int
	// TTL is time-to-live for cache entries (0 = no expiration)
	TTL time.Duration
	// CheckFileModTime validates file modification time before using cache
	CheckFileModTime bool
	// CheckFileHash validates file content hash before using cache (more expensive)
	CheckFileHash bool
}

// CacheStats tracks cache performance metrics.
type CacheStats struct {
	mu        sync.RWMutex
	Hits      int64
	Misses    int64
	Evictions int64
}

// DefaultCacheConfig returns a sensible default configuration.
func DefaultCacheConfig() *CacheConfig {
	return &CacheConfig{
		MaxSize:          100,
		TTL:              5 * time.Minute,
		CheckFileModTime: true,
		CheckFileHash:    false,
	}
}

// NewParserCache creates a new parser cache with the given configuration.
func NewParserCache(config *CacheConfig) *ParserCache {
	if config == nil {
		config = DefaultCacheConfig()
	}

	return &ParserCache{
		entries: make(map[string]*CacheEntry),
		config:  config,
		stats:   &CacheStats{},
	}
}

// Get retrieves a cached entry if valid, otherwise returns nil.
func (c *ParserCache) Get(filename string) *APIMockFile {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.entries[filename]
	if !exists {
		c.recordMiss()
		return nil
	}

	// Check TTL expiration
	if c.config.TTL > 0 && time.Since(entry.CachedAt) > c.config.TTL {
		c.recordMiss()
		return nil
	}

	c.recordHit()
	return entry.File
}

// Set stores a parsed file in the cache.
func (c *ParserCache) Set(filename string, file *APIMockFile, modTime time.Time, hash uint64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Enforce max size by evicting oldest entry
	if c.config.MaxSize > 0 && len(c.entries) >= c.config.MaxSize {
		c.evictOldest()
	}

	c.entries[filename] = &CacheEntry{
		File:     file,
		ModTime:  modTime,
		FileHash: hash,
		CachedAt: time.Now(),
	}
}

// Invalidate removes a specific file from the cache.
func (c *ParserCache) Invalidate(filename string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.entries, filename)
}

// Clear removes all entries from the cache.
func (c *ParserCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries = make(map[string]*CacheEntry)
}

// Size returns the current number of cached entries.
func (c *ParserCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.entries)
}

// Stats returns a copy of the cache statistics.
func (c *ParserCache) Stats() CacheStats {
	c.stats.mu.RLock()
	defer c.stats.mu.RUnlock()
	return CacheStats{
		Hits:      c.stats.Hits,
		Misses:    c.stats.Misses,
		Evictions: c.stats.Evictions,
	}
}

// HitRate returns the cache hit rate as a percentage (0-100).
func (c *ParserCache) HitRate() float64 {
	stats := c.Stats()
	total := stats.Hits + stats.Misses
	if total == 0 {
		return 0
	}
	return float64(stats.Hits) / float64(total) * 100
}

// evictOldest removes the oldest cache entry (must be called with lock held).
func (c *ParserCache) evictOldest() {
	var oldestKey string
	var oldestTime time.Time

	for key, entry := range c.entries {
		if oldestKey == "" || entry.CachedAt.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.CachedAt
		}
	}

	if oldestKey != "" {
		delete(c.entries, oldestKey)
		c.stats.mu.Lock()
		c.stats.Evictions++
		c.stats.mu.Unlock()
	}
}

// IsValid checks if a cached entry is still valid based on file metadata.
func (c *ParserCache) IsValid(filename string) bool {
	c.mu.RLock()
	entry, exists := c.entries[filename]
	c.mu.RUnlock()

	if !exists {
		return false
	}

	// Check TTL
	if c.config.TTL > 0 && time.Since(entry.CachedAt) > c.config.TTL {
		return false
	}

	// Check file modification time
	if c.config.CheckFileModTime {
		fileInfo, err := os.Stat(filename)
		if err != nil {
			return false
		}

		if !fileInfo.ModTime().Equal(entry.ModTime) {
			return false
		}
	}

	// Optionally check file hash for content changes
	if c.config.CheckFileHash {
		hash, err := hashFile(filename)
		if err != nil || hash != entry.FileHash {
			return false
		}
	}

	return true
}

// recordHit increments the hit counter.
func (c *ParserCache) recordHit() {
	c.stats.mu.Lock()
	c.stats.Hits++
	c.stats.mu.Unlock()
}

// recordMiss increments the miss counter.
func (c *ParserCache) recordMiss() {
	c.stats.mu.Lock()
	c.stats.Misses++
	c.stats.mu.Unlock()
}

// hashFile computes a hash of the file contents.
func hashFile(filename string) (uint64, error) {
	file, err := os.Open(filename)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	hash := fnv.New64a()
	if _, err := io.Copy(hash, file); err != nil {
		return 0, err
	}

	return hash.Sum64(), nil
}

// CachedParser wraps a Parser with caching capabilities.
type CachedParser struct {
	cache *ParserCache
}

// NewCachedParser creates a new parser with caching.
func NewCachedParser(config *CacheConfig) *CachedParser {
	return &CachedParser{
		cache: NewParserCache(config),
	}
}

// ParseFile parses a file with caching.
// It checks the cache first and only parses from disk if needed.
func (cp *CachedParser) ParseFile(filename string) (*APIMockFile, error) {
	// Check if cache entry is valid
	if !cp.cache.IsValid(filename) {
		cp.cache.Invalidate(filename)
	}

	// Try to get from cache
	if cached := cp.cache.Get(filename); cached != nil {
		return cached, nil
	}

	// Not in cache or invalid, parse the file
	fileInfo, err := os.Stat(filename)
	if err != nil {
		return nil, err
	}

	parser, err := NewParser(filename)
	if err != nil {
		return nil, err
	}

	file, err := parser.Parse()
	if err != nil {
		return nil, err
	}

	// Compute hash if needed
	var hash uint64
	if cp.cache.config.CheckFileHash {
		hash, _ = hashFile(filename)
	}

	// Store in cache
	cp.cache.Set(filename, file, fileInfo.ModTime(), hash)

	return file, nil
}

// Cache returns the underlying cache for direct access.
func (cp *CachedParser) Cache() *ParserCache {
	return cp.cache
}

// InvalidateAll clears all cached entries.
func (cp *CachedParser) InvalidateAll() {
	cp.cache.Clear()
}

// Stats returns cache statistics.
func (cp *CachedParser) Stats() CacheStats {
	return cp.cache.Stats()
}

// HitRate returns the cache hit rate as a percentage.
func (cp *CachedParser) HitRate() float64 {
	return cp.cache.HitRate()
}
