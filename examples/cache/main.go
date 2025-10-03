// Package main demonstrates how to use the APIMock parser with caching
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/pretodev/anansi-proxy/pkg/apimock"
)

func main() {
	// Example 1: Basic usage with default cache configuration
	fmt.Println("=== Example 1: Basic Cached Parser ===")
	basicExample()

	fmt.Println()

	// Example 2: Custom cache configuration
	fmt.Println("=== Example 2: Custom Cache Configuration ===")
	customConfigExample()

	fmt.Println()

	// Example 3: Cache statistics and monitoring
	fmt.Println("=== Example 3: Cache Statistics ===")
	statsExample()

	fmt.Println()

	// Example 4: Manual cache management
	fmt.Println("=== Example 4: Manual Cache Management ===")
	manualCacheExample()
}

func basicExample() {
	// Create a cached parser with default settings
	// - Max 100 entries
	// - 5 minute TTL
	// - File modification time checking enabled
	parser := apimock.NewCachedParser(apimock.DefaultCacheConfig())

	// Parse a file - first time will read from disk
	file, err := parser.ParseFile("../../docs/apimock/examples/json.apimock")
	if err != nil {
		log.Printf("Parse error: %v", err)
		return
	}

	fmt.Printf("Parsed: %s %s\n", file.Request.Method, file.Request.Path)
	fmt.Printf("Responses: %d\n", len(file.Responses))

	// Parse again - will use cache (much faster!)
	start := time.Now()
	file2, err := parser.ParseFile("../../docs/apimock/examples/json.apimock")
	if err != nil {
		log.Printf("Parse error: %v", err)
		return
	}
	elapsed := time.Since(start)

	fmt.Printf("Second parse took: %v (from cache)\n", elapsed)
	fmt.Printf("Same instance: %v\n", file == file2)
}

func customConfigExample() {
	// Create custom cache configuration
	config := &apimock.CacheConfig{
		MaxSize:          50,               // Cache up to 50 files
		TTL:              10 * time.Minute, // Expire after 10 minutes
		CheckFileModTime: true,             // Check file modification time
		CheckFileHash:    true,             // Also check file content hash (slower but safer)
	}

	parser := apimock.NewCachedParser(config)

	file, err := parser.ParseFile("../../docs/apimock/examples/simple.apimock")
	if err != nil {
		log.Printf("Parse error: %v", err)
		return
	}

	if file.Request != nil {
		fmt.Printf("Parsed with custom config: %s %s\n", file.Request.Method, file.Request.Path)
	} else {
		fmt.Println("Parsed with custom config: (no request section)")
	}
	fmt.Printf("Cache size: %d\n", parser.Cache().Size())
}

func statsExample() {
	parser := apimock.NewCachedParser(apimock.DefaultCacheConfig())

	// Parse multiple files
	files := []string{
		"../../docs/apimock/examples/json.apimock",
		"../../docs/apimock/examples/simple.apimock",
		"../../docs/apimock/examples/xml.apimock",
	}

	for _, filename := range files {
		// Parse twice to generate cache hits
		parser.ParseFile(filename)
		parser.ParseFile(filename)
	}

	// Get statistics
	stats := parser.Stats()
	fmt.Printf("Cache Hits: %d\n", stats.Hits)
	fmt.Printf("Cache Misses: %d\n", stats.Misses)
	fmt.Printf("Cache Evictions: %d\n", stats.Evictions)
	fmt.Printf("Hit Rate: %.2f%%\n", parser.HitRate())
}

func manualCacheExample() {
	parser := apimock.NewCachedParser(apimock.DefaultCacheConfig())

	// Parse and cache a file
	file, err := parser.ParseFile("../../docs/apimock/examples/json.apimock")
	if err != nil {
		log.Printf("Parse error: %v", err)
		return
	}

	fmt.Printf("Initial cache size: %d\n", parser.Cache().Size())

	// Manually invalidate a specific file
	parser.Cache().Invalidate("../../docs/apimock/examples/json.apimock")
	fmt.Printf("After invalidation: %d\n", parser.Cache().Size())

	// Re-parse (will be a cache miss)
	file2, _ := parser.ParseFile("../../docs/apimock/examples/json.apimock")
	fmt.Printf("After re-parse: %d\n", parser.Cache().Size())
	fmt.Printf("Different instance: %v\n", file != file2)

	// Clear entire cache
	parser.InvalidateAll()
	fmt.Printf("After clearing all: %d\n", parser.Cache().Size())
}
