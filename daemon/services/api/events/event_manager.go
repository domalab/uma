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
	wsHandler *handlers.EnhancedWebSocketHandler
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup

	// Temperature monitoring
	tempMonitor    *monitoring.TemperatureMonitor
	tempThresholds map[string]float64
	lastTempAlerts map[string]time.Time
	// Removed unused field: tempMutex
}

// NewEventManager creates a new event manager
func NewEventManager(api utils.APIInterface, hub *pubsub.PubSub, wsHandler *handlers.EnhancedWebSocketHandler) *EventManager {
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

	return em
}

// Start starts the event manager
func (em *EventManager) Start() {
	logger.Green("Starting Event Manager")

	// Start temperature monitor
	if em.tempMonitor != nil {
		em.tempMonitor.Start()
	}

	// Start periodic data collectors
	em.wg.Add(6)
	go em.collectSystemStats()
	go em.collectDockerEvents()
	go em.collectStorageStatus()
	go em.collectVMEvents()
	go em.collectInfrastructureStatus()
	go em.collectResourceAlerts()
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

// collectSystemStats collects and broadcasts system statistics
func (em *EventManager) collectSystemStats() {
	defer em.wg.Done()

	ticker := time.NewTicker(3 * time.Second) // 3-second intervals for system stats
	defer ticker.Stop()

	for {
		select {
		case <-em.ctx.Done():
			return
		case <-ticker.C:
			stats := em.getSystemStats()
			em.wsHandler.BroadcastEvent(handlers.EventSystemStats, stats)
		}
	}
}

// collectDockerEvents collects and broadcasts Docker events
func (em *EventManager) collectDockerEvents() {
	defer em.wg.Done()

	ticker := time.NewTicker(5 * time.Second) // 5-second intervals for Docker events
	defer ticker.Stop()

	for {
		select {
		case <-em.ctx.Done():
			return
		case <-ticker.C:
			events := em.getDockerEvents()
			em.wsHandler.BroadcastEvent(handlers.EventDockerEvents, events)
		}
	}
}

// collectStorageStatus collects and broadcasts storage status
func (em *EventManager) collectStorageStatus() {
	defer em.wg.Done()

	ticker := time.NewTicker(10 * time.Second) // 10-second intervals for storage status
	defer ticker.Stop()

	for {
		select {
		case <-em.ctx.Done():
			return
		case <-ticker.C:
			status := em.getStorageStatus()
			em.wsHandler.BroadcastEvent(handlers.EventStorageStatus, status)
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
	}

	// Get Docker system info
	if info, err := em.api.GetDocker().GetSystemInfo(); err == nil {
		events["system_info"] = info
	}

	return events
}

// getStorageStatus retrieves current storage status
func (em *EventManager) getStorageStatus() map[string]interface{} {
	status := make(map[string]interface{})

	// Get array info
	if arrayInfo, err := em.api.GetStorage().GetArrayInfo(); err == nil {
		status["array"] = arrayInfo
	}

	// Get disk info
	if disks, err := em.api.GetStorage().GetDisks(); err == nil {
		status["disks"] = disks
	}

	// Get cache info
	if cacheInfo, err := em.api.GetStorage().GetCacheInfo(); err == nil {
		status["cache"] = cacheInfo
	}

	return status
}

// collectVMEvents collects and broadcasts VM events
func (em *EventManager) collectVMEvents() {
	defer em.wg.Done()

	ticker := time.NewTicker(8 * time.Second) // 8-second intervals for VM events
	defer ticker.Stop()

	for {
		select {
		case <-em.ctx.Done():
			return
		case <-ticker.C:
			events := em.getVMEvents()
			em.wsHandler.BroadcastEvent(handlers.EventVMEvents, events)
		}
	}
}

// collectInfrastructureStatus collects and broadcasts infrastructure status
func (em *EventManager) collectInfrastructureStatus() {
	defer em.wg.Done()

	ticker := time.NewTicker(15 * time.Second) // 15-second intervals for infrastructure
	defer ticker.Stop()

	for {
		select {
		case <-em.ctx.Done():
			return
		case <-ticker.C:
			status := em.getInfrastructureStatus()
			em.wsHandler.BroadcastEvent(handlers.EventUPSStatus, status["ups"])
			em.wsHandler.BroadcastEvent(handlers.EventFanStatus, status["fans"])
			em.wsHandler.BroadcastEvent(handlers.EventPowerStatus, status["power"])
		}
	}
}

// collectResourceAlerts monitors and broadcasts resource alerts
func (em *EventManager) collectResourceAlerts() {
	defer em.wg.Done()

	ticker := time.NewTicker(5 * time.Second) // 5-second intervals for resource monitoring
	defer ticker.Stop()

	for {
		select {
		case <-em.ctx.Done():
			return
		case <-ticker.C:
			em.checkResourceAlerts()
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
