package services

import (
	"sync"
	"time"

	"github.com/domalab/uma/daemon/services/api/types/models"
)

// CacheService handles caching operations for the API
type CacheService struct {
	mu                sync.RWMutex
	systemData        *models.CacheEntry
	dockerData        *models.CacheEntry
	vmData            *models.CacheEntry
	cacheDuration     time.Duration
	lastArrayInfoHash string
}

// NewCacheService creates a new cache service instance
func NewCacheService() *CacheService {
	return &CacheService{
		cacheDuration: 30 * time.Second, // Cache for 30 seconds
	}
}

// GetCachedSystemData retrieves cached system data if valid, otherwise returns nil
func (c *CacheService) GetCachedSystemData() interface{} {
	return c.getCachedData(&c.systemData)
}

// SetCachedSystemData stores system data in cache with expiration
func (c *CacheService) SetCachedSystemData(data interface{}) {
	c.setCachedData(&c.systemData, data)
}

// GetCachedDockerData retrieves cached docker data if valid, otherwise returns nil
func (c *CacheService) GetCachedDockerData() interface{} {
	return c.getCachedData(&c.dockerData)
}

// SetCachedDockerData stores docker data in cache with expiration
func (c *CacheService) SetCachedDockerData(data interface{}) {
	c.setCachedData(&c.dockerData, data)
}

// GetCachedVMData retrieves cached VM data if valid, otherwise returns nil
func (c *CacheService) GetCachedVMData() interface{} {
	return c.getCachedData(&c.vmData)
}

// SetCachedVMData stores VM data in cache with expiration
func (c *CacheService) SetCachedVMData(data interface{}) {
	c.setCachedData(&c.vmData, data)
}

// GetLastArrayInfoHash returns the last array info hash
func (c *CacheService) GetLastArrayInfoHash() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.lastArrayInfoHash
}

// SetLastArrayInfoHash sets the last array info hash
func (c *CacheService) SetLastArrayInfoHash(hash string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.lastArrayInfoHash = hash
}

// getCachedData retrieves cached data if valid, otherwise returns nil
func (c *CacheService) getCachedData(entry **models.CacheEntry) interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if *entry != nil && time.Now().Before((*entry).ExpiresAt) {
		return (*entry).Data
	}
	return nil
}

// setCachedData stores data in cache with expiration
func (c *CacheService) setCachedData(entry **models.CacheEntry, data interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	*entry = &models.CacheEntry{
		Data:      data,
		ExpiresAt: time.Now().Add(c.cacheDuration),
	}
}

// ClearCache clears all cached data
func (c *CacheService) ClearCache() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.systemData = nil
	c.dockerData = nil
	c.vmData = nil
	c.lastArrayInfoHash = ""
}

// SetCacheDuration sets the cache duration
func (c *CacheService) SetCacheDuration(duration time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cacheDuration = duration
}

// GetCacheDuration returns the current cache duration
func (c *CacheService) GetCacheDuration() time.Duration {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.cacheDuration
}
