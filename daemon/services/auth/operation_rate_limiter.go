package auth

import (
	"sync"
	"time"

	"github.com/domalab/uma/daemon/logger"
)

// OperationType represents different types of operations for rate limiting
type OperationType string

const (
	// General operations
	OpTypeGeneral     OperationType = "general"
	OpTypeHealthCheck OperationType = "health_check"

	// Storage operations
	OpTypeSMARTData    OperationType = "smart_data"
	OpTypeParityCheck  OperationType = "parity_check"
	OpTypeArrayControl OperationType = "array_control"
	OpTypeDiskInfo     OperationType = "disk_info"

	// Docker operations
	OpTypeDockerList    OperationType = "docker_list"
	OpTypeDockerControl OperationType = "docker_control"
	OpTypeDockerBulk    OperationType = "docker_bulk"

	// VM operations
	OpTypeVMList    OperationType = "vm_list"
	OpTypeVMControl OperationType = "vm_control"
	OpTypeVMBulk    OperationType = "vm_bulk"

	// System operations
	OpTypeSystemInfo    OperationType = "system_info"
	OpTypeSystemControl OperationType = "system_control"
	OpTypeSensorData    OperationType = "sensor_data"

	// Async operations
	OpTypeAsyncCreate OperationType = "async_create"
	OpTypeAsyncList   OperationType = "async_list"
	OpTypeAsyncCancel OperationType = "async_cancel"
)

// RateLimit represents a rate limit configuration
type RateLimit struct {
	Requests int           `json:"requests"`
	Window   time.Duration `json:"window"`
	Burst    int           `json:"burst,omitempty"` // Optional burst allowance
}

// OperationRateLimiter manages rate limiting for different operation types
type OperationRateLimiter struct {
	limits  map[OperationType]RateLimit
	buckets map[string]map[OperationType]*TokenBucket
	mutex   sync.RWMutex

	// Default limits
	defaultLimit RateLimit
}

// TokenBucket represents a token bucket for rate limiting
type TokenBucket struct {
	tokens     int
	maxTokens  int
	refillRate time.Duration
	lastRefill time.Time
	mutex      sync.Mutex
}

// NewTokenBucket creates a new token bucket
func NewTokenBucket(maxTokens int, refillRate time.Duration) *TokenBucket {
	return &TokenBucket{
		tokens:     maxTokens,
		maxTokens:  maxTokens,
		refillRate: refillRate,
		lastRefill: time.Now(),
	}
}

// TryConsume attempts to consume a token from the bucket
func (tb *TokenBucket) TryConsume() bool {
	tb.mutex.Lock()
	defer tb.mutex.Unlock()

	// Refill tokens based on elapsed time
	now := time.Now()
	elapsed := now.Sub(tb.lastRefill)

	if elapsed >= tb.refillRate {
		tokensToAdd := int(elapsed / tb.refillRate)
		tb.tokens = min(tb.maxTokens, tb.tokens+tokensToAdd)
		tb.lastRefill = now
	}

	// Try to consume a token
	if tb.tokens > 0 {
		tb.tokens--
		return true
	}

	return false
}

// GetTokens returns the current number of tokens
func (tb *TokenBucket) GetTokens() int {
	tb.mutex.Lock()
	defer tb.mutex.Unlock()
	return tb.tokens
}

// NewOperationRateLimiter creates a new operation-specific rate limiter
func NewOperationRateLimiter() *OperationRateLimiter {
	limiter := &OperationRateLimiter{
		limits:  make(map[OperationType]RateLimit),
		buckets: make(map[string]map[OperationType]*TokenBucket),
		defaultLimit: RateLimit{
			Requests: 60,
			Window:   time.Minute,
		},
	}

	// Set default limits for different operation types
	limiter.setDefaultLimits()

	return limiter
}

// setDefaultLimits sets default rate limits for different operation types
func (orl *OperationRateLimiter) setDefaultLimits() {
	// General operations - moderate limits
	orl.limits[OpTypeGeneral] = RateLimit{Requests: 60, Window: time.Minute}
	orl.limits[OpTypeHealthCheck] = RateLimit{Requests: 120, Window: time.Minute}

	// Storage operations - conservative limits for expensive operations
	orl.limits[OpTypeSMARTData] = RateLimit{Requests: 1, Window: time.Minute}    // Very expensive
	orl.limits[OpTypeParityCheck] = RateLimit{Requests: 1, Window: time.Hour}    // Extremely expensive
	orl.limits[OpTypeArrayControl] = RateLimit{Requests: 2, Window: time.Minute} // Critical operations
	orl.limits[OpTypeDiskInfo] = RateLimit{Requests: 10, Window: time.Minute}    // Moderate cost

	// Docker operations - moderate limits
	orl.limits[OpTypeDockerList] = RateLimit{Requests: 30, Window: time.Minute}
	orl.limits[OpTypeDockerControl] = RateLimit{Requests: 20, Window: time.Minute}
	orl.limits[OpTypeDockerBulk] = RateLimit{Requests: 5, Window: time.Minute} // Bulk operations are expensive

	// VM operations - moderate limits
	orl.limits[OpTypeVMList] = RateLimit{Requests: 30, Window: time.Minute}
	orl.limits[OpTypeVMControl] = RateLimit{Requests: 10, Window: time.Minute}
	orl.limits[OpTypeVMBulk] = RateLimit{Requests: 3, Window: time.Minute} // Bulk operations are expensive

	// System operations - varied limits
	orl.limits[OpTypeSystemInfo] = RateLimit{Requests: 60, Window: time.Minute}
	orl.limits[OpTypeSystemControl] = RateLimit{Requests: 5, Window: time.Minute} // Critical operations
	orl.limits[OpTypeSensorData] = RateLimit{Requests: 30, Window: time.Minute}   // Sensor polling

	// Async operations - moderate limits
	orl.limits[OpTypeAsyncCreate] = RateLimit{Requests: 10, Window: time.Minute}
	orl.limits[OpTypeAsyncList] = RateLimit{Requests: 60, Window: time.Minute}
	orl.limits[OpTypeAsyncCancel] = RateLimit{Requests: 20, Window: time.Minute}
}

// Allow checks if an operation is allowed for a client
func (orl *OperationRateLimiter) Allow(clientID string, operationType OperationType) bool {
	orl.mutex.Lock()
	defer orl.mutex.Unlock()

	// Get or create client buckets
	if orl.buckets[clientID] == nil {
		orl.buckets[clientID] = make(map[OperationType]*TokenBucket)
	}

	// Get or create bucket for this operation type
	bucket := orl.buckets[clientID][operationType]
	if bucket == nil {
		limit, exists := orl.limits[operationType]
		if !exists {
			limit = orl.defaultLimit
		}

		// Calculate refill rate based on requests per window
		refillRate := limit.Window / time.Duration(limit.Requests)
		bucket = NewTokenBucket(limit.Requests, refillRate)
		orl.buckets[clientID][operationType] = bucket
	}

	// Try to consume a token
	allowed := bucket.TryConsume()

	if !allowed {
		logger.Yellow("Rate limit exceeded for client %s, operation %s", clientID, operationType)
	}

	return allowed
}

// SetLimit sets a custom rate limit for an operation type
func (orl *OperationRateLimiter) SetLimit(operationType OperationType, limit RateLimit) {
	orl.mutex.Lock()
	defer orl.mutex.Unlock()

	orl.limits[operationType] = limit

	// Clear existing buckets for this operation type to apply new limits
	for clientID := range orl.buckets {
		if orl.buckets[clientID][operationType] != nil {
			delete(orl.buckets[clientID], operationType)
		}
	}

	logger.Blue("Set rate limit for operation %s: %d requests per %v",
		operationType, limit.Requests, limit.Window)
}

// GetLimit returns the rate limit for an operation type
func (orl *OperationRateLimiter) GetLimit(operationType OperationType) RateLimit {
	orl.mutex.RLock()
	defer orl.mutex.RUnlock()

	if limit, exists := orl.limits[operationType]; exists {
		return limit
	}
	return orl.defaultLimit
}

// GetStats returns rate limiting statistics
func (orl *OperationRateLimiter) GetStats() map[string]interface{} {
	orl.mutex.RLock()
	defer orl.mutex.RUnlock()

	stats := map[string]interface{}{
		"total_clients":    len(orl.buckets),
		"operation_limits": make(map[string]interface{}),
		"client_stats":     make(map[string]interface{}),
	}

	// Add operation limits
	for opType, limit := range orl.limits {
		stats["operation_limits"].(map[string]interface{})[string(opType)] = map[string]interface{}{
			"requests": limit.Requests,
			"window":   limit.Window.String(),
		}
	}

	// Add client statistics
	for clientID, buckets := range orl.buckets {
		clientStats := make(map[string]interface{})
		for opType, bucket := range buckets {
			clientStats[string(opType)] = map[string]interface{}{
				"tokens_remaining": bucket.GetTokens(),
				"max_tokens":       bucket.maxTokens,
			}
		}
		stats["client_stats"].(map[string]interface{})[clientID] = clientStats
	}

	return stats
}

// CleanupStaleClients removes buckets for clients that haven't been seen recently
func (orl *OperationRateLimiter) CleanupStaleClients(maxAge time.Duration) {
	orl.mutex.Lock()
	defer orl.mutex.Unlock()

	cutoff := time.Now().Add(-maxAge)
	var staleClients []string

	for clientID, buckets := range orl.buckets {
		allStale := true
		for _, bucket := range buckets {
			if bucket.lastRefill.After(cutoff) {
				allStale = false
				break
			}
		}

		if allStale {
			staleClients = append(staleClients, clientID)
		}
	}

	for _, clientID := range staleClients {
		delete(orl.buckets, clientID)
	}

	if len(staleClients) > 0 {
		logger.Blue("Cleaned up %d stale rate limit clients", len(staleClients))
	}
}

// Reset resets rate limits for a specific client
func (orl *OperationRateLimiter) Reset(clientID string) {
	orl.mutex.Lock()
	defer orl.mutex.Unlock()

	delete(orl.buckets, clientID)
	logger.Blue("Reset rate limits for client: %s", clientID)
}

// ResetAll resets all rate limits
func (orl *OperationRateLimiter) ResetAll() {
	orl.mutex.Lock()
	defer orl.mutex.Unlock()

	orl.buckets = make(map[string]map[OperationType]*TokenBucket)
	logger.Blue("Reset all rate limits")
}

// Helper function for min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// OperationRateLimitMiddleware creates middleware for operation-specific rate limiting
func (orl *OperationRateLimiter) OperationRateLimitMiddleware(operationType OperationType) func(next func()) bool {
	return func(next func()) bool {
		// This would be used in HTTP handlers to check rate limits
		// Implementation depends on how client ID is extracted
		return true // Placeholder
	}
}

// GetOperationTypeFromPath determines operation type from request path
func GetOperationTypeFromPath(path string) OperationType {
	switch {
	// Health checks
	case path == "/api/v1/health":
		return OpTypeHealthCheck

	// Storage operations
	case path == "/api/v1/storage/smart" || path == "/api/v1/storage/smart/":
		return OpTypeSMARTData
	case path == "/api/v1/storage/parity" || path == "/api/v1/storage/parity/":
		return OpTypeParityCheck
	case path == "/api/v1/storage/array" || path == "/api/v1/storage/array/":
		return OpTypeArrayControl
	case path == "/api/v1/storage/disks" || path == "/api/v1/storage/disks/":
		return OpTypeDiskInfo

	// Docker operations
	case path == "/api/v1/docker/containers":
		return OpTypeDockerList
	case path == "/api/v1/docker/bulk":
		return OpTypeDockerBulk

	// VM operations
	case path == "/api/v1/vms":
		return OpTypeVMList
	case path == "/api/v1/vms/bulk":
		return OpTypeVMBulk

	// System operations
	case path == "/api/v1/system/info" || path == "/api/v1/system/cpu" || path == "/api/v1/system/memory":
		return OpTypeSystemInfo
	case path == "/api/v1/system/reboot" || path == "/api/v1/system/shutdown":
		return OpTypeSystemControl
	case path == "/api/v1/sensors" || path == "/api/v1/sensors/":
		return OpTypeSensorData

	// Async operations
	case path == "/api/v1/operations":
		return OpTypeAsyncCreate // POST creates, GET lists
	case path == "/api/v1/operations/":
		return OpTypeAsyncCancel // DELETE cancels

	default:
		return OpTypeGeneral
	}
}
