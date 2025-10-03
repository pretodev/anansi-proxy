package apimock

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestParserCache_GetSet(t *testing.T) {
	cache := NewParserCache(DefaultCacheConfig())

	file := &APIMockFile{
		Request: &RequestSection{
			Method: "GET",
			Path:   "/test",
		},
	}

	// Cache should be empty initially
	if got := cache.Get("test.apimock"); got != nil {
		t.Errorf("Expected nil for non-existent entry, got %v", got)
	}

	// Set and retrieve
	modTime := time.Now()
	cache.Set("test.apimock", file, modTime, 12345)

	if got := cache.Get("test.apimock"); got == nil {
		t.Error("Expected cached entry, got nil")
	} else if got.Request.Method != "GET" {
		t.Errorf("Expected method GET, got %s", got.Request.Method)
	}
}

func TestParserCache_TTLExpiration(t *testing.T) {
	config := &CacheConfig{
		MaxSize:          10,
		TTL:              100 * time.Millisecond,
		CheckFileModTime: false,
	}
	cache := NewParserCache(config)

	file := &APIMockFile{
		Request: &RequestSection{Method: "GET", Path: "/test"},
	}

	// Set entry
	cache.Set("test.apimock", file, time.Now(), 0)

	// Should be valid immediately
	if got := cache.Get("test.apimock"); got == nil {
		t.Error("Expected valid entry immediately after set")
	}

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Should be expired
	if got := cache.Get("test.apimock"); got != nil {
		t.Error("Expected nil for expired entry")
	}
}

func TestParserCache_MaxSize(t *testing.T) {
	config := &CacheConfig{
		MaxSize:          3,
		TTL:              0,
		CheckFileModTime: false,
	}
	cache := NewParserCache(config)

	file := &APIMockFile{
		Request: &RequestSection{Method: "GET", Path: "/test"},
	}

	// Add 4 entries (max is 3)
	for i := 0; i < 4; i++ {
		time.Sleep(10 * time.Millisecond) // Ensure different timestamps
		filename := filepath.Join("test", string(rune('a'+i))+".apimock")
		cache.Set(filename, file, time.Now(), 0)
	}

	// Cache should have exactly 3 entries
	if size := cache.Size(); size != 3 {
		t.Errorf("Expected size 3, got %d", size)
	}

	// Check eviction stat
	stats := cache.Stats()
	if stats.Evictions != 1 {
		t.Errorf("Expected 1 eviction, got %d", stats.Evictions)
	}
}

func TestParserCache_Invalidate(t *testing.T) {
	cache := NewParserCache(DefaultCacheConfig())

	file := &APIMockFile{
		Request: &RequestSection{Method: "GET", Path: "/test"},
	}

	cache.Set("test.apimock", file, time.Now(), 0)

	if cache.Size() != 1 {
		t.Errorf("Expected size 1, got %d", cache.Size())
	}

	cache.Invalidate("test.apimock")

	if cache.Size() != 0 {
		t.Errorf("Expected size 0 after invalidation, got %d", cache.Size())
	}

	if got := cache.Get("test.apimock"); got != nil {
		t.Error("Expected nil after invalidation")
	}
}

func TestParserCache_Clear(t *testing.T) {
	cache := NewParserCache(DefaultCacheConfig())

	file := &APIMockFile{
		Request: &RequestSection{Method: "GET", Path: "/test"},
	}

	cache.Set("test1.apimock", file, time.Now(), 0)
	cache.Set("test2.apimock", file, time.Now(), 0)
	cache.Set("test3.apimock", file, time.Now(), 0)

	if cache.Size() != 3 {
		t.Errorf("Expected size 3, got %d", cache.Size())
	}

	cache.Clear()

	if cache.Size() != 0 {
		t.Errorf("Expected size 0 after clear, got %d", cache.Size())
	}
}

func TestParserCache_Stats(t *testing.T) {
	cache := NewParserCache(DefaultCacheConfig())

	file := &APIMockFile{
		Request: &RequestSection{Method: "GET", Path: "/test"},
	}

	// Initial stats should be zero
	stats := cache.Stats()
	if stats.Hits != 0 || stats.Misses != 0 {
		t.Error("Expected zero initial stats")
	}

	// Miss on non-existent entry
	cache.Get("test.apimock")
	stats = cache.Stats()
	if stats.Misses != 1 {
		t.Errorf("Expected 1 miss, got %d", stats.Misses)
	}

	// Set and hit
	cache.Set("test.apimock", file, time.Now(), 0)
	cache.Get("test.apimock")
	stats = cache.Stats()
	if stats.Hits != 1 {
		t.Errorf("Expected 1 hit, got %d", stats.Hits)
	}

	// Test hit rate
	hitRate := cache.HitRate()
	expectedRate := 50.0 // 1 hit, 1 miss = 50%
	if hitRate != expectedRate {
		t.Errorf("Expected hit rate %.2f%%, got %.2f%%", expectedRate, hitRate)
	}
}

func TestParserCache_IsValid(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.apimock")

	content := `GET /api/test

-- 200: OK

{}`

	if err := os.WriteFile(testFile, []byte(content), 0o644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	config := &CacheConfig{
		MaxSize:          10,
		TTL:              0,
		CheckFileModTime: true,
	}
	cache := NewParserCache(config)

	file := &APIMockFile{
		Request: &RequestSection{Method: "GET", Path: "/api/test"},
	}

	info, _ := os.Stat(testFile)
	cache.Set(testFile, file, info.ModTime(), 0)

	// Should be valid with same mod time
	if !cache.IsValid(testFile) {
		t.Error("Expected cache to be valid")
	}

	// Modify file
	time.Sleep(10 * time.Millisecond)
	if err := os.WriteFile(testFile, []byte(content+"modified"), 0o644); err != nil {
		t.Fatalf("Failed to modify test file: %v", err)
	}

	// Should be invalid after modification
	if cache.IsValid(testFile) {
		t.Error("Expected cache to be invalid after file modification")
	}
}

func TestCachedParser_ParseFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.apimock")

	content := `GET /api/users

-- 200: OK
Content-Type: application/json

{"users": []}`

	if err := os.WriteFile(testFile, []byte(content), 0o644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	config := DefaultCacheConfig()
	config.CheckFileModTime = true
	parser := NewCachedParser(config)

	// First parse - should hit file system
	file1, err := parser.ParseFile(testFile)
	if err != nil {
		t.Fatalf("Failed to parse file: %v", err)
	}

	if file1.Request.Method != "GET" {
		t.Errorf("Expected method GET, got %s", file1.Request.Method)
	}

	stats1 := parser.Stats()
	if stats1.Misses != 1 {
		t.Errorf("Expected 1 miss after first parse, got %d", stats1.Misses)
	}

	// Second parse - should hit cache
	file2, err := parser.ParseFile(testFile)
	if err != nil {
		t.Fatalf("Failed to parse file from cache: %v", err)
	}

	// Should be the same instance
	if file1 != file2 {
		t.Error("Expected same instance from cache")
	}

	stats2 := parser.Stats()
	if stats2.Hits != 1 {
		t.Errorf("Expected 1 hit after second parse, got %d", stats2.Hits)
	}

	if parser.Cache().Size() != 1 {
		t.Errorf("Expected cache size 1, got %d", parser.Cache().Size())
	}
}

func TestCachedParser_FileModification(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.apimock")

	content1 := `GET /api/users

-- 200: OK

{"users": []}`

	if err := os.WriteFile(testFile, []byte(content1), 0o644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	config := DefaultCacheConfig()
	config.CheckFileModTime = true
	parser := NewCachedParser(config)

	// Parse original file
	file1, err := parser.ParseFile(testFile)
	if err != nil {
		t.Fatalf("Failed to parse file: %v", err)
	}

	if len(file1.Responses) == 0 {
		t.Fatal("Expected at least one response")
	}

	if file1.Responses[0].StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", file1.Responses[0].StatusCode)
	}

	// Wait a bit to ensure different mod time
	time.Sleep(10 * time.Millisecond)

	// Modify file
	content2 := `GET /api/users

-- 404: Not Found

{"error": "not found"}`

	if err := os.WriteFile(testFile, []byte(content2), 0o644); err != nil {
		t.Fatalf("Failed to modify test file: %v", err)
	}

	// Parse again - should detect modification and re-parse
	file2, err := parser.ParseFile(testFile)
	if err != nil {
		t.Fatalf("Failed to parse modified file: %v", err)
	}

	if len(file2.Responses) == 0 {
		t.Fatal("Expected at least one response")
	}

	if file2.Responses[0].StatusCode != 404 {
		t.Errorf("Expected status 404 after modification, got %d", file2.Responses[0].StatusCode)
	}

	// Should have 2 misses (original + re-parse after modification)
	stats := parser.Stats()
	if stats.Misses < 2 {
		t.Errorf("Expected at least 2 misses, got %d", stats.Misses)
	}
}

func TestCachedParser_InvalidateAll(t *testing.T) {
	tmpDir := t.TempDir()

	config := DefaultCacheConfig()
	parser := NewCachedParser(config)

	// Parse multiple files
	for i := 0; i < 3; i++ {
		testFile := filepath.Join(tmpDir, filepath.Join("test", string(rune('a'+i))+".apimock"))
		if err := os.MkdirAll(filepath.Dir(testFile), 0o755); err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}

		content := `GET /api/test

-- 200: OK

{}`
		if err := os.WriteFile(testFile, []byte(content), 0o644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		if _, err := parser.ParseFile(testFile); err != nil {
			t.Fatalf("Failed to parse file: %v", err)
		}
	}

	if parser.Cache().Size() != 3 {
		t.Errorf("Expected cache size 3, got %d", parser.Cache().Size())
	}

	parser.InvalidateAll()

	if parser.Cache().Size() != 0 {
		t.Errorf("Expected cache size 0 after invalidation, got %d", parser.Cache().Size())
	}
}

func TestHashFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	content := "test content"
	if err := os.WriteFile(testFile, []byte(content), 0o644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	hash1, err := hashFile(testFile)
	if err != nil {
		t.Fatalf("Failed to hash file: %v", err)
	}

	// Same content should produce same hash
	hash2, err := hashFile(testFile)
	if err != nil {
		t.Fatalf("Failed to hash file: %v", err)
	}

	if hash1 != hash2 {
		t.Error("Expected same hash for same content")
	}

	// Different content should produce different hash
	if err := os.WriteFile(testFile, []byte("different content"), 0o644); err != nil {
		t.Fatalf("Failed to modify test file: %v", err)
	}

	hash3, err := hashFile(testFile)
	if err != nil {
		t.Fatalf("Failed to hash modified file: %v", err)
	}

	if hash1 == hash3 {
		t.Error("Expected different hash for different content")
	}
}

func TestHashFile_NonExistent(t *testing.T) {
	_, err := hashFile("/non/existent/file.txt")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func BenchmarkCachedParser_WithCache(b *testing.B) {
	tmpDir := b.TempDir()
	testFile := filepath.Join(tmpDir, "test.apimock")

	content := `POST /api/users/{id}
Content-Type: application/json

{"name": "John Doe"}

-- 200: Success
Content-Type: application/json

{"id": 123, "name": "John Doe"}`

	if err := os.WriteFile(testFile, []byte(content), 0o644); err != nil {
		b.Fatalf("Failed to create test file: %v", err)
	}

	parser := NewCachedParser(DefaultCacheConfig())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := parser.ParseFile(testFile)
		if err != nil {
			b.Fatalf("Parse error: %v", err)
		}
	}
}

func BenchmarkCachedParser_WithoutCache(b *testing.B) {
	tmpDir := b.TempDir()
	testFile := filepath.Join(tmpDir, "test.apimock")

	content := `POST /api/users/{id}
Content-Type: application/json

{"name": "John Doe"}

-- 200: Success
Content-Type: application/json

{"id": 123, "name": "John Doe"}`

	if err := os.WriteFile(testFile, []byte(content), 0o644); err != nil {
		b.Fatalf("Failed to create test file: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parser, err := NewParser(testFile)
		if err != nil {
			b.Fatalf("Parser creation error: %v", err)
		}
		_, err = parser.Parse()
		if err != nil {
			b.Fatalf("Parse error: %v", err)
		}
	}
}
