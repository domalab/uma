package daemon

import (
	"testing"
	"time"

	"github.com/domalab/uma/daemon/services/async"
	"github.com/domalab/uma/daemon/services/cache"
)

// TestCacheServiceIntegration tests cache service functionality
func TestCacheServiceIntegration(t *testing.T) {
	// Initialize cache manager
	cacheManager := cache.NewCacheManager()
	testCache := cacheManager.GetCache("test_cache", 5*time.Minute, 100)

	// Test basic cache operations
	testKey := "test_key"
	testValue := map[string]interface{}{
		"data":      "test_data",
		"timestamp": time.Now(),
	}

	// Set value in cache
	testCache.SetWithTTL(testKey, testValue, 1*time.Minute)

	// Retrieve value from cache
	cachedValue, found := testCache.Get(testKey)
	if !found {
		t.Error("Expected to find cached value")
	}

	if cachedValue == nil {
		t.Error("Expected non-nil cached value")
	}

	// Test cache expiration
	testCache.SetWithTTL("expire_test", "expire_value", 100*time.Millisecond)
	time.Sleep(150 * time.Millisecond)

	_, found = testCache.Get("expire_test")
	if found {
		t.Error("Expected expired value to not be found")
	}

	// Test cache stats
	stats := testCache.GetStats()
	if totalEntries, ok := stats["total_entries"].(int); ok && totalEntries < 0 {
		t.Error("Expected non-negative total entries")
	}

	t.Logf("Cache stats: %+v", stats)
}

// TestAsyncManagerBasicOperations tests basic async manager functionality
func TestAsyncManagerBasicOperations(t *testing.T) {
	// Initialize async manager
	asyncManager := async.NewAsyncManager()
	defer asyncManager.Stop()

	// Test listing operations when empty
	response := asyncManager.ListOperations("", "")
	if response.Total != 0 {
		t.Errorf("Expected 0 operations initially, got %d", response.Total)
	}

	// Test getting stats
	stats := asyncManager.GetStats()
	if totalOps, ok := stats["total_operations"].(int); ok && totalOps != 0 {
		t.Errorf("Expected 0 total operations initially, got %d", totalOps)
	}

	// Test getting non-existent operation
	_, err := asyncManager.GetOperation("non-existent-id")
	if err == nil {
		t.Error("Expected error when getting non-existent operation")
	}

	t.Logf("Async manager stats: %+v", stats)
}
