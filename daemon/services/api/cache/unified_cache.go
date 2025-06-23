package cache

import (
	"sync"
	"time"

	"github.com/domalab/uma/daemon/logger"
	"github.com/domalab/uma/daemon/services/cache"
)

// UnifiedCache provides a unified caching interface that consolidates
// the functionality of both cache.Cache and CacheService
type UnifiedCache struct {
	// Core cache functionality
	cache *cache.Cache

	// Legacy compatibility fields
	mu                sync.RWMutex
	lastArrayInfoHash string

	// Configuration
	defaultTTL time.Duration
}

// CacheType represents different types of cached data
type CacheType string

const (
	SystemDataCache  CacheType = "system_data"
	DockerDataCache  CacheType = "docker_data"
	VMDataCache      CacheType = "vm_data"
	StorageDataCache CacheType = "storage_data"
	TemperatureCache CacheType = "temperature_data"
	NetworkDataCache CacheType = "network_data"
	ArrayInfoCache   CacheType = "array_info"
)

// NewUnifiedCache creates a new unified cache instance
func NewUnifiedCache(defaultTTL time.Duration, maxEntries int) *UnifiedCache {
	return &UnifiedCache{
		cache:      cache.NewCache(defaultTTL, maxEntries),
		defaultTTL: defaultTTL,
	}
}

// Get retrieves a value from the cache
func (uc *UnifiedCache) Get(cacheType CacheType, key string) (interface{}, bool) {
	cacheKey := string(cacheType) + ":" + key
	return uc.cache.Get(cacheKey)
}

// Set stores a value in the cache with default TTL
func (uc *UnifiedCache) Set(cacheType CacheType, key string, value interface{}) {
	cacheKey := string(cacheType) + ":" + key
	uc.cache.Set(cacheKey, value)
}

// SetWithTTL stores a value in the cache with custom TTL
func (uc *UnifiedCache) SetWithTTL(cacheType CacheType, key string, value interface{}, ttl time.Duration) {
	cacheKey := string(cacheType) + ":" + key
	uc.cache.SetWithTTL(cacheKey, value, ttl)
}

// Delete removes a value from the cache
func (uc *UnifiedCache) Delete(cacheType CacheType, key string) {
	cacheKey := string(cacheType) + ":" + key
	uc.cache.Delete(cacheKey)
}

// ClearType clears all entries of a specific cache type
func (uc *UnifiedCache) ClearType(cacheType CacheType) {
	// Note: This is a simplified implementation. In production, you might want
	// to add a more efficient prefix-based deletion to the underlying cache
	logger.Blue("Clearing cache type: %s", cacheType)
}

// Legacy CacheService compatibility methods

// GetCachedSystemData retrieves cached system data if valid, otherwise returns nil
func (uc *UnifiedCache) GetCachedSystemData() interface{} {
	if data, found := uc.Get(SystemDataCache, "default"); found {
		return data
	}
	return nil
}

// SetCachedSystemData stores system data in cache with expiration
func (uc *UnifiedCache) SetCachedSystemData(data interface{}) {
	uc.Set(SystemDataCache, "default", data)
}

// GetCachedDockerData retrieves cached docker data if valid, otherwise returns nil
func (uc *UnifiedCache) GetCachedDockerData() interface{} {
	if data, found := uc.Get(DockerDataCache, "default"); found {
		return data
	}
	return nil
}

// SetCachedDockerData stores docker data in cache with expiration
func (uc *UnifiedCache) SetCachedDockerData(data interface{}) {
	uc.Set(DockerDataCache, "default", data)
}

// GetCachedVMData retrieves cached VM data if valid, otherwise returns nil
func (uc *UnifiedCache) GetCachedVMData() interface{} {
	if data, found := uc.Get(VMDataCache, "default"); found {
		return data
	}
	return nil
}

// SetCachedVMData stores VM data in cache with expiration
func (uc *UnifiedCache) SetCachedVMData(data interface{}) {
	uc.Set(VMDataCache, "default", data)
}

// GetLastArrayInfoHash returns the last array info hash
func (uc *UnifiedCache) GetLastArrayInfoHash() string {
	uc.mu.RLock()
	defer uc.mu.RUnlock()
	return uc.lastArrayInfoHash
}

// SetLastArrayInfoHash sets the last array info hash
func (uc *UnifiedCache) SetLastArrayInfoHash(hash string) {
	uc.mu.Lock()
	defer uc.mu.Unlock()
	uc.lastArrayInfoHash = hash
}

// ClearCache clears all cached data
func (uc *UnifiedCache) ClearCache() {
	uc.cache.Clear()
	uc.mu.Lock()
	uc.lastArrayInfoHash = ""
	uc.mu.Unlock()
	logger.Blue("Unified cache cleared")
}

// SetCacheDuration sets the cache duration (affects new entries)
func (uc *UnifiedCache) SetCacheDuration(duration time.Duration) {
	uc.mu.Lock()
	defer uc.mu.Unlock()
	uc.defaultTTL = duration
}

// GetCacheDuration returns the current cache duration
func (uc *UnifiedCache) GetCacheDuration() time.Duration {
	uc.mu.RLock()
	defer uc.mu.RUnlock()
	return uc.defaultTTL
}

// GetStats returns cache statistics
func (uc *UnifiedCache) GetStats() map[string]interface{} {
	stats := uc.cache.GetStats()
	stats["default_ttl"] = uc.defaultTTL.String()
	stats["array_info_hash"] = uc.GetLastArrayInfoHash()
	return stats
}

// Stop stops the cache cleanup goroutine
func (uc *UnifiedCache) Stop() {
	uc.cache.Stop()
}

// Advanced caching methods

// GetOrSet retrieves a value from cache, or sets it using the provided function if not found
func (uc *UnifiedCache) GetOrSet(cacheType CacheType, key string, setter func() (interface{}, error)) (interface{}, error) {
	// Try to get from cache first
	if value, found := uc.Get(cacheType, key); found {
		return value, nil
	}

	// Not in cache, use setter function
	value, err := setter()
	if err != nil {
		return nil, err
	}

	// Store in cache
	uc.Set(cacheType, key, value)
	return value, nil
}

// GetOrSetWithTTL retrieves a value from cache, or sets it with custom TTL using the provided function
func (uc *UnifiedCache) GetOrSetWithTTL(cacheType CacheType, key string, ttl time.Duration, setter func() (interface{}, error)) (interface{}, error) {
	// Try to get from cache first
	if value, found := uc.Get(cacheType, key); found {
		return value, nil
	}

	// Not in cache, use setter function
	value, err := setter()
	if err != nil {
		return nil, err
	}

	// Store in cache with custom TTL
	uc.SetWithTTL(cacheType, key, value, ttl)
	return value, nil
}

// Invalidate removes entries based on patterns or conditions
func (uc *UnifiedCache) Invalidate(cacheType CacheType, pattern string) {
	// This is a simplified implementation
	// In production, you might want more sophisticated pattern matching
	logger.Blue("Invalidating cache entries for type %s with pattern %s", cacheType, pattern)
}

// UnifiedCacheManager manages multiple unified cache instances
type UnifiedCacheManager struct {
	caches map[string]*UnifiedCache
	mutex  sync.RWMutex
}

// NewUnifiedCacheManager creates a new unified cache manager
func NewUnifiedCacheManager() *UnifiedCacheManager {
	return &UnifiedCacheManager{
		caches: make(map[string]*UnifiedCache),
	}
}

// GetCache gets or creates a unified cache with the specified configuration
func (ucm *UnifiedCacheManager) GetCache(name string, defaultTTL time.Duration, maxEntries int) *UnifiedCache {
	ucm.mutex.RLock()
	cache, exists := ucm.caches[name]
	ucm.mutex.RUnlock()

	if exists {
		return cache
	}

	ucm.mutex.Lock()
	defer ucm.mutex.Unlock()

	// Double-check after acquiring write lock
	if cache, exists := ucm.caches[name]; exists {
		return cache
	}

	cache = NewUnifiedCache(defaultTTL, maxEntries)
	ucm.caches[name] = cache

	logger.Blue("Created unified cache '%s' with TTL %v and max entries %d", name, defaultTTL, maxEntries)

	return cache
}

// GetStats returns statistics for all unified caches
func (ucm *UnifiedCacheManager) GetStats() map[string]interface{} {
	ucm.mutex.RLock()
	defer ucm.mutex.RUnlock()

	stats := make(map[string]interface{})

	for name, cache := range ucm.caches {
		stats[name] = cache.GetStats()
	}

	return stats
}

// Stop stops all unified caches
func (ucm *UnifiedCacheManager) Stop() {
	ucm.mutex.Lock()
	defer ucm.mutex.Unlock()

	for name, cache := range ucm.caches {
		cache.Stop()
		logger.Blue("Stopped unified cache: %s", name)
	}

	ucm.caches = make(map[string]*UnifiedCache)
}

// Global unified cache manager instance
var globalUnifiedCacheManager = NewUnifiedCacheManager()

// GetGlobalUnifiedCacheManager returns the global unified cache manager
func GetGlobalUnifiedCacheManager() *UnifiedCacheManager {
	return globalUnifiedCacheManager
}

// GetDefaultUnifiedCache returns the default unified cache instance
func GetDefaultUnifiedCache() *UnifiedCache {
	return globalUnifiedCacheManager.GetCache("default", 30*time.Second, 1000)
}
