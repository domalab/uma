package metrics

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/domalab/uma/daemon/logger"
)

// MetricsCollector manages background collection of system metrics
type MetricsCollector struct {
	ctx        context.Context
	cancel     context.CancelFunc
	cache      *MetricsCache
	collectors map[string]*CollectorConfig
	running    bool
	mutex      sync.RWMutex
}

// CollectorConfig defines configuration for a specific metric collector
type CollectorConfig struct {
	Name        string
	Interval    time.Duration
	Enabled     bool
	LastRun     time.Time
	CollectFunc func() (interface{}, error)
}

// MetricsCache stores collected metrics with TTL
type MetricsCache struct {
	data  map[string]*CacheEntry
	mutex sync.RWMutex
}

// CacheEntry represents a cached metric with timestamp
type CacheEntry struct {
	Data      interface{}
	Timestamp time.Time
	TTL       time.Duration
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector() *MetricsCollector {
	ctx, cancel := context.WithCancel(context.Background())

	return &MetricsCollector{
		ctx:        ctx,
		cancel:     cancel,
		cache:      NewMetricsCache(),
		collectors: make(map[string]*CollectorConfig),
		running:    false,
	}
}

// NewMetricsCache creates a new metrics cache
func NewMetricsCache() *MetricsCache {
	return &MetricsCache{
		data: make(map[string]*CacheEntry),
	}
}

// RegisterCollector registers a new metric collector
func (mc *MetricsCollector) RegisterCollector(name string, interval time.Duration, collectFunc func() (interface{}, error)) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	mc.collectors[name] = &CollectorConfig{
		Name:        name,
		Interval:    interval,
		Enabled:     true,
		CollectFunc: collectFunc,
	}

	logger.Blue("Registered metrics collector: %s (interval: %v)", name, interval)
}

// Start begins background metrics collection
func (mc *MetricsCollector) Start() error {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	if mc.running {
		return nil
	}

	mc.running = true

	// Start collection goroutines for each collector
	for name, config := range mc.collectors {
		if config.Enabled {
			go mc.runCollector(name, config)
		}
	}

	logger.Green("Started background metrics collection with %d collectors", len(mc.collectors))
	return nil
}

// Stop stops background metrics collection
func (mc *MetricsCollector) Stop() error {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	if !mc.running {
		return nil
	}

	mc.cancel()
	mc.running = false

	logger.Yellow("Stopped background metrics collection")
	return nil
}

// runCollector runs a specific collector in a goroutine
func (mc *MetricsCollector) runCollector(name string, config *CollectorConfig) {
	ticker := time.NewTicker(config.Interval)
	defer ticker.Stop()

	// Initial collection
	mc.collectMetric(name, config)

	for {
		select {
		case <-mc.ctx.Done():
			logger.LightGreen("Stopping collector: %s", name)
			return
		case <-ticker.C:
			mc.collectMetric(name, config)
		}
	}
}

// collectMetric collects a single metric and caches the result
func (mc *MetricsCollector) collectMetric(name string, config *CollectorConfig) {
	start := time.Now()

	data, err := config.CollectFunc()
	if err != nil {
		logger.Yellow("Failed to collect metric %s: %v", name, err)
		return
	}

	// Cache the result
	mc.cache.Set(name, data, config.Interval*2) // TTL = 2x collection interval

	// Update last run time
	mc.mutex.Lock()
	config.LastRun = time.Now()
	mc.mutex.Unlock()

	duration := time.Since(start)
	if duration > 100*time.Millisecond {
		logger.LightGreen("Collected metric %s in %v", name, duration)
	}
}

// GetMetric retrieves a cached metric
func (mc *MetricsCollector) GetMetric(name string) (interface{}, bool) {
	return mc.cache.Get(name)
}

// GetMetricWithFallback retrieves a cached metric or collects it if not available
func (mc *MetricsCollector) GetMetricWithFallback(name string) (interface{}, error) {
	// Try cache first
	if data, found := mc.cache.Get(name); found {
		return data, nil
	}

	// Fallback to direct collection
	mc.mutex.RLock()
	config, exists := mc.collectors[name]
	mc.mutex.RUnlock()

	if !exists {
		return nil, ErrCollectorNotFound
	}

	return config.CollectFunc()
}

// Set stores data in the cache
func (cache *MetricsCache) Set(key string, data interface{}, ttl time.Duration) {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	cache.data[key] = &CacheEntry{
		Data:      data,
		Timestamp: time.Now(),
		TTL:       ttl,
	}
}

// Get retrieves data from the cache
func (cache *MetricsCache) Get(key string) (interface{}, bool) {
	cache.mutex.RLock()
	defer cache.mutex.RUnlock()

	entry, exists := cache.data[key]
	if !exists {
		return nil, false
	}

	// Check if entry has expired
	if time.Since(entry.Timestamp) > entry.TTL {
		// Clean up expired entry
		go func() {
			cache.mutex.Lock()
			delete(cache.data, key)
			cache.mutex.Unlock()
		}()
		return nil, false
	}

	return entry.Data, true
}

// GetStats returns collector statistics
func (mc *MetricsCollector) GetStats() map[string]interface{} {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()

	stats := map[string]interface{}{
		"running":          mc.running,
		"collectors_count": len(mc.collectors),
		"cache_entries":    len(mc.cache.data),
		"collectors":       make(map[string]interface{}),
	}

	for name, config := range mc.collectors {
		stats["collectors"].(map[string]interface{})[name] = map[string]interface{}{
			"enabled":  config.Enabled,
			"interval": config.Interval.String(),
			"last_run": config.LastRun.Format(time.RFC3339),
		}
	}

	return stats
}

// Error definitions
var (
	ErrCollectorNotFound = fmt.Errorf("collector not found")
)
