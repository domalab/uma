package services

import (
	"time"

	"github.com/domalab/uma/daemon/services/api/cache"
)

// CacheService handles caching operations for the API using unified cache
type CacheService struct {
	unifiedCache *cache.UnifiedCache
}

// NewCacheService creates a new cache service instance
func NewCacheService() *CacheService {
	return &CacheService{
		unifiedCache: cache.GetDefaultUnifiedCache(),
	}
}

// GetCachedSystemData retrieves cached system data if valid, otherwise returns nil
func (c *CacheService) GetCachedSystemData() interface{} {
	return c.unifiedCache.GetCachedSystemData()
}

// SetCachedSystemData stores system data in cache with expiration
func (c *CacheService) SetCachedSystemData(data interface{}) {
	c.unifiedCache.SetCachedSystemData(data)
}

// GetCachedDockerData retrieves cached docker data if valid, otherwise returns nil
func (c *CacheService) GetCachedDockerData() interface{} {
	return c.unifiedCache.GetCachedDockerData()
}

// SetCachedDockerData stores docker data in cache with expiration
func (c *CacheService) SetCachedDockerData(data interface{}) {
	c.unifiedCache.SetCachedDockerData(data)
}

// GetCachedVMData retrieves cached VM data if valid, otherwise returns nil
func (c *CacheService) GetCachedVMData() interface{} {
	return c.unifiedCache.GetCachedVMData()
}

// SetCachedVMData stores VM data in cache with expiration
func (c *CacheService) SetCachedVMData(data interface{}) {
	c.unifiedCache.SetCachedVMData(data)
}

// GetLastArrayInfoHash returns the last array info hash
func (c *CacheService) GetLastArrayInfoHash() string {
	return c.unifiedCache.GetLastArrayInfoHash()
}

// SetLastArrayInfoHash sets the last array info hash
func (c *CacheService) SetLastArrayInfoHash(hash string) {
	c.unifiedCache.SetLastArrayInfoHash(hash)
}

// ClearCache clears all cached data
func (c *CacheService) ClearCache() {
	c.unifiedCache.ClearCache()
}

// SetCacheDuration sets the cache duration
func (c *CacheService) SetCacheDuration(duration time.Duration) {
	c.unifiedCache.SetCacheDuration(duration)
}

// GetCacheDuration returns the current cache duration
func (c *CacheService) GetCacheDuration() time.Duration {
	return c.unifiedCache.GetCacheDuration()
}

// GetStats returns cache statistics
func (c *CacheService) GetStats() map[string]interface{} {
	return c.unifiedCache.GetStats()
}

// Advanced caching methods using unified cache

// GetOrSetSystemData retrieves system data from cache or sets it using the provided function
func (c *CacheService) GetOrSetSystemData(setter func() (interface{}, error)) (interface{}, error) {
	return c.unifiedCache.GetOrSet(cache.SystemDataCache, "default", setter)
}

// GetOrSetDockerData retrieves docker data from cache or sets it using the provided function
func (c *CacheService) GetOrSetDockerData(setter func() (interface{}, error)) (interface{}, error) {
	return c.unifiedCache.GetOrSet(cache.DockerDataCache, "default", setter)
}

// GetOrSetVMData retrieves VM data from cache or sets it using the provided function
func (c *CacheService) GetOrSetVMData(setter func() (interface{}, error)) (interface{}, error) {
	return c.unifiedCache.GetOrSet(cache.VMDataCache, "default", setter)
}
