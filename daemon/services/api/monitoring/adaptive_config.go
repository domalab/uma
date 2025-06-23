package monitoring

import (
	"sync"
	"time"

	"github.com/domalab/uma/daemon/logger"
)

// AdaptiveConfig manages adaptive monitoring intervals based on system load and activity
type AdaptiveConfig struct {
	mutex sync.RWMutex

	// Base intervals (normal operation)
	baseIntervals map[string]time.Duration

	// Current intervals (adaptive)
	currentIntervals map[string]time.Duration

	// System load thresholds
	lowLoadThreshold    float64 // Below this, use slower intervals
	highLoadThreshold   float64 // Above this, use faster intervals
	criticalThreshold   float64 // Above this, use critical intervals

	// Activity tracking
	lastActivity     map[string]time.Time
	activityTimeout  time.Duration
	errorCounts      map[string]int
	errorThreshold   int

	// Adaptive factors
	slowFactor       float64 // Multiply interval by this when load is low
	fastFactor       float64 // Multiply interval by this when load is high
	criticalFactor   float64 // Multiply interval by this when critical
}

// MonitoringType represents different types of monitoring
type MonitoringType string

const (
	SystemStats      MonitoringType = "system_stats"
	DockerEvents     MonitoringType = "docker_events"
	StorageStatus    MonitoringType = "storage_status"
	VMEvents         MonitoringType = "vm_events"
	Infrastructure   MonitoringType = "infrastructure"
	ResourceAlerts   MonitoringType = "resource_alerts"
	Temperature      MonitoringType = "temperature"
	UPSDetection     MonitoringType = "ups_detection"
)

// NewAdaptiveConfig creates a new adaptive monitoring configuration
func NewAdaptiveConfig() *AdaptiveConfig {
	config := &AdaptiveConfig{
		baseIntervals: map[string]time.Duration{
			string(SystemStats):      3 * time.Second,
			string(DockerEvents):     5 * time.Second,
			string(StorageStatus):    10 * time.Second,
			string(VMEvents):         8 * time.Second,
			string(Infrastructure):   15 * time.Second,
			string(ResourceAlerts):   5 * time.Second,
			string(Temperature):      2 * time.Second,
			string(UPSDetection):     30 * time.Second,
		},
		currentIntervals: make(map[string]time.Duration),
		lastActivity:     make(map[string]time.Time),
		errorCounts:      make(map[string]int),

		// Load thresholds (CPU load average)
		lowLoadThreshold:  0.5,
		highLoadThreshold: 2.0,
		criticalThreshold: 4.0,

		// Activity timeout
		activityTimeout: 5 * time.Minute,
		errorThreshold:  3,

		// Adaptive factors
		slowFactor:     1.5, // 50% slower when load is low
		fastFactor:     0.7, // 30% faster when load is high
		criticalFactor: 0.5, // 50% faster when critical
	}

	// Initialize current intervals with base intervals
	for k, v := range config.baseIntervals {
		config.currentIntervals[k] = v
	}

	return config
}

// GetInterval returns the current adaptive interval for a monitoring type
func (ac *AdaptiveConfig) GetInterval(monitorType MonitoringType) time.Duration {
	ac.mutex.RLock()
	defer ac.mutex.RUnlock()

	if interval, exists := ac.currentIntervals[string(monitorType)]; exists {
		return interval
	}

	// Fallback to base interval
	if baseInterval, exists := ac.baseIntervals[string(monitorType)]; exists {
		return baseInterval
	}

	// Default fallback
	return 10 * time.Second
}

// UpdateSystemLoad updates the adaptive intervals based on current system load
func (ac *AdaptiveConfig) UpdateSystemLoad(loadAverage float64, memoryUsage float64, diskIO float64) {
	ac.mutex.Lock()
	defer ac.mutex.Unlock()

	var factor float64 = 1.0

	// Determine adaptive factor based on system load
	if loadAverage >= ac.criticalThreshold || memoryUsage >= 90.0 || diskIO >= 90.0 {
		factor = ac.criticalFactor
		logger.Yellow("High system load detected (%.2f), using critical monitoring intervals", loadAverage)
	} else if loadAverage >= ac.highLoadThreshold || memoryUsage >= 75.0 || diskIO >= 75.0 {
		factor = ac.fastFactor
		logger.Blue("Elevated system load detected (%.2f), using faster monitoring intervals", loadAverage)
	} else if loadAverage <= ac.lowLoadThreshold && memoryUsage <= 50.0 && diskIO <= 50.0 {
		factor = ac.slowFactor
		logger.Blue("Low system load detected (%.2f), using slower monitoring intervals", loadAverage)
	}

	// Apply factor to all intervals
	for monitorType, baseInterval := range ac.baseIntervals {
		newInterval := time.Duration(float64(baseInterval) * factor)
		
		// Enforce minimum and maximum bounds
		minInterval := 1 * time.Second
		maxInterval := 2 * time.Minute
		
		if newInterval < minInterval {
			newInterval = minInterval
		} else if newInterval > maxInterval {
			newInterval = maxInterval
		}
		
		ac.currentIntervals[monitorType] = newInterval
	}
}

// RecordActivity records activity for a monitoring type
func (ac *AdaptiveConfig) RecordActivity(monitorType MonitoringType) {
	ac.mutex.Lock()
	defer ac.mutex.Unlock()
	
	ac.lastActivity[string(monitorType)] = time.Now()
	
	// Reset error count on successful activity
	ac.errorCounts[string(monitorType)] = 0
}

// RecordError records an error for a monitoring type
func (ac *AdaptiveConfig) RecordError(monitorType MonitoringType, err error) {
	ac.mutex.Lock()
	defer ac.mutex.Unlock()
	
	ac.errorCounts[string(monitorType)]++
	
	// If error threshold is exceeded, slow down monitoring for this type
	if ac.errorCounts[string(monitorType)] >= ac.errorThreshold {
		currentInterval := ac.currentIntervals[string(monitorType)]
		newInterval := time.Duration(float64(currentInterval) * 2.0) // Double the interval
		
		// Enforce maximum bound
		maxInterval := 5 * time.Minute
		if newInterval > maxInterval {
			newInterval = maxInterval
		}
		
		ac.currentIntervals[string(monitorType)] = newInterval
		logger.Yellow("Monitoring type %s has %d errors, slowing interval to %v", 
			monitorType, ac.errorCounts[string(monitorType)], newInterval)
	}
}

// CheckInactiveMonitors checks for inactive monitors and adjusts their intervals
func (ac *AdaptiveConfig) CheckInactiveMonitors() {
	ac.mutex.Lock()
	defer ac.mutex.Unlock()
	
	now := time.Now()
	
	for monitorType, lastActivity := range ac.lastActivity {
		if now.Sub(lastActivity) > ac.activityTimeout {
			// Monitor has been inactive, slow it down
			currentInterval := ac.currentIntervals[monitorType]
			newInterval := time.Duration(float64(currentInterval) * 1.2) // 20% slower
			
			// Enforce maximum bound
			maxInterval := 2 * time.Minute
			if newInterval > maxInterval {
				newInterval = maxInterval
			}
			
			ac.currentIntervals[monitorType] = newInterval
			logger.Blue("Monitor %s inactive for %v, adjusting interval to %v", 
				monitorType, now.Sub(lastActivity), newInterval)
		}
	}
}

// GetConfiguration returns the current monitoring configuration
func (ac *AdaptiveConfig) GetConfiguration() map[string]interface{} {
	ac.mutex.RLock()
	defer ac.mutex.RUnlock()
	
	config := make(map[string]interface{})
	config["base_intervals"] = ac.baseIntervals
	config["current_intervals"] = ac.currentIntervals
	config["load_thresholds"] = map[string]float64{
		"low":      ac.lowLoadThreshold,
		"high":     ac.highLoadThreshold,
		"critical": ac.criticalThreshold,
	}
	config["adaptive_factors"] = map[string]float64{
		"slow":     ac.slowFactor,
		"fast":     ac.fastFactor,
		"critical": ac.criticalFactor,
	}
	config["error_counts"] = ac.errorCounts
	config["last_activity"] = ac.lastActivity
	
	return config
}

// ResetToDefaults resets all intervals to their base values
func (ac *AdaptiveConfig) ResetToDefaults() {
	ac.mutex.Lock()
	defer ac.mutex.Unlock()
	
	for k, v := range ac.baseIntervals {
		ac.currentIntervals[k] = v
	}
	
	// Clear error counts and activity tracking
	for k := range ac.errorCounts {
		ac.errorCounts[k] = 0
	}
	
	logger.Blue("Adaptive monitoring configuration reset to defaults")
}

// SetCustomInterval allows manual override of a specific monitoring interval
func (ac *AdaptiveConfig) SetCustomInterval(monitorType MonitoringType, interval time.Duration) {
	ac.mutex.Lock()
	defer ac.mutex.Unlock()
	
	ac.currentIntervals[string(monitorType)] = interval
	logger.Blue("Custom interval set for %s: %v", monitorType, interval)
}
