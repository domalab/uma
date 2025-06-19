package cache

import (
	"sync"
	"time"

	"github.com/domalab/uma/daemon/logger"
)

// CacheEntry represents a cached item with expiration
type CacheEntry struct {
	Value     interface{}
	ExpiresAt time.Time
	CreatedAt time.Time
	AccessCount int64
	LastAccess  time.Time
}

// IsExpired checks if the cache entry has expired
func (e *CacheEntry) IsExpired() bool {
	return time.Now().After(e.ExpiresAt)
}

// Cache represents a thread-safe cache with TTL support
type Cache struct {
	entries map[string]*CacheEntry
	mutex   sync.RWMutex
	
	// Configuration
	defaultTTL    time.Duration
	maxEntries    int
	cleanupInterval time.Duration
	
	// Statistics
	hits   int64
	misses int64
	
	// Cleanup
	stopCleanup chan struct{}
	cleanupWG   sync.WaitGroup
}

// NewCache creates a new cache with the specified configuration
func NewCache(defaultTTL time.Duration, maxEntries int) *Cache {
	cache := &Cache{
		entries:         make(map[string]*CacheEntry),
		defaultTTL:      defaultTTL,
		maxEntries:      maxEntries,
		cleanupInterval: 5 * time.Minute,
		stopCleanup:     make(chan struct{}),
	}
	
	// Start cleanup goroutine
	cache.startCleanup()
	
	return cache
}

// Get retrieves a value from the cache
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	
	entry, exists := c.entries[key]
	if !exists {
		c.misses++
		return nil, false
	}
	
	if entry.IsExpired() {
		c.misses++
		// Don't delete here to avoid write lock, cleanup will handle it
		return nil, false
	}
	
	// Update access statistics
	entry.AccessCount++
	entry.LastAccess = time.Now()
	c.hits++
	
	return entry.Value, true
}

// Set stores a value in the cache with default TTL
func (c *Cache) Set(key string, value interface{}) {
	c.SetWithTTL(key, value, c.defaultTTL)
}

// SetWithTTL stores a value in the cache with custom TTL
func (c *Cache) SetWithTTL(key string, value interface{}, ttl time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	// Check if we need to evict entries
	if len(c.entries) >= c.maxEntries {
		c.evictLRU()
	}
	
	now := time.Now()
	c.entries[key] = &CacheEntry{
		Value:       value,
		ExpiresAt:   now.Add(ttl),
		CreatedAt:   now,
		AccessCount: 0,
		LastAccess:  now,
	}
}

// Delete removes a value from the cache
func (c *Cache) Delete(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	delete(c.entries, key)
}

// Clear removes all entries from the cache
func (c *Cache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	c.entries = make(map[string]*CacheEntry)
	c.hits = 0
	c.misses = 0
}

// GetStats returns cache statistics
func (c *Cache) GetStats() map[string]interface{} {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	
	total := c.hits + c.misses
	hitRate := float64(0)
	if total > 0 {
		hitRate = float64(c.hits) / float64(total) * 100
	}
	
	return map[string]interface{}{
		"entries":   len(c.entries),
		"hits":      c.hits,
		"misses":    c.misses,
		"hit_rate":  hitRate,
		"max_entries": c.maxEntries,
	}
}

// evictLRU evicts the least recently used entry
func (c *Cache) evictLRU() {
	if len(c.entries) == 0 {
		return
	}
	
	var oldestKey string
	var oldestTime time.Time
	
	for key, entry := range c.entries {
		if oldestKey == "" || entry.LastAccess.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.LastAccess
		}
	}
	
	if oldestKey != "" {
		delete(c.entries, oldestKey)
		logger.Blue("Evicted cache entry: %s", oldestKey)
	}
}

// startCleanup starts the cleanup goroutine
func (c *Cache) startCleanup() {
	c.cleanupWG.Add(1)
	go func() {
		defer c.cleanupWG.Done()
		
		ticker := time.NewTicker(c.cleanupInterval)
		defer ticker.Stop()
		
		for {
			select {
			case <-ticker.C:
				c.cleanup()
			case <-c.stopCleanup:
				return
			}
		}
	}()
}

// cleanup removes expired entries
func (c *Cache) cleanup() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	var expiredKeys []string
	
	for key, entry := range c.entries {
		if entry.IsExpired() {
			expiredKeys = append(expiredKeys, key)
		}
	}
	
	for _, key := range expiredKeys {
		delete(c.entries, key)
	}
	
	if len(expiredKeys) > 0 {
		logger.Blue("Cleaned up %d expired cache entries", len(expiredKeys))
	}
}

// Stop stops the cache cleanup goroutine
func (c *Cache) Stop() {
	close(c.stopCleanup)
	c.cleanupWG.Wait()
	logger.Blue("Cache stopped")
}

// CacheManager manages multiple named caches
type CacheManager struct {
	caches map[string]*Cache
	mutex  sync.RWMutex
}

// NewCacheManager creates a new cache manager
func NewCacheManager() *CacheManager {
	return &CacheManager{
		caches: make(map[string]*Cache),
	}
}

// GetCache gets or creates a cache with the specified name and configuration
func (cm *CacheManager) GetCache(name string, defaultTTL time.Duration, maxEntries int) *Cache {
	cm.mutex.RLock()
	cache, exists := cm.caches[name]
	cm.mutex.RUnlock()
	
	if exists {
		return cache
	}
	
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	
	// Double-check after acquiring write lock
	if cache, exists := cm.caches[name]; exists {
		return cache
	}
	
	cache = NewCache(defaultTTL, maxEntries)
	cm.caches[name] = cache
	
	logger.Blue("Created cache '%s' with TTL %v and max entries %d", name, defaultTTL, maxEntries)
	
	return cache
}

// GetStats returns statistics for all caches
func (cm *CacheManager) GetStats() map[string]interface{} {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	
	stats := make(map[string]interface{})
	
	for name, cache := range cm.caches {
		stats[name] = cache.GetStats()
	}
	
	return stats
}

// Stop stops all caches
func (cm *CacheManager) Stop() {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	
	for name, cache := range cm.caches {
		cache.Stop()
		logger.Blue("Stopped cache: %s", name)
	}
	
	cm.caches = make(map[string]*Cache)
}

// Global cache manager instance
var globalCacheManager = NewCacheManager()

// GetGlobalCacheManager returns the global cache manager
func GetGlobalCacheManager() *CacheManager {
	return globalCacheManager
}

// Predefined cache configurations
const (
	SMARTDataTTL     = 5 * time.Minute  // SMART data cache TTL
	SensorDataTTL    = 30 * time.Second // Sensor data cache TTL
	SystemInfoTTL    = 1 * time.Minute  // System info cache TTL
	DiskInfoTTL      = 2 * time.Minute  // Disk info cache TTL
	ContainerInfoTTL = 30 * time.Second // Container info cache TTL
	VMInfoTTL        = 30 * time.Second // VM info cache TTL
)

// Cache names
const (
	SMARTDataCache     = "smart_data"
	SensorDataCache    = "sensor_data"
	SystemInfoCache    = "system_info"
	DiskInfoCache      = "disk_info"
	ContainerInfoCache = "container_info"
	VMInfoCache        = "vm_info"
	GeneralCache       = "general"
)

// Helper functions for common cache operations

// GetSMARTDataCache returns the SMART data cache
func GetSMARTDataCache() *Cache {
	return globalCacheManager.GetCache(SMARTDataCache, SMARTDataTTL, 100)
}

// GetSensorDataCache returns the sensor data cache
func GetSensorDataCache() *Cache {
	return globalCacheManager.GetCache(SensorDataCache, SensorDataTTL, 50)
}

// GetSystemInfoCache returns the system info cache
func GetSystemInfoCache() *Cache {
	return globalCacheManager.GetCache(SystemInfoCache, SystemInfoTTL, 20)
}

// GetDiskInfoCache returns the disk info cache
func GetDiskInfoCache() *Cache {
	return globalCacheManager.GetCache(DiskInfoCache, DiskInfoTTL, 50)
}

// GetContainerInfoCache returns the container info cache
func GetContainerInfoCache() *Cache {
	return globalCacheManager.GetCache(ContainerInfoCache, ContainerInfoTTL, 100)
}

// GetVMInfoCache returns the VM info cache
func GetVMInfoCache() *Cache {
	return globalCacheManager.GetCache(VMInfoCache, VMInfoTTL, 50)
}

// GetGeneralCache returns the general purpose cache
func GetGeneralCache() *Cache {
	return globalCacheManager.GetCache(GeneralCache, 5*time.Minute, 200)
}
