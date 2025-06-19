package cache

import (
	"testing"
	"time"
)

func TestCache_SetAndGet(t *testing.T) {
	cache := NewCache(1*time.Minute, 10)
	defer cache.Stop()
	
	// Test setting and getting a value
	key := "test-key"
	value := "test-value"
	
	cache.Set(key, value)
	
	retrieved, found := cache.Get(key)
	if !found {
		t.Error("Expected to find cached value")
	}
	
	if retrieved != value {
		t.Errorf("Expected %s, got %s", value, retrieved)
	}
}

func TestCache_Expiration(t *testing.T) {
	cache := NewCache(100*time.Millisecond, 10)
	defer cache.Stop()
	
	key := "test-key"
	value := "test-value"
	
	cache.Set(key, value)
	
	// Should be available immediately
	_, found := cache.Get(key)
	if !found {
		t.Error("Expected to find cached value immediately")
	}
	
	// Wait for expiration
	time.Sleep(150 * time.Millisecond)
	
	// Should be expired
	_, found = cache.Get(key)
	if found {
		t.Error("Expected cached value to be expired")
	}
}

func TestCache_CustomTTL(t *testing.T) {
	cache := NewCache(1*time.Minute, 10)
	defer cache.Stop()
	
	key := "test-key"
	value := "test-value"
	customTTL := 50 * time.Millisecond
	
	cache.SetWithTTL(key, value, customTTL)
	
	// Should be available immediately
	_, found := cache.Get(key)
	if !found {
		t.Error("Expected to find cached value immediately")
	}
	
	// Wait for custom TTL expiration
	time.Sleep(75 * time.Millisecond)
	
	// Should be expired
	_, found = cache.Get(key)
	if found {
		t.Error("Expected cached value to be expired after custom TTL")
	}
}

func TestCache_Delete(t *testing.T) {
	cache := NewCache(1*time.Minute, 10)
	defer cache.Stop()
	
	key := "test-key"
	value := "test-value"
	
	cache.Set(key, value)
	
	// Verify it's there
	_, found := cache.Get(key)
	if !found {
		t.Error("Expected to find cached value")
	}
	
	// Delete it
	cache.Delete(key)
	
	// Should be gone
	_, found = cache.Get(key)
	if found {
		t.Error("Expected cached value to be deleted")
	}
}

func TestCache_Clear(t *testing.T) {
	cache := NewCache(1*time.Minute, 10)
	defer cache.Stop()
	
	// Add multiple entries
	for i := 0; i < 5; i++ {
		cache.Set(string(rune('a'+i)), i)
	}
	
	// Verify they're there
	stats := cache.GetStats()
	if stats["entries"].(int) != 5 {
		t.Errorf("Expected 5 entries, got %d", stats["entries"].(int))
	}
	
	// Clear cache
	cache.Clear()
	
	// Should be empty
	stats = cache.GetStats()
	if stats["entries"].(int) != 0 {
		t.Errorf("Expected 0 entries after clear, got %d", stats["entries"].(int))
	}
}

func TestCache_LRUEviction(t *testing.T) {
	cache := NewCache(1*time.Minute, 3) // Small cache for testing eviction
	defer cache.Stop()
	
	// Fill cache to capacity
	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	cache.Set("key3", "value3")
	
	// Access key1 to make it recently used
	cache.Get("key1")
	
	// Add another entry, should evict key2 (least recently used)
	cache.Set("key4", "value4")
	
	// key1 should still be there
	_, found := cache.Get("key1")
	if !found {
		t.Error("Expected key1 to still be in cache (recently accessed)")
	}
	
	// key2 should be evicted
	_, found = cache.Get("key2")
	if found {
		t.Error("Expected key2 to be evicted (least recently used)")
	}
	
	// key3 and key4 should be there
	_, found = cache.Get("key3")
	if !found {
		t.Error("Expected key3 to still be in cache")
	}
	
	_, found = cache.Get("key4")
	if !found {
		t.Error("Expected key4 to be in cache (just added)")
	}
}

func TestCache_Stats(t *testing.T) {
	cache := NewCache(1*time.Minute, 10)
	defer cache.Stop()
	
	// Add some entries
	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	
	// Generate some hits and misses
	cache.Get("key1") // hit
	cache.Get("key1") // hit
	cache.Get("key3") // miss
	cache.Get("key2") // hit
	cache.Get("key4") // miss
	
	stats := cache.GetStats()
	
	if stats["entries"].(int) != 2 {
		t.Errorf("Expected 2 entries, got %d", stats["entries"].(int))
	}
	
	if stats["hits"].(int64) != 3 {
		t.Errorf("Expected 3 hits, got %d", stats["hits"].(int64))
	}
	
	if stats["misses"].(int64) != 2 {
		t.Errorf("Expected 2 misses, got %d", stats["misses"].(int64))
	}
	
	expectedHitRate := float64(3) / float64(5) * 100 // 60%
	if stats["hit_rate"].(float64) != expectedHitRate {
		t.Errorf("Expected hit rate %.2f, got %.2f", expectedHitRate, stats["hit_rate"].(float64))
	}
}

func TestCacheManager_GetCache(t *testing.T) {
	manager := NewCacheManager()
	defer manager.Stop()
	
	// Get a cache
	cache1 := manager.GetCache("test-cache", 1*time.Minute, 10)
	if cache1 == nil {
		t.Error("Expected to get a cache instance")
	}
	
	// Get the same cache again
	cache2 := manager.GetCache("test-cache", 2*time.Minute, 20)
	if cache1 != cache2 {
		t.Error("Expected to get the same cache instance")
	}
	
	// Get a different cache
	cache3 := manager.GetCache("other-cache", 1*time.Minute, 10)
	if cache1 == cache3 {
		t.Error("Expected to get a different cache instance")
	}
}

func TestCacheManager_Stats(t *testing.T) {
	manager := NewCacheManager()
	defer manager.Stop()
	
	// Create multiple caches and add data
	cache1 := manager.GetCache("cache1", 1*time.Minute, 10)
	cache2 := manager.GetCache("cache2", 1*time.Minute, 10)
	
	cache1.Set("key1", "value1")
	cache2.Set("key2", "value2")
	
	// Get stats
	stats := manager.GetStats()
	
	cache1Stats, ok := stats["cache1"]
	if !ok {
		t.Error("Expected cache1 stats")
	}
	
	cache2Stats, ok := stats["cache2"]
	if !ok {
		t.Error("Expected cache2 stats")
	}
	
	// Verify cache1 has 1 entry
	if cache1Stats.(map[string]interface{})["entries"].(int) != 1 {
		t.Error("Expected cache1 to have 1 entry")
	}
	
	// Verify cache2 has 1 entry
	if cache2Stats.(map[string]interface{})["entries"].(int) != 1 {
		t.Error("Expected cache2 to have 1 entry")
	}
}

func TestPredefinedCaches(t *testing.T) {
	// Test that predefined cache functions work
	smartCache := GetSMARTDataCache()
	if smartCache == nil {
		t.Error("Expected SMART data cache")
	}
	
	sensorCache := GetSensorDataCache()
	if sensorCache == nil {
		t.Error("Expected sensor data cache")
	}
	
	systemCache := GetSystemInfoCache()
	if systemCache == nil {
		t.Error("Expected system info cache")
	}
	
	diskCache := GetDiskInfoCache()
	if diskCache == nil {
		t.Error("Expected disk info cache")
	}
	
	containerCache := GetContainerInfoCache()
	if containerCache == nil {
		t.Error("Expected container info cache")
	}
	
	vmCache := GetVMInfoCache()
	if vmCache == nil {
		t.Error("Expected VM info cache")
	}
	
	generalCache := GetGeneralCache()
	if generalCache == nil {
		t.Error("Expected general cache")
	}
	
	// Verify they're different instances
	if smartCache == sensorCache {
		t.Error("Expected different cache instances")
	}
}

func TestCache_Cleanup(t *testing.T) {
	cache := NewCache(50*time.Millisecond, 10)
	defer cache.Stop()
	
	// Add entries that will expire
	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	
	// Verify they're there
	stats := cache.GetStats()
	if stats["entries"].(int) != 2 {
		t.Errorf("Expected 2 entries, got %d", stats["entries"].(int))
	}
	
	// Wait for expiration and cleanup
	time.Sleep(100 * time.Millisecond)
	
	// Trigger cleanup by accessing cache
	cache.Get("nonexistent")
	
	// Give cleanup time to run
	time.Sleep(50 * time.Millisecond)
	
	// Note: Cleanup runs in background, so we can't reliably test it
	// without exposing internal cleanup methods. This test mainly
	// ensures the cleanup doesn't crash.
}
