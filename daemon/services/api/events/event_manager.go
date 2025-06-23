package events

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cskr/pubsub"
	"github.com/domalab/uma/daemon/logger"
	"github.com/domalab/uma/daemon/services/api/handlers"
	"github.com/domalab/uma/daemon/services/api/monitoring"
	"github.com/domalab/uma/daemon/services/api/utils"
)

// EventManager manages periodic data collection and event broadcasting
type EventManager struct {
	api       utils.APIInterface
	hub       *pubsub.PubSub
	wsHandler *handlers.WebSocketHandler
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup

	// Temperature monitoring
	tempMonitor    *monitoring.TemperatureMonitor
	tempThresholds map[string]float64
	lastTempAlerts map[string]time.Time

	// Adaptive monitoring configuration
	adaptiveConfig *monitoring.AdaptiveConfig
}

// NewEventManager creates a new event manager
func NewEventManager(api utils.APIInterface, hub *pubsub.PubSub, wsHandler *handlers.WebSocketHandler) *EventManager {
	ctx, cancel := context.WithCancel(context.Background())

	em := &EventManager{
		api:       api,
		hub:       hub,
		wsHandler: wsHandler,
		ctx:       ctx,
		cancel:    cancel,
		tempThresholds: map[string]float64{
			"cpu_warning":   70.0,
			"cpu_critical":  80.0,
			"disk_warning":  45.0,
			"disk_critical": 55.0,
		},
		lastTempAlerts: make(map[string]time.Time),
	}

	// Initialize temperature monitor
	em.tempMonitor = monitoring.NewTemperatureMonitor(api, wsHandler)

	// Initialize adaptive monitoring configuration
	em.adaptiveConfig = monitoring.NewAdaptiveConfig()

	return em
}

// Start starts the event manager
func (em *EventManager) Start() {
	logger.Green("Starting Event Manager")

	// Start temperature monitor
	if em.tempMonitor != nil {
		em.tempMonitor.Start()
	}

	// Start periodic data collectors with adaptive monitoring
	em.wg.Add(7)
	go em.collectSystemStats()
	go em.collectDockerEvents()
	go em.collectStorageStatus()
	go em.collectVMEvents()
	go em.collectInfrastructureStatus()
	go em.collectResourceAlerts()
	go em.adaptiveMonitoringManager()
}

// Stop stops the event manager
func (em *EventManager) Stop() {
	logger.Blue("Stopping Event Manager")

	// Stop temperature monitor
	if em.tempMonitor != nil {
		em.tempMonitor.Stop()
	}

	em.cancel()
	em.wg.Wait()
	logger.Green("Event Manager stopped")
}

// adaptiveMonitoringManager manages adaptive monitoring intervals
func (em *EventManager) adaptiveMonitoringManager() {
	defer em.wg.Done()

	ticker := time.NewTicker(30 * time.Second) // Check every 30 seconds
	defer ticker.Stop()

	for {
		select {
		case <-em.ctx.Done():
			return
		case <-ticker.C:
			em.updateAdaptiveIntervals()
		}
	}
}

// updateAdaptiveIntervals updates monitoring intervals based on system load
func (em *EventManager) updateAdaptiveIntervals() {
	// Get current system load
	if loadInfo, err := em.api.GetSystem().GetLoadInfo(); err == nil {
		if loadMap, ok := loadInfo.(map[string]interface{}); ok {
			var loadAverage float64
			if load1, ok := loadMap["load1"].(float64); ok {
				loadAverage = load1
			}

			// Get memory usage
			var memoryUsage float64
			if memInfo, err := em.api.GetSystem().GetMemoryInfo(); err == nil {
				if memMap, ok := memInfo.(map[string]interface{}); ok {
					if used, ok := memMap["used"].(float64); ok {
						if total, ok := memMap["total"].(float64); ok && total > 0 {
							memoryUsage = (used / total) * 100
						}
					}
				}
			}

			// Update adaptive configuration
			em.adaptiveConfig.UpdateSystemLoad(loadAverage, memoryUsage, 0) // TODO: Add disk I/O monitoring
		}
	}

	// Check for inactive monitors
	em.adaptiveConfig.CheckInactiveMonitors()
}

// collectSystemStats collects and broadcasts system statistics with adaptive intervals
func (em *EventManager) collectSystemStats() {
	defer em.wg.Done()

	// Use adaptive interval
	interval := em.adaptiveConfig.GetInterval(monitoring.SystemStats)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-em.ctx.Done():
			return
		case <-ticker.C:
			stats := em.getSystemStats()
			if stats != nil {
				em.adaptiveConfig.RecordActivity(monitoring.SystemStats)
				em.wsHandler.BroadcastEvent(handlers.EventSystemStats, stats)
			}

			// Update ticker with new adaptive interval
			newInterval := em.adaptiveConfig.GetInterval(monitoring.SystemStats)
			if newInterval != interval {
				interval = newInterval
				ticker.Stop()
				ticker = time.NewTicker(interval)
			}
		}
	}
}

// collectDockerEvents collects and broadcasts Docker events with adaptive intervals
func (em *EventManager) collectDockerEvents() {
	defer em.wg.Done()

	interval := em.adaptiveConfig.GetInterval(monitoring.DockerEvents)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-em.ctx.Done():
			return
		case <-ticker.C:
			events := em.getDockerEvents()
			if events != nil {
				em.adaptiveConfig.RecordActivity(monitoring.DockerEvents)
				em.wsHandler.BroadcastEvent(handlers.EventDockerEvents, events)
			}

			// Update ticker with new adaptive interval
			newInterval := em.adaptiveConfig.GetInterval(monitoring.DockerEvents)
			if newInterval != interval {
				interval = newInterval
				ticker.Stop()
				ticker = time.NewTicker(interval)
			}
		}
	}
}

// collectStorageStatus collects and broadcasts storage status with adaptive intervals
func (em *EventManager) collectStorageStatus() {
	defer em.wg.Done()

	interval := em.adaptiveConfig.GetInterval(monitoring.StorageStatus)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-em.ctx.Done():
			return
		case <-ticker.C:
			status := em.getStorageStatus()
			if status != nil {
				em.adaptiveConfig.RecordActivity(monitoring.StorageStatus)
				em.wsHandler.BroadcastEvent(handlers.EventStorageStatus, status)
			}

			// Update ticker with new adaptive interval
			newInterval := em.adaptiveConfig.GetInterval(monitoring.StorageStatus)
			if newInterval != interval {
				interval = newInterval
				ticker.Stop()
				ticker = time.NewTicker(interval)
			}
		}
	}
}

// Removed unused function: monitorTemperatures

// getSystemStats retrieves current system statistics
func (em *EventManager) getSystemStats() map[string]interface{} {
	stats := make(map[string]interface{})

	// Get CPU info
	if cpuInfo, err := em.api.GetSystem().GetCPUInfo(); err == nil {
		stats["cpu"] = cpuInfo
	} else {
		logger.Yellow("Event Manager: Failed to get CPU info: %v", err)
		stats["cpu"] = map[string]interface{}{"error": "failed to get CPU info"}
	}

	// Get memory info
	if memInfo, err := em.api.GetSystem().GetMemoryInfo(); err == nil {
		stats["memory"] = memInfo
	} else {
		logger.Yellow("Event Manager: Failed to get memory info: %v", err)
		stats["memory"] = map[string]interface{}{"error": "failed to get memory info"}
	}

	// Get load info
	if loadInfo, err := em.api.GetSystem().GetLoadInfo(); err == nil {
		stats["load"] = loadInfo
	} else {
		logger.Yellow("Event Manager: Failed to get load info: %v", err)
		stats["load"] = map[string]interface{}{"error": "failed to get load info"}
	}

	// Get network info
	if networkInfo, err := em.api.GetSystem().GetNetworkInfo(); err == nil {
		stats["network"] = networkInfo
	} else {
		logger.Yellow("Event Manager: Failed to get network info: %v", err)
		stats["network"] = map[string]interface{}{"error": "failed to get network info"}
	}

	// Get temperature data
	if tempData, err := em.api.GetSystem().GetEnhancedTemperatureData(); err == nil {
		stats["temperature"] = tempData
	} else {
		logger.Yellow("Event Manager: Failed to get temperature data: %v", err)
		stats["temperature"] = map[string]interface{}{"error": "failed to get temperature data"}
	}

	return stats
}

// getDockerEvents retrieves current Docker events/status
func (em *EventManager) getDockerEvents() map[string]interface{} {
	events := make(map[string]interface{})

	// Get container status
	if containers, err := em.api.GetDocker().GetContainers(); err == nil {
		events["containers"] = containers
	} else {
		em.adaptiveConfig.RecordError(monitoring.DockerEvents, err)
		logger.Yellow("Event Manager: Failed to get Docker containers: %v", err)
		events["containers"] = map[string]interface{}{"error": "failed to get containers"}
	}

	// Get Docker system info
	if info, err := em.api.GetDocker().GetSystemInfo(); err == nil {
		events["system_info"] = info
	} else {
		em.adaptiveConfig.RecordError(monitoring.DockerEvents, err)
		logger.Yellow("Event Manager: Failed to get Docker system info: %v", err)
		events["system_info"] = map[string]interface{}{"error": "failed to get system info"}
	}

	return events
}

// getStorageStatus retrieves current storage status
func (em *EventManager) getStorageStatus() map[string]interface{} {
	status := make(map[string]interface{})

	// Get array info
	if arrayInfo, err := em.api.GetStorage().GetArrayInfo(); err == nil {
		status["array"] = arrayInfo
	} else {
		em.adaptiveConfig.RecordError(monitoring.StorageStatus, err)
		logger.Yellow("Event Manager: Failed to get array info: %v", err)
		status["array"] = map[string]interface{}{"error": "failed to get array info"}
	}

	// Get disk info
	if disks, err := em.api.GetStorage().GetDisks(); err == nil {
		status["disks"] = disks
	} else {
		em.adaptiveConfig.RecordError(monitoring.StorageStatus, err)
		logger.Yellow("Event Manager: Failed to get disk info: %v", err)
		status["disks"] = map[string]interface{}{"error": "failed to get disk info"}
	}

	// Get cache info
	if cacheInfo, err := em.api.GetStorage().GetCacheInfo(); err == nil {
		status["cache"] = cacheInfo
	} else {
		em.adaptiveConfig.RecordError(monitoring.StorageStatus, err)
		logger.Yellow("Event Manager: Failed to get cache info: %v", err)
		status["cache"] = map[string]interface{}{"error": "failed to get cache info"}
	}

	return status
}

// collectVMEvents collects and broadcasts VM events with adaptive intervals
func (em *EventManager) collectVMEvents() {
	defer em.wg.Done()

	interval := em.adaptiveConfig.GetInterval(monitoring.VMEvents)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-em.ctx.Done():
			return
		case <-ticker.C:
			events := em.getVMEvents()
			if events != nil {
				em.adaptiveConfig.RecordActivity(monitoring.VMEvents)
				em.wsHandler.BroadcastEvent(handlers.EventVMEvents, events)
			}

			// Update ticker with new adaptive interval
			newInterval := em.adaptiveConfig.GetInterval(monitoring.VMEvents)
			if newInterval != interval {
				interval = newInterval
				ticker.Stop()
				ticker = time.NewTicker(interval)
			}
		}
	}
}

// collectInfrastructureStatus collects and broadcasts infrastructure status with adaptive intervals
func (em *EventManager) collectInfrastructureStatus() {
	defer em.wg.Done()

	interval := em.adaptiveConfig.GetInterval(monitoring.Infrastructure)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-em.ctx.Done():
			return
		case <-ticker.C:
			status := em.getInfrastructureStatus()
			if status != nil {
				em.adaptiveConfig.RecordActivity(monitoring.Infrastructure)
				em.wsHandler.BroadcastEvent(handlers.EventUPSStatus, status["ups"])
				em.wsHandler.BroadcastEvent(handlers.EventFanStatus, status["fans"])
				em.wsHandler.BroadcastEvent(handlers.EventPowerStatus, status["power"])
			}

			// Update ticker with new adaptive interval
			newInterval := em.adaptiveConfig.GetInterval(monitoring.Infrastructure)
			if newInterval != interval {
				interval = newInterval
				ticker.Stop()
				ticker = time.NewTicker(interval)
			}
		}
	}
}

// collectResourceAlerts monitors and broadcasts resource alerts with adaptive intervals
func (em *EventManager) collectResourceAlerts() {
	defer em.wg.Done()

	interval := em.adaptiveConfig.GetInterval(monitoring.ResourceAlerts)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-em.ctx.Done():
			return
		case <-ticker.C:
			em.checkResourceAlerts()
			em.adaptiveConfig.RecordActivity(monitoring.ResourceAlerts)

			// Update ticker with new adaptive interval
			newInterval := em.adaptiveConfig.GetInterval(monitoring.ResourceAlerts)
			if newInterval != interval {
				interval = newInterval
				ticker.Stop()
				ticker = time.NewTicker(interval)
			}
		}
	}
}

// getVMEvents retrieves current VM events/status
func (em *EventManager) getVMEvents() map[string]interface{} {
	events := make(map[string]interface{})

	// Get VM list and status
	if vms, err := em.api.GetVM().GetVMs(); err == nil {
		events["vms"] = vms
	}

	// Get VM system info if available
	events["vm_system_info"] = map[string]interface{}{
		"hypervisor": "KVM",
		"enabled":    true,
	}

	return events
}

// getInfrastructureStatus retrieves infrastructure status
func (em *EventManager) getInfrastructureStatus() map[string]interface{} {
	status := make(map[string]interface{})

	// UPS Status - use UPS detector interface
	upsDetector := em.api.GetUPSDetector()
	if upsDetector != nil && upsDetector.IsAvailable() {
		status["ups"] = upsDetector.GetStatus()
	} else {
		status["ups"] = map[string]interface{}{
			"available": false,
			"status":    "not_detected",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		}
	}

	// Fan Status - get from temperature data which includes fan info
	if tempData, err := em.api.GetSystem().GetEnhancedTemperatureData(); err == nil {
		if tempMap, ok := tempData.(map[string]interface{}); ok {
			if fans, exists := tempMap["fans"]; exists {
				status["fans"] = fans
			} else {
				status["fans"] = map[string]interface{}{"error": "Fan data not available in temperature data"}
			}
		}
	} else {
		status["fans"] = map[string]interface{}{"error": "Temperature data unavailable"}
	}

	// Power Status
	status["power"] = map[string]interface{}{
		"status":    "normal",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	return status
}

// checkResourceAlerts checks for resource threshold violations
func (em *EventManager) checkResourceAlerts() {
	// Check CPU usage
	if cpuInfo, err := em.api.GetSystem().GetCPUInfo(); err == nil {
		if cpuMap, ok := cpuInfo.(map[string]interface{}); ok {
			if usage, ok := cpuMap["usage_percent"].(float64); ok && usage > 90.0 {
				alert := map[string]interface{}{
					"resource":  "cpu",
					"metric":    "usage_percent",
					"value":     usage,
					"threshold": 90.0,
					"level":     "warning",
					"message":   fmt.Sprintf("High CPU usage: %.1f%%", usage),
					"timestamp": time.Now().UTC().Format(time.RFC3339),
				}
				em.wsHandler.BroadcastEvent(handlers.EventResourceAlert, alert)
			}
		}
	}

	// Check memory usage
	if memInfo, err := em.api.GetSystem().GetMemoryInfo(); err == nil {
		if memMap, ok := memInfo.(map[string]interface{}); ok {
			if usage, ok := memMap["usage_percent"].(float64); ok && usage > 85.0 {
				alert := map[string]interface{}{
					"resource":  "memory",
					"metric":    "usage_percent",
					"value":     usage,
					"threshold": 85.0,
					"level":     "warning",
					"message":   fmt.Sprintf("High memory usage: %.1f%%", usage),
					"timestamp": time.Now().UTC().Format(time.RFC3339),
				}
				em.wsHandler.BroadcastEvent(handlers.EventResourceAlert, alert)
			}
		}
	}

	// Check disk space
	if disks, err := em.api.GetStorage().GetDisks(); err == nil {
		if diskSlice, ok := disks.([]interface{}); ok {
			for _, disk := range diskSlice {
				if diskMap, ok := disk.(map[string]interface{}); ok {
					if usage, ok := diskMap["usage_percent"].(float64); ok && usage > 90.0 {
						diskName, _ := diskMap["name"].(string)
						alert := map[string]interface{}{
							"resource":  "disk",
							"metric":    "usage_percent",
							"value":     usage,
							"threshold": 90.0,
							"level":     "critical",
							"message":   fmt.Sprintf("High disk usage on %s: %.1f%%", diskName, usage),
							"timestamp": time.Now().UTC().Format(time.RFC3339),
						}
						em.wsHandler.BroadcastEvent(handlers.EventResourceAlert, alert)
					}
				}
			}
		}
	}
}

// Removed unused function: checkTemperatureAlerts

// Removed unused function: checkSensorTemperature

// Removed unused functions: sendTemperatureAlert, containsCPU, containsDisk
