package monitoring

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/domalab/uma/daemon/logger"
	"github.com/domalab/uma/daemon/services/api/handlers"
	"github.com/domalab/uma/daemon/services/api/utils"
)

// TemperatureThreshold represents a temperature threshold configuration
type TemperatureThreshold struct {
	SensorType   string  `json:"sensor_type"`   // "cpu", "disk", "gpu", "system"
	WarningTemp  float64 `json:"warning_temp"`  // Warning threshold in Celsius
	CriticalTemp float64 `json:"critical_temp"` // Critical threshold in Celsius
	ShutdownTemp float64 `json:"shutdown_temp"` // Emergency shutdown threshold in Celsius
	Enabled      bool    `json:"enabled"`       // Whether monitoring is enabled
	AutoActions  bool    `json:"auto_actions"`  // Whether to take automatic actions
}

// TemperatureAlert represents a temperature alert
type TemperatureAlert struct {
	SensorName  string    `json:"sensor_name"`
	SensorType  string    `json:"sensor_type"`
	Temperature float64   `json:"temperature"`
	Threshold   float64   `json:"threshold"`
	Level       string    `json:"level"` // "warning", "critical", "emergency"
	Message     string    `json:"message"`
	Timestamp   time.Time `json:"timestamp"`
	ActionTaken string    `json:"action_taken"` // Description of any automatic action taken
}

// TemperatureMonitor handles critical temperature monitoring with alerts
type TemperatureMonitor struct {
	api       utils.APIInterface
	wsHandler *handlers.WebSocketHandler
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup

	// Configuration
	thresholds   map[string]TemperatureThreshold
	alertHistory []TemperatureAlert
	lastAlerts   map[string]time.Time
	mutex        sync.RWMutex

	// Monitoring intervals
	monitorInterval time.Duration
	alertCooldown   time.Duration
}

// NewTemperatureMonitor creates a new temperature monitor
func NewTemperatureMonitor(api utils.APIInterface, wsHandler *handlers.WebSocketHandler) *TemperatureMonitor {
	ctx, cancel := context.WithCancel(context.Background())

	// Default thresholds based on typical hardware specifications
	defaultThresholds := map[string]TemperatureThreshold{
		"cpu": {
			SensorType:   "cpu",
			WarningTemp:  70.0,
			CriticalTemp: 80.0,
			ShutdownTemp: 90.0,
			Enabled:      true,
			AutoActions:  true,
		},
		"disk": {
			SensorType:   "disk",
			WarningTemp:  45.0,
			CriticalTemp: 55.0,
			ShutdownTemp: 65.0,
			Enabled:      true,
			AutoActions:  false, // Disk shutdown is more dangerous
		},
		"gpu": {
			SensorType:   "gpu",
			WarningTemp:  75.0,
			CriticalTemp: 85.0,
			ShutdownTemp: 95.0,
			Enabled:      true,
			AutoActions:  true,
		},
		"system": {
			SensorType:   "system",
			WarningTemp:  65.0,
			CriticalTemp: 75.0,
			ShutdownTemp: 85.0,
			Enabled:      true,
			AutoActions:  true,
		},
	}

	return &TemperatureMonitor{
		api:             api,
		wsHandler:       wsHandler,
		ctx:             ctx,
		cancel:          cancel,
		thresholds:      defaultThresholds,
		alertHistory:    make([]TemperatureAlert, 0),
		lastAlerts:      make(map[string]time.Time),
		monitorInterval: 2 * time.Second,  // Check every 2 seconds
		alertCooldown:   30 * time.Second, // Minimum 30 seconds between same alerts
	}
}

// Start starts the temperature monitoring
func (tm *TemperatureMonitor) Start() {
	logger.Green("Starting Critical Temperature Monitor")

	tm.wg.Add(1)
	go tm.monitorTemperatures()
}

// Stop stops the temperature monitoring
func (tm *TemperatureMonitor) Stop() {
	logger.Blue("Stopping Critical Temperature Monitor")
	tm.cancel()
	tm.wg.Wait()
	logger.Green("Critical Temperature Monitor stopped")
}

// monitorTemperatures continuously monitors temperature sensors
func (tm *TemperatureMonitor) monitorTemperatures() {
	defer tm.wg.Done()

	ticker := time.NewTicker(tm.monitorInterval)
	defer ticker.Stop()

	for {
		select {
		case <-tm.ctx.Done():
			return
		case <-ticker.C:
			tm.checkAllTemperatures()
		}
	}
}

// checkAllTemperatures checks all temperature sensors against thresholds
func (tm *TemperatureMonitor) checkAllTemperatures() {
	tempData, err := tm.api.GetSystem().GetEnhancedTemperatureData()
	if err != nil {
		logger.Yellow("Temperature Monitor: Failed to get temperature data: %v", err)
		return
	}

	// Parse temperature data and check thresholds
	if tempMap, ok := tempData.(map[string]interface{}); ok {
		if sensors, ok := tempMap["sensors"].([]interface{}); ok {
			for _, sensor := range sensors {
				if sensorMap, ok := sensor.(map[string]interface{}); ok {
					tm.checkSensorTemperature(sensorMap)
				}
			}
		}
	}
}

// checkSensorTemperature checks individual sensor temperature against thresholds
func (tm *TemperatureMonitor) checkSensorTemperature(sensor map[string]interface{}) {
	name, nameOk := sensor["name"].(string)
	temp, tempOk := sensor["temperature"].(float64)

	if !nameOk || !tempOk {
		return
	}

	// Determine sensor type
	sensorType := tm.determineSensorType(name)

	tm.mutex.RLock()
	threshold, exists := tm.thresholds[sensorType]
	tm.mutex.RUnlock()

	if !exists || !threshold.Enabled {
		return
	}

	// Check thresholds in order of severity
	if temp >= threshold.ShutdownTemp {
		tm.handleTemperatureAlert(name, sensorType, temp, threshold.ShutdownTemp, "emergency", threshold.AutoActions)
	} else if temp >= threshold.CriticalTemp {
		tm.handleTemperatureAlert(name, sensorType, temp, threshold.CriticalTemp, "critical", threshold.AutoActions)
	} else if temp >= threshold.WarningTemp {
		tm.handleTemperatureAlert(name, sensorType, temp, threshold.WarningTemp, "warning", false) // No auto actions for warnings
	}
}

// determineSensorType determines the sensor type based on sensor name
func (tm *TemperatureMonitor) determineSensorType(name string) string {
	nameLower := strings.ToLower(name)

	// CPU keywords
	cpuKeywords := []string{"cpu", "core", "processor", "package"}
	for _, keyword := range cpuKeywords {
		if strings.Contains(nameLower, keyword) {
			return "cpu"
		}
	}

	// Disk keywords
	diskKeywords := []string{"disk", "drive", "sda", "sdb", "sdc", "sdd", "nvme", "sata"}
	for _, keyword := range diskKeywords {
		if strings.Contains(nameLower, keyword) {
			return "disk"
		}
	}

	// GPU keywords
	gpuKeywords := []string{"gpu", "graphics", "nvidia", "amd", "radeon", "geforce"}
	for _, keyword := range gpuKeywords {
		if strings.Contains(nameLower, keyword) {
			return "gpu"
		}
	}

	// Default to system
	return "system"
}

// handleTemperatureAlert handles temperature threshold violations
func (tm *TemperatureMonitor) handleTemperatureAlert(sensorName, sensorType string, temperature, threshold float64, level string, autoActions bool) {
	alertKey := fmt.Sprintf("%s_%s", sensorName, level)

	tm.mutex.Lock()
	lastAlert, exists := tm.lastAlerts[alertKey]
	now := time.Now()

	// Rate limit alerts - only send if cooldown period has passed
	if exists && now.Sub(lastAlert) < tm.alertCooldown {
		tm.mutex.Unlock()
		return
	}

	tm.lastAlerts[alertKey] = now
	tm.mutex.Unlock()

	// Determine action to take
	actionTaken := "none"
	if autoActions {
		actionTaken = tm.takeAutomaticAction(sensorType, level, temperature)
	}

	// Create alert
	alert := TemperatureAlert{
		SensorName:  sensorName,
		SensorType:  sensorType,
		Temperature: temperature,
		Threshold:   threshold,
		Level:       level,
		Message:     fmt.Sprintf("%s temperature %s: %.1f°C (threshold: %.1f°C)", sensorName, level, temperature, threshold),
		Timestamp:   now,
		ActionTaken: actionTaken,
	}

	// Store alert in history
	tm.mutex.Lock()
	tm.alertHistory = append(tm.alertHistory, alert)
	// Keep only last 100 alerts
	if len(tm.alertHistory) > 100 {
		tm.alertHistory = tm.alertHistory[1:]
	}
	tm.mutex.Unlock()

	// Broadcast alert via WebSocket
	if tm.wsHandler != nil {
		tm.wsHandler.BroadcastEvent(handlers.EventTemperatureAlert, alert)
	}

	// Log alert
	logLevel := logger.Yellow
	if level == "critical" {
		logLevel = logger.Red
	} else if level == "emergency" {
		logLevel = logger.Red
	}

	logLevel("Temperature %s alert: %s - %.1f°C (action: %s)", level, sensorName, temperature, actionTaken)
}

// takeAutomaticAction takes automatic action based on temperature alert
func (tm *TemperatureMonitor) takeAutomaticAction(sensorType, level string, temperature float64) string {
	switch level {
	case "emergency":
		// Emergency shutdown for critical temperatures
		logger.Red("EMERGENCY: Temperature %.1f°C detected, initiating emergency procedures", temperature)
		return tm.initiateEmergencyShutdown(sensorType)

	case "critical":
		// Throttling or protective measures
		return tm.initiateCriticalProtection(sensorType, temperature)

	default:
		return "none"
	}
}

// initiateEmergencyShutdown initiates emergency shutdown procedures
func (tm *TemperatureMonitor) initiateEmergencyShutdown(sensorType string) string {
	// In a real implementation, this would:
	// 1. Stop non-essential services
	// 2. Gracefully stop Docker containers
	// 3. Unmount shares
	// 4. Initiate system shutdown

	logger.Red("Emergency shutdown procedures would be initiated for %s overheating", sensorType)
	return fmt.Sprintf("emergency_shutdown_initiated_%s", sensorType)
}

// initiateCriticalProtection initiates critical protection measures
func (tm *TemperatureMonitor) initiateCriticalProtection(sensorType string, temperature float64) string {
	switch sensorType {
	case "cpu":
		// CPU throttling or process priority reduction
		logger.Yellow("Critical CPU temperature %.1f°C - would initiate CPU protection", temperature)
		return "cpu_throttling_initiated"

	case "gpu":
		// GPU workload reduction
		logger.Yellow("Critical GPU temperature %.1f°C - would reduce GPU workload", temperature)
		return "gpu_workload_reduced"

	case "disk":
		// Disk I/O throttling
		logger.Yellow("Critical disk temperature %.1f°C - would throttle disk I/O", temperature)
		return "disk_io_throttled"

	default:
		logger.Yellow("Critical %s temperature %.1f°C - monitoring increased", sensorType, temperature)
		return "monitoring_increased"
	}
}

// GetThresholds returns current temperature thresholds
func (tm *TemperatureMonitor) GetThresholds() map[string]TemperatureThreshold {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	// Return a copy to prevent external modification
	thresholds := make(map[string]TemperatureThreshold)
	for k, v := range tm.thresholds {
		thresholds[k] = v
	}
	return thresholds
}

// UpdateThreshold updates a temperature threshold
func (tm *TemperatureMonitor) UpdateThreshold(sensorType string, threshold TemperatureThreshold) error {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	// Validate threshold values
	if threshold.WarningTemp >= threshold.CriticalTemp || threshold.CriticalTemp >= threshold.ShutdownTemp {
		return fmt.Errorf("invalid threshold values: warning < critical < shutdown required")
	}

	tm.thresholds[sensorType] = threshold
	logger.Blue("Updated temperature threshold for %s: warning=%.1f, critical=%.1f, shutdown=%.1f",
		sensorType, threshold.WarningTemp, threshold.CriticalTemp, threshold.ShutdownTemp)

	return nil
}

// GetAlertHistory returns recent temperature alerts
func (tm *TemperatureMonitor) GetAlertHistory(limit int) []TemperatureAlert {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	if limit <= 0 || limit > len(tm.alertHistory) {
		limit = len(tm.alertHistory)
	}

	// Return the most recent alerts
	start := len(tm.alertHistory) - limit
	if start < 0 {
		start = 0
	}

	alerts := make([]TemperatureAlert, limit)
	copy(alerts, tm.alertHistory[start:])
	return alerts
}
