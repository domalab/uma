package cache

import (
	"fmt"
	"strings"
	"time"

	"github.com/domalab/uma/daemon/logger"
)

// InvalidationStrategy defines how cache entries should be invalidated
type InvalidationStrategy interface {
	ShouldInvalidate(key string, entry *CacheEntry, event InvalidationEvent) bool
	GetDescription() string
}

// InvalidationEvent represents an event that might trigger cache invalidation
type InvalidationEvent struct {
	Type      EventType              `json:"type"`
	Resource  string                 `json:"resource"`
	Action    string                 `json:"action"`
	Details   map[string]interface{} `json:"details"`
	Timestamp time.Time              `json:"timestamp"`
}

// EventType represents the type of invalidation event
type EventType string

const (
	EventTypeStorage   EventType = "storage"
	EventTypeDocker    EventType = "docker"
	EventTypeVM        EventType = "vm"
	EventTypeSystem    EventType = "system"
	EventTypeNetwork   EventType = "network"
	EventTypeOperation EventType = "operation"
)

// PrefixInvalidationStrategy invalidates entries with matching key prefixes
type PrefixInvalidationStrategy struct {
	Prefixes []string
}

// ShouldInvalidate checks if the entry should be invalidated based on key prefix
func (s *PrefixInvalidationStrategy) ShouldInvalidate(key string, entry *CacheEntry, event InvalidationEvent) bool {
	for _, prefix := range s.Prefixes {
		if strings.HasPrefix(key, prefix) {
			return true
		}
	}
	return false
}

// GetDescription returns a description of the strategy
func (s *PrefixInvalidationStrategy) GetDescription() string {
	return fmt.Sprintf("Prefix invalidation for: %s", strings.Join(s.Prefixes, ", "))
}

// ResourceInvalidationStrategy invalidates entries related to specific resources
type ResourceInvalidationStrategy struct {
	ResourceType string
	Actions      []string
}

// ShouldInvalidate checks if the entry should be invalidated based on resource and action
func (s *ResourceInvalidationStrategy) ShouldInvalidate(key string, entry *CacheEntry, event InvalidationEvent) bool {
	// Check if the event matches our resource type
	if s.ResourceType != "" && string(event.Type) != s.ResourceType {
		return false
	}
	
	// Check if the action matches (if specified)
	if len(s.Actions) > 0 {
		actionMatches := false
		for _, action := range s.Actions {
			if event.Action == action {
				actionMatches = true
				break
			}
		}
		if !actionMatches {
			return false
		}
	}
	
	// Check if the key is related to the resource
	return strings.Contains(key, event.Resource)
}

// GetDescription returns a description of the strategy
func (s *ResourceInvalidationStrategy) GetDescription() string {
	return fmt.Sprintf("Resource invalidation for %s (actions: %s)", s.ResourceType, strings.Join(s.Actions, ", "))
}

// TimeBasedInvalidationStrategy invalidates entries older than a certain age
type TimeBasedInvalidationStrategy struct {
	MaxAge time.Duration
}

// ShouldInvalidate checks if the entry should be invalidated based on age
func (s *TimeBasedInvalidationStrategy) ShouldInvalidate(key string, entry *CacheEntry, event InvalidationEvent) bool {
	return time.Since(entry.CreatedAt) > s.MaxAge
}

// GetDescription returns a description of the strategy
func (s *TimeBasedInvalidationStrategy) GetDescription() string {
	return fmt.Sprintf("Time-based invalidation (max age: %v)", s.MaxAge)
}

// CacheInvalidator manages cache invalidation across multiple caches
type CacheInvalidator struct {
	strategies map[string][]InvalidationStrategy
	cacheManager *CacheManager
}

// NewCacheInvalidator creates a new cache invalidator
func NewCacheInvalidator(cacheManager *CacheManager) *CacheInvalidator {
	invalidator := &CacheInvalidator{
		strategies:   make(map[string][]InvalidationStrategy),
		cacheManager: cacheManager,
	}
	
	// Register default invalidation strategies
	invalidator.registerDefaultStrategies()
	
	return invalidator
}

// RegisterStrategy registers an invalidation strategy for a cache
func (ci *CacheInvalidator) RegisterStrategy(cacheName string, strategy InvalidationStrategy) {
	if ci.strategies[cacheName] == nil {
		ci.strategies[cacheName] = make([]InvalidationStrategy, 0)
	}
	
	ci.strategies[cacheName] = append(ci.strategies[cacheName], strategy)
	logger.Blue("Registered invalidation strategy for cache '%s': %s", cacheName, strategy.GetDescription())
}

// InvalidateOnEvent invalidates cache entries based on an event
func (ci *CacheInvalidator) InvalidateOnEvent(event InvalidationEvent) {
	totalInvalidated := 0
	
	for cacheName, strategies := range ci.strategies {
		cache := ci.cacheManager.caches[cacheName]
		if cache == nil {
			continue
		}
		
		invalidated := ci.invalidateCacheEntries(cache, strategies, event)
		if invalidated > 0 {
			logger.Yellow("Invalidated %d entries in cache '%s' due to %s event", 
				invalidated, cacheName, event.Type)
			totalInvalidated += invalidated
		}
	}
	
	if totalInvalidated > 0 {
		logger.Green("Total cache entries invalidated: %d", totalInvalidated)
	}
}

// invalidateCacheEntries invalidates entries in a specific cache
func (ci *CacheInvalidator) invalidateCacheEntries(cache *Cache, strategies []InvalidationStrategy, event InvalidationEvent) int {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()
	
	var keysToInvalidate []string
	
	for key, entry := range cache.entries {
		for _, strategy := range strategies {
			if strategy.ShouldInvalidate(key, entry, event) {
				keysToInvalidate = append(keysToInvalidate, key)
				break // One strategy match is enough
			}
		}
	}
	
	// Remove invalidated entries
	for _, key := range keysToInvalidate {
		delete(cache.entries, key)
	}
	
	return len(keysToInvalidate)
}

// registerDefaultStrategies registers default invalidation strategies for each cache
func (ci *CacheInvalidator) registerDefaultStrategies() {
	// SMART data cache invalidation
	ci.RegisterStrategy(SMARTDataCache, &ResourceInvalidationStrategy{
		ResourceType: "storage",
		Actions:      []string{"disk_added", "disk_removed", "array_started", "array_stopped"},
	})
	
	ci.RegisterStrategy(SMARTDataCache, &PrefixInvalidationStrategy{
		Prefixes: []string{"smart_", "disk_"},
	})
	
	// Sensor data cache invalidation
	ci.RegisterStrategy(SensorDataCache, &TimeBasedInvalidationStrategy{
		MaxAge: 2 * time.Minute, // Force refresh sensor data every 2 minutes
	})
	
	// System info cache invalidation
	ci.RegisterStrategy(SystemInfoCache, &ResourceInvalidationStrategy{
		ResourceType: "system",
		Actions:      []string{"reboot", "shutdown", "config_changed"},
	})
	
	// Disk info cache invalidation
	ci.RegisterStrategy(DiskInfoCache, &ResourceInvalidationStrategy{
		ResourceType: "storage",
		Actions:      []string{"disk_added", "disk_removed", "array_started", "array_stopped", "parity_check_completed"},
	})
	
	ci.RegisterStrategy(DiskInfoCache, &PrefixInvalidationStrategy{
		Prefixes: []string{"disk_", "array_"},
	})
	
	// Container info cache invalidation
	ci.RegisterStrategy(ContainerInfoCache, &ResourceInvalidationStrategy{
		ResourceType: "docker",
		Actions:      []string{"container_started", "container_stopped", "container_created", "container_removed"},
	})
	
	ci.RegisterStrategy(ContainerInfoCache, &PrefixInvalidationStrategy{
		Prefixes: []string{"container_", "docker_"},
	})
	
	// VM info cache invalidation
	ci.RegisterStrategy(VMInfoCache, &ResourceInvalidationStrategy{
		ResourceType: "vm",
		Actions:      []string{"vm_started", "vm_stopped", "vm_created", "vm_removed"},
	})
	
	ci.RegisterStrategy(VMInfoCache, &PrefixInvalidationStrategy{
		Prefixes: []string{"vm_", "libvirt_"},
	})
}

// Helper functions for creating invalidation events

// CreateStorageEvent creates a storage-related invalidation event
func CreateStorageEvent(resource, action string, details map[string]interface{}) InvalidationEvent {
	return InvalidationEvent{
		Type:      EventTypeStorage,
		Resource:  resource,
		Action:    action,
		Details:   details,
		Timestamp: time.Now(),
	}
}

// CreateDockerEvent creates a Docker-related invalidation event
func CreateDockerEvent(resource, action string, details map[string]interface{}) InvalidationEvent {
	return InvalidationEvent{
		Type:      EventTypeDocker,
		Resource:  resource,
		Action:    action,
		Details:   details,
		Timestamp: time.Now(),
	}
}

// CreateVMEvent creates a VM-related invalidation event
func CreateVMEvent(resource, action string, details map[string]interface{}) InvalidationEvent {
	return InvalidationEvent{
		Type:      EventTypeVM,
		Resource:  resource,
		Action:    action,
		Details:   details,
		Timestamp: time.Now(),
	}
}

// CreateSystemEvent creates a system-related invalidation event
func CreateSystemEvent(resource, action string, details map[string]interface{}) InvalidationEvent {
	return InvalidationEvent{
		Type:      EventTypeSystem,
		Resource:  resource,
		Action:    action,
		Details:   details,
		Timestamp: time.Now(),
	}
}

// CreateOperationEvent creates an operation-related invalidation event
func CreateOperationEvent(resource, action string, details map[string]interface{}) InvalidationEvent {
	return InvalidationEvent{
		Type:      EventTypeOperation,
		Resource:  resource,
		Action:    action,
		Details:   details,
		Timestamp: time.Now(),
	}
}

// Global cache invalidator instance
var globalInvalidator *CacheInvalidator

// InitializeGlobalInvalidator initializes the global cache invalidator
func InitializeGlobalInvalidator() {
	globalInvalidator = NewCacheInvalidator(globalCacheManager)
	logger.Blue("Initialized global cache invalidator")
}

// GetGlobalInvalidator returns the global cache invalidator
func GetGlobalInvalidator() *CacheInvalidator {
	if globalInvalidator == nil {
		InitializeGlobalInvalidator()
	}
	return globalInvalidator
}

// InvalidateCache is a convenience function to invalidate cache entries
func InvalidateCache(event InvalidationEvent) {
	if globalInvalidator != nil {
		globalInvalidator.InvalidateOnEvent(event)
	}
}
