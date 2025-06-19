package daemon

import (
	"testing"
	"time"

	"github.com/domalab/uma/daemon/services/async"
	"github.com/domalab/uma/daemon/services/auth"
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

// TestRateLimitingBasicFunctionality tests basic rate limiting functionality
func TestRateLimitingBasicFunctionality(t *testing.T) {
	// Initialize rate limiter
	rateLimiter := auth.NewOperationRateLimiter()

	// Set a very restrictive rate limit for testing
	rateLimiter.SetLimit(auth.OpTypeGeneral, auth.RateLimit{Requests: 1, Window: time.Second})

	// Test first request - should succeed
	clientID := "test-client-1"
	allowed := rateLimiter.Allow(clientID, auth.OpTypeGeneral)
	if !allowed {
		t.Error("First request should be allowed")
	}

	// Test second request immediately - should be rate limited
	allowed = rateLimiter.Allow(clientID, auth.OpTypeGeneral)
	if allowed {
		t.Error("Second request should be rate limited")
	}

	// Test with different client ID - should be allowed
	allowed = rateLimiter.Allow("test-client-2", auth.OpTypeGeneral)
	if !allowed {
		t.Error("Request from different client should be allowed")
	}

	// Test with different operation type - should be allowed
	allowed = rateLimiter.Allow(clientID, auth.OpTypeHealthCheck)
	if !allowed {
		t.Error("Request for different operation should be allowed")
	}

	// Test rate limiter stats
	stats := rateLimiter.GetStats()
	if totalClients, ok := stats["total_clients"].(int); ok && totalClients < 1 {
		t.Error("Expected at least 1 client in stats")
	}

	t.Logf("Rate limiter stats: %+v", stats)

	// Wait for rate limit to reset and test again
	time.Sleep(1100 * time.Millisecond)
	allowed = rateLimiter.Allow(clientID, auth.OpTypeGeneral)
	if !allowed {
		t.Error("Request after rate limit reset should be allowed")
	}
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

// TestCacheAndRateLimiterIntegration tests integration between cache and rate limiter
func TestCacheAndRateLimiterIntegration(t *testing.T) {
	// Initialize cache manager
	cacheManager := cache.NewCacheManager()
	testCache := cacheManager.GetCache("integration_cache", 1*time.Minute, 50)

	// Initialize rate limiter
	rateLimiter := auth.NewOperationRateLimiter()

	// Test that both services work together
	clientID := "integration-test-client"

	// Store some data in cache
	testCache.Set("rate_limit_data", map[string]interface{}{
		"client_id": clientID,
		"requests":  0,
	})

	// Test rate limiting
	allowed := rateLimiter.Allow(clientID, auth.OpTypeGeneral)
	if !allowed {
		t.Error("First request should be allowed")
	}

	// Update cache with request count
	if data, found := testCache.Get("rate_limit_data"); found {
		if dataMap, ok := data.(map[string]interface{}); ok {
			dataMap["requests"] = 1
			testCache.Set("rate_limit_data", dataMap)
		}
	}

	// Verify cache update
	if data, found := testCache.Get("rate_limit_data"); found {
		if dataMap, ok := data.(map[string]interface{}); ok {
			if requests, ok := dataMap["requests"].(int); ok && requests != 1 {
				t.Errorf("Expected 1 request in cache, got %d", requests)
			}
		}
	}

	// Test rate limiter stats
	rateLimiterStats := rateLimiter.GetStats()
	cacheStats := testCache.GetStats()

	t.Logf("Rate limiter stats: %+v", rateLimiterStats)
	t.Logf("Cache stats: %+v", cacheStats)

	// Verify both services are working
	if totalClients, ok := rateLimiterStats["total_clients"].(int); ok && totalClients < 1 {
		t.Error("Expected at least 1 client in rate limiter")
	}

	if totalEntries, ok := cacheStats["total_entries"].(int); ok && totalEntries < 1 {
		t.Error("Expected at least 1 entry in cache")
	}
}
