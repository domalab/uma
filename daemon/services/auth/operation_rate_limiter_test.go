package auth

import (
	"testing"
	"time"
)

// TestNewOperationRateLimiter tests the creation of a new operation rate limiter
func TestNewOperationRateLimiter(t *testing.T) {
	limiter := NewOperationRateLimiter()

	if limiter == nil {
		t.Fatal("Expected non-nil operation rate limiter")
	}

	// Check default limits are set
	generalLimit := limiter.GetLimit(OpTypeGeneral)
	if generalLimit.Requests != 60 {
		t.Errorf("Expected general limit 60 requests, got %d", generalLimit.Requests)
	}

	if generalLimit.Window != time.Minute {
		t.Errorf("Expected general window 1 minute, got %v", generalLimit.Window)
	}

	// Check expensive operation limits
	smartLimit := limiter.GetLimit(OpTypeSMARTData)
	if smartLimit.Requests != 1 {
		t.Errorf("Expected SMART data limit 1 request, got %d", smartLimit.Requests)
	}

	if smartLimit.Window != time.Minute {
		t.Errorf("Expected SMART data window 1 minute, got %v", smartLimit.Window)
	}

	parityLimit := limiter.GetLimit(OpTypeParityCheck)
	if parityLimit.Requests != 1 {
		t.Errorf("Expected parity check limit 1 request, got %d", parityLimit.Requests)
	}

	if parityLimit.Window != time.Hour {
		t.Errorf("Expected parity check window 1 hour, got %v", parityLimit.Window)
	}
}

// TestOperationRateLimiterAllow tests the Allow method
func TestOperationRateLimiterAllow(t *testing.T) {
	limiter := NewOperationRateLimiter()

	// Set a restrictive limit for testing
	limiter.SetLimit(OpTypeGeneral, RateLimit{
		Requests: 2,
		Window:   time.Second,
	})

	clientID := "test-client-1"

	// First request should be allowed
	if !limiter.Allow(clientID, OpTypeGeneral) {
		t.Error("First request should be allowed")
	}

	// Second request should be allowed
	if !limiter.Allow(clientID, OpTypeGeneral) {
		t.Error("Second request should be allowed")
	}

	// Third request should be denied (rate limit exceeded)
	if limiter.Allow(clientID, OpTypeGeneral) {
		t.Error("Third request should be denied")
	}

	// Different client should be allowed
	if !limiter.Allow("test-client-2", OpTypeGeneral) {
		t.Error("Different client should be allowed")
	}

	// Different operation type should be allowed
	if !limiter.Allow(clientID, OpTypeHealthCheck) {
		t.Error("Different operation type should be allowed")
	}
}

// TestOperationRateLimiterRefill tests token bucket refill
func TestOperationRateLimiterRefill(t *testing.T) {
	limiter := NewOperationRateLimiter()

	// Set a limit that refills quickly for testing
	limiter.SetLimit(OpTypeGeneral, RateLimit{
		Requests: 1,
		Window:   100 * time.Millisecond,
	})

	clientID := "test-client"

	// First request should be allowed
	if !limiter.Allow(clientID, OpTypeGeneral) {
		t.Error("First request should be allowed")
	}

	// Second request should be denied immediately
	if limiter.Allow(clientID, OpTypeGeneral) {
		t.Error("Second request should be denied immediately")
	}

	// Wait for refill
	time.Sleep(150 * time.Millisecond)

	// Request should be allowed after refill
	if !limiter.Allow(clientID, OpTypeGeneral) {
		t.Error("Request should be allowed after refill")
	}
}

// TestSetLimit tests setting custom rate limits
func TestSetLimit(t *testing.T) {
	limiter := NewOperationRateLimiter()

	customLimit := RateLimit{
		Requests: 10,
		Window:   5 * time.Minute,
	}

	limiter.SetLimit(OpTypeDockerList, customLimit)

	retrievedLimit := limiter.GetLimit(OpTypeDockerList)
	if retrievedLimit.Requests != customLimit.Requests {
		t.Errorf("Expected %d requests, got %d", customLimit.Requests, retrievedLimit.Requests)
	}

	if retrievedLimit.Window != customLimit.Window {
		t.Errorf("Expected %v window, got %v", customLimit.Window, retrievedLimit.Window)
	}
}

// TestGetLimit tests getting rate limits
func TestGetLimit(t *testing.T) {
	limiter := NewOperationRateLimiter()

	// Test existing operation type
	limit := limiter.GetLimit(OpTypeDockerList)
	if limit.Requests <= 0 {
		t.Error("Expected positive request limit")
	}

	if limit.Window <= 0 {
		t.Error("Expected positive window duration")
	}

	// Test non-existent operation type (should return default)
	unknownLimit := limiter.GetLimit(OperationType("unknown"))
	defaultLimit := limiter.GetLimit(OpTypeGeneral)

	if unknownLimit.Requests != defaultLimit.Requests {
		t.Errorf("Expected default requests %d for unknown type, got %d",
			defaultLimit.Requests, unknownLimit.Requests)
	}
}

// TestGetStats tests statistics retrieval
func TestGetStats(t *testing.T) {
	limiter := NewOperationRateLimiter()

	// Make some requests to generate stats
	limiter.Allow("client1", OpTypeGeneral)
	limiter.Allow("client1", OpTypeDockerList)
	limiter.Allow("client2", OpTypeGeneral)

	stats := limiter.GetStats()

	// Check total clients
	totalClients, ok := stats["total_clients"].(int)
	if !ok {
		t.Fatal("Expected total_clients in stats")
	}

	if totalClients != 2 {
		t.Errorf("Expected 2 total clients, got %d", totalClients)
	}

	// Check operation limits
	operationLimits, ok := stats["operation_limits"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected operation_limits in stats")
	}

	if len(operationLimits) == 0 {
		t.Error("Expected non-empty operation limits")
	}

	// Check client stats
	clientStats, ok := stats["client_stats"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected client_stats in stats")
	}

	if len(clientStats) != 2 {
		t.Errorf("Expected 2 clients in stats, got %d", len(clientStats))
	}
}

// TestReset tests resetting rate limits for a client
func TestReset(t *testing.T) {
	limiter := NewOperationRateLimiter()

	// Set restrictive limit
	limiter.SetLimit(OpTypeGeneral, RateLimit{
		Requests: 1,
		Window:   time.Hour,
	})

	clientID := "test-client"

	// Use up the limit
	if !limiter.Allow(clientID, OpTypeGeneral) {
		t.Error("First request should be allowed")
	}

	if limiter.Allow(clientID, OpTypeGeneral) {
		t.Error("Second request should be denied")
	}

	// Reset the client
	limiter.Reset(clientID)

	// Should be allowed again after reset
	if !limiter.Allow(clientID, OpTypeGeneral) {
		t.Error("Request should be allowed after reset")
	}
}

// TestResetAll tests resetting all rate limits
func TestResetAll(t *testing.T) {
	limiter := NewOperationRateLimiter()

	// Set restrictive limit
	limiter.SetLimit(OpTypeGeneral, RateLimit{
		Requests: 1,
		Window:   time.Hour,
	})

	// Use up limits for multiple clients
	limiter.Allow("client1", OpTypeGeneral)
	limiter.Allow("client2", OpTypeGeneral)

	// Both should be denied
	if limiter.Allow("client1", OpTypeGeneral) {
		t.Error("Client1 second request should be denied")
	}

	if limiter.Allow("client2", OpTypeGeneral) {
		t.Error("Client2 second request should be denied")
	}

	// Reset all
	limiter.ResetAll()

	// Both should be allowed again
	if !limiter.Allow("client1", OpTypeGeneral) {
		t.Error("Client1 should be allowed after reset all")
	}

	if !limiter.Allow("client2", OpTypeGeneral) {
		t.Error("Client2 should be allowed after reset all")
	}
}

// TestCleanupStaleClients tests cleanup of stale client buckets
func TestCleanupStaleClients(t *testing.T) {
	limiter := NewOperationRateLimiter()

	// Create some client activity
	limiter.Allow("active-client", OpTypeGeneral)
	limiter.Allow("stale-client", OpTypeGeneral)

	// Manually set last refill time to simulate stale client
	limiter.mutex.Lock()
	if bucket := limiter.buckets["stale-client"][OpTypeGeneral]; bucket != nil {
		bucket.lastRefill = time.Now().Add(-2 * time.Hour)
	}
	limiter.mutex.Unlock()

	// Cleanup stale clients (older than 1 hour)
	limiter.CleanupStaleClients(time.Hour)

	stats := limiter.GetStats()
	totalClients := stats["total_clients"].(int)

	// Should only have the active client
	if totalClients != 1 {
		t.Errorf("Expected 1 client after cleanup, got %d", totalClients)
	}

	clientStats := stats["client_stats"].(map[string]interface{})
	if _, exists := clientStats["stale-client"]; exists {
		t.Error("Stale client should have been cleaned up")
	}

	if _, exists := clientStats["active-client"]; !exists {
		t.Error("Active client should still exist")
	}
}

// TestGetOperationTypeFromPath tests operation type detection from paths
func TestGetOperationTypeFromPath(t *testing.T) {
	testCases := []struct {
		path     string
		expected OperationType
	}{
		{"/api/v1/health", OpTypeHealthCheck},
		{"/api/v1/storage/smart", OpTypeSMARTData},
		{"/api/v1/storage/parity", OpTypeParityCheck},
		{"/api/v1/storage/array", OpTypeArrayControl},
		{"/api/v1/storage/disks", OpTypeDiskInfo},
		{"/api/v1/docker/containers", OpTypeDockerList},
		{"/api/v1/docker/bulk", OpTypeDockerBulk},
		{"/api/v1/vms", OpTypeVMList},
		{"/api/v1/vms/bulk", OpTypeVMBulk},
		{"/api/v1/system/info", OpTypeSystemInfo},
		{"/api/v1/system/reboot", OpTypeSystemControl},
		{"/api/v1/sensors", OpTypeSensorData},
		{"/api/v1/operations", OpTypeAsyncCreate},
		{"/api/v1/operations/", OpTypeAsyncCancel},
		{"/unknown/path", OpTypeGeneral},
	}

	for _, tc := range testCases {
		result := GetOperationTypeFromPath(tc.path)
		if result != tc.expected {
			t.Errorf("Path %s: expected %s, got %s", tc.path, tc.expected, result)
		}
	}
}

// TestNewTokenBucket tests token bucket creation
func TestNewTokenBucket(t *testing.T) {
	maxTokens := 10
	refillRate := time.Second

	bucket := NewTokenBucket(maxTokens, refillRate)

	if bucket == nil {
		t.Fatal("Expected non-nil token bucket")
	}

	if bucket.tokens != maxTokens {
		t.Errorf("Expected %d initial tokens, got %d", maxTokens, bucket.tokens)
	}

	if bucket.maxTokens != maxTokens {
		t.Errorf("Expected %d max tokens, got %d", maxTokens, bucket.maxTokens)
	}

	if bucket.refillRate != refillRate {
		t.Errorf("Expected %v refill rate, got %v", refillRate, bucket.refillRate)
	}
}

// TestTokenBucketTryConsume tests token consumption
func TestTokenBucketTryConsume(t *testing.T) {
	bucket := NewTokenBucket(2, time.Second)

	// First consumption should succeed
	if !bucket.TryConsume() {
		t.Error("First consumption should succeed")
	}

	if bucket.GetTokens() != 1 {
		t.Errorf("Expected 1 token remaining, got %d", bucket.GetTokens())
	}

	// Second consumption should succeed
	if !bucket.TryConsume() {
		t.Error("Second consumption should succeed")
	}

	if bucket.GetTokens() != 0 {
		t.Errorf("Expected 0 tokens remaining, got %d", bucket.GetTokens())
	}

	// Third consumption should fail
	if bucket.TryConsume() {
		t.Error("Third consumption should fail")
	}
}

// TestTokenBucketRefill tests token refill
func TestTokenBucketRefill(t *testing.T) {
	bucket := NewTokenBucket(3, 50*time.Millisecond)

	// Consume all tokens
	bucket.TryConsume()
	bucket.TryConsume()
	bucket.TryConsume()

	if bucket.GetTokens() != 0 {
		t.Errorf("Expected 0 tokens after consumption, got %d", bucket.GetTokens())
	}

	// Wait for refill
	time.Sleep(100 * time.Millisecond) // Should refill 2 tokens

	// Should be able to consume again
	if !bucket.TryConsume() {
		t.Error("Should be able to consume after refill")
	}

	// Should have at least 1 token remaining
	if bucket.GetTokens() < 1 {
		t.Errorf("Expected at least 1 token after refill, got %d", bucket.GetTokens())
	}
}

// TestNewRateLimiter tests basic rate limiter creation
func TestNewRateLimiter(t *testing.T) {
	limit := 10
	window := time.Minute

	limiter := NewRateLimiter(limit, window)

	if limiter == nil {
		t.Fatal("Expected non-nil rate limiter")
	}

	if limiter.limit != limit {
		t.Errorf("Expected limit %d, got %d", limit, limiter.limit)
	}

	if limiter.window != window {
		t.Errorf("Expected window %v, got %v", window, limiter.window)
	}

	if limiter.requests == nil {
		t.Error("Expected non-nil requests map")
	}
}

// TestRateLimiterAllow tests basic rate limiter allow method
func TestRateLimiterAllow(t *testing.T) {
	limiter := NewRateLimiter(2, time.Second)

	ip := "192.168.1.100"

	// First request should be allowed
	if !limiter.Allow(ip) {
		t.Error("First request should be allowed")
	}

	// Second request should be allowed
	if !limiter.Allow(ip) {
		t.Error("Second request should be allowed")
	}

	// Third request should be denied
	if limiter.Allow(ip) {
		t.Error("Third request should be denied")
	}

	// Different IP should be allowed
	if !limiter.Allow("192.168.1.101") {
		t.Error("Different IP should be allowed")
	}
}

// TestRateLimiterWindowExpiry tests request window expiry
func TestRateLimiterWindowExpiry(t *testing.T) {
	limiter := NewRateLimiter(1, 100*time.Millisecond)

	ip := "192.168.1.100"

	// First request should be allowed
	if !limiter.Allow(ip) {
		t.Error("First request should be allowed")
	}

	// Second request should be denied immediately
	if limiter.Allow(ip) {
		t.Error("Second request should be denied immediately")
	}

	// Wait for window to expire
	time.Sleep(150 * time.Millisecond)

	// Request should be allowed after window expiry
	if !limiter.Allow(ip) {
		t.Error("Request should be allowed after window expiry")
	}
}

// BenchmarkOperationRateLimiterAllow benchmarks the Allow method
func BenchmarkOperationRateLimiterAllow(b *testing.B) {
	limiter := NewOperationRateLimiter()
	clientID := "benchmark-client"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		limiter.Allow(clientID, OpTypeGeneral)
	}
}

// BenchmarkTokenBucketTryConsume benchmarks token bucket consumption
func BenchmarkTokenBucketTryConsume(b *testing.B) {
	bucket := NewTokenBucket(1000, time.Microsecond)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bucket.TryConsume()
	}
}
