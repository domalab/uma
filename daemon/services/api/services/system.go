package services

import (
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/domalab/uma/daemon/logger"
	"github.com/domalab/uma/daemon/services/api/utils"
)

// SystemService handles system-related business logic
type SystemService struct {
	api utils.APIInterface
}

// NewSystemService creates a new system service
func NewSystemService(api utils.APIInterface) *SystemService {
	return &SystemService{
		api: api,
	}
}

// GetCPUData retrieves CPU information and metrics
func (s *SystemService) GetCPUData() map[string]interface{} {
	timestamp := time.Now().UTC().Format(time.RFC3339)

	cpuInfo, err := s.api.GetSystem().GetCPUInfo()
	if err != nil {
		logger.Red("Failed to get CPU info: %v", err)
		return map[string]interface{}{
			"usage":        0.0,
			"temperature":  0.0,
			"cores":        0,
			"model":        "Unknown",
			"last_updated": timestamp,
		}
	}

	// Convert to map if needed
	if cpuMap, ok := cpuInfo.(map[string]interface{}); ok {
		cpuMap["last_updated"] = timestamp
		return cpuMap
	}

	return map[string]interface{}{
		"data":         cpuInfo,
		"last_updated": timestamp,
	}
}

// GetMemoryData retrieves memory information and metrics
func (s *SystemService) GetMemoryData() map[string]interface{} {
	timestamp := time.Now().UTC().Format(time.RFC3339)

	memInfo, err := s.api.GetSystem().GetMemoryInfo()
	if err != nil {
		logger.Red("Failed to get memory info: %v", err)
		return map[string]interface{}{
			"total":        0,
			"used":         0,
			"free":         0,
			"available":    0,
			"usage":        0.0,
			"last_updated": timestamp,
		}
	}

	// Convert to map if needed
	if memMap, ok := memInfo.(map[string]interface{}); ok {
		memMap["last_updated"] = timestamp
		return memMap
	}

	return map[string]interface{}{
		"data":         memInfo,
		"last_updated": timestamp,
	}
}

// GetTemperatureData retrieves temperature sensor information
func (s *SystemService) GetTemperatureData() map[string]interface{} {
	// Get enhanced temperature data
	enhancedData, err := s.api.GetSystem().GetEnhancedTemperatureData()
	if err != nil {
		logger.Yellow("Enhanced temperature data not available, falling back to basic: %v", err)
		return s.getBasicTemperatureData()
	}

	// Convert to map if needed
	if tempMap, ok := enhancedData.(map[string]interface{}); ok {
		tempMap["last_updated"] = time.Now().UTC().Format(time.RFC3339)
		// Add overall_status if missing
		if _, exists := tempMap["overall_status"]; !exists {
			tempMap["overall_status"] = s.calculateOverallTemperatureStatus(tempMap)
		}
		return tempMap
	}

	return map[string]interface{}{
		"data":           enhancedData,
		"last_updated":   time.Now().UTC().Format(time.RFC3339),
		"overall_status": "normal", // Default overall status
	}
}

// getBasicTemperatureData provides fallback temperature data
func (s *SystemService) getBasicTemperatureData() map[string]interface{} {
	sensors := make([]map[string]interface{}, 0)

	// Get CPU temperature from system plugin
	cpuTemp := s.getCPUTemperature()
	if cpuTemp > 0 {
		sensors = append(sensors, map[string]interface{}{
			"name":        "CPU",
			"temperature": cpuTemp,
			"unit":        "°C",
			"status":      s.getTemperatureStatus(cpuTemp),
		})
	}

	// Calculate overall status based on sensors
	overallStatus := "normal"
	for _, sensor := range sensors {
		if status, exists := sensor["status"]; exists {
			if statusStr, ok := status.(string); ok {
				if statusStr == "critical" {
					overallStatus = "critical"
					break
				} else if statusStr == "warning" && overallStatus != "critical" {
					overallStatus = "warm"
				}
			}
		}
	}

	return map[string]interface{}{
		"sensors":        sensors,
		"fans":           []map[string]interface{}{},
		"overall_status": overallStatus,
		"last_updated":   time.Now().UTC().Format(time.RFC3339),
	}
}

// getCPUTemperature gets CPU temperature from system files
func (s *SystemService) getCPUTemperature() float64 {
	// Try different thermal zone files
	thermalPaths := []string{
		"/sys/class/thermal/thermal_zone0/temp",
		"/sys/class/thermal/thermal_zone1/temp",
		"/sys/class/hwmon/hwmon0/temp1_input",
		"/sys/class/hwmon/hwmon1/temp1_input",
	}

	for _, path := range thermalPaths {
		if content, err := os.ReadFile(path); err == nil {
			if temp, err := strconv.ParseFloat(strings.TrimSpace(string(content)), 64); err == nil {
				// Convert millidegrees to degrees if needed
				if temp > 1000 {
					temp = temp / 1000
				}
				return temp
			}
		}
	}

	return 0
}

// getTemperatureStatus determines temperature status based on value
func (s *SystemService) getTemperatureStatus(temp float64) string {
	if temp > 80 {
		return "critical"
	} else if temp > 70 {
		return "warning"
	} else if temp > 60 {
		return "warm"
	}
	return "normal"
}

// calculateOverallTemperatureStatus calculates overall temperature status from sensor data
func (s *SystemService) calculateOverallTemperatureStatus(tempData map[string]interface{}) string {
	overallStatus := "normal"

	// Check sensors array if present
	if sensors, exists := tempData["sensors"]; exists {
		if sensorsArray, ok := sensors.([]interface{}); ok {
			for _, sensor := range sensorsArray {
				if sensorMap, ok := sensor.(map[string]interface{}); ok {
					if status, exists := sensorMap["status"]; exists {
						if statusStr, ok := status.(string); ok {
							if statusStr == "critical" {
								return "critical" // Critical takes precedence
							} else if statusStr == "warning" && overallStatus != "critical" {
								overallStatus = "warm"
							}
						}
					}
				}
			}
		}
	}

	return overallStatus
}

// GetNetworkData retrieves network interface information
func (s *SystemService) GetNetworkData() map[string]interface{} {
	networkInfo, err := s.api.GetSystem().GetNetworkInfo()
	if err != nil {
		logger.Red("Failed to get network info: %v", err)
		return map[string]interface{}{
			"interfaces":   []interface{}{},
			"last_updated": time.Now().UTC().Format(time.RFC3339),
		}
	}

	// Convert to map if needed
	if netMap, ok := networkInfo.(map[string]interface{}); ok {
		netMap["last_updated"] = time.Now().UTC().Format(time.RFC3339)
		return netMap
	}

	return map[string]interface{}{
		"data":         networkInfo,
		"last_updated": time.Now().UTC().Format(time.RFC3339),
	}
}

// GetUPSData retrieves UPS status information
func (s *SystemService) GetUPSData() map[string]interface{} {
	// Check if UPS is available through auto-detection
	upsDetector := s.api.GetUPSDetector()
	if upsDetector == nil || !upsDetector.IsAvailable() {
		// Return "not available" response when UPS is not detected
		return map[string]interface{}{
			"available":      false,
			"status":         "not_detected",
			"battery_charge": 0,
			"runtime":        0,
			"load":           0,
			"voltage":        0,
			"detection":      upsDetector.GetStatus(),
			"last_updated":   time.Now().UTC().Format(time.RFC3339),
		}
	}

	// UPS is available, try to get real data
	// Try to get real UPS data from apcupsd first
	if apcData := s.GetAPCUPSData(); apcData != nil {
		// Add detection info to the response
		apcData["available"] = true
		apcData["detection"] = upsDetector.GetStatus()
		return apcData
	}

	// Try NUT (Network UPS Tools) as fallback
	if nutData := s.GetNUTUPSData(); nutData != nil {
		// Add detection info to the response
		nutData["available"] = true
		nutData["detection"] = upsDetector.GetStatus()
		return nutData
	}

	// UPS detected but communication failed
	return map[string]interface{}{
		"available":      true,
		"status":         "communication_error",
		"battery_charge": 0,
		"runtime":        0,
		"load":           0,
		"voltage":        0,
		"detection":      upsDetector.GetStatus(),
		"last_updated":   time.Now().UTC().Format(time.RFC3339),
	}
}

// GetAPCUPSData retrieves UPS data from apcupsd daemon
func (s *SystemService) GetAPCUPSData() map[string]interface{} {
	// Execute apcaccess command to get UPS status
	cmd := exec.Command("apcaccess")
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	upsData := map[string]interface{}{
		"status":            "unknown",
		"battery_charge":    0,
		"runtime":           0,
		"load":              0,
		"voltage":           0,
		"power_consumption": 0,
		"nominal_power":     0,
		"last_updated":      time.Now().UTC().Format(time.RFC3339),
	}

	var nominalPower float64
	var loadPercent float64

	// Parse apcaccess output
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "STATUS":
			upsData["status"] = strings.ToLower(value)
		case "BCHARGE":
			if charge, err := strconv.ParseFloat(strings.TrimSuffix(value, " Percent"), 64); err == nil {
				upsData["battery_charge"] = charge
			}
		case "TIMELEFT":
			if runtime, err := strconv.ParseFloat(strings.TrimSuffix(value, " Minutes"), 64); err == nil {
				upsData["runtime"] = runtime
			}
		case "LOADPCT":
			if load, err := strconv.ParseFloat(strings.TrimSuffix(value, " Percent"), 64); err == nil {
				upsData["load"] = load
				loadPercent = load
			}
		case "LINEV":
			if voltage, err := strconv.ParseFloat(strings.TrimSuffix(value, " Volts"), 64); err == nil {
				upsData["voltage"] = voltage
			}
		case "NOMPOWER":
			if power, err := strconv.ParseFloat(strings.TrimSuffix(value, " Watts"), 64); err == nil {
				upsData["nominal_power"] = power
				nominalPower = power
			}
		}
	}

	// Calculate real power consumption: nominal_power * load_percent / 100
	if nominalPower > 0 && loadPercent >= 0 {
		powerConsumption := nominalPower * loadPercent / 100
		upsData["power_consumption"] = powerConsumption
	}

	return upsData
}

// GetNUTUPSData retrieves UPS data from NUT daemon
func (s *SystemService) GetNUTUPSData() map[string]interface{} {
	// Execute upsc command to get UPS status
	cmd := exec.Command("upsc", "ups")
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	upsData := map[string]interface{}{
		"status":            "unknown",
		"battery_charge":    0,
		"runtime":           0,
		"load":              0,
		"voltage":           0,
		"power_consumption": 0,
		"nominal_power":     0,
		"last_updated":      time.Now().UTC().Format(time.RFC3339),
	}

	var nominalPower float64
	var loadPercent float64

	// Parse upsc output
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "ups.status":
			upsData["status"] = strings.ToLower(value)
		case "battery.charge":
			if charge, err := strconv.ParseFloat(value, 64); err == nil {
				upsData["battery_charge"] = charge
			}
		case "battery.runtime":
			if runtime, err := strconv.ParseFloat(value, 64); err == nil {
				upsData["runtime"] = runtime / 60 // Convert seconds to minutes
			}
		case "ups.load":
			if load, err := strconv.ParseFloat(value, 64); err == nil {
				upsData["load"] = load
				loadPercent = load
			}
		case "input.voltage":
			if voltage, err := strconv.ParseFloat(value, 64); err == nil {
				upsData["voltage"] = voltage
			}
		case "ups.realpower.nominal", "ups.power.nominal":
			if power, err := strconv.ParseFloat(value, 64); err == nil {
				upsData["nominal_power"] = power
				nominalPower = power
			}
		}
	}

	// Calculate real power consumption: nominal_power * load_percent / 100
	if nominalPower > 0 && loadPercent >= 0 {
		powerConsumption := nominalPower * loadPercent / 100
		upsData["power_consumption"] = powerConsumption
	}

	return upsData
}

// GetFilesystemData retrieves filesystem usage information
func (s *SystemService) GetFilesystemData() map[string]interface{} {
	result := make(map[string]interface{})

	// Docker vDisk usage
	result["docker"] = s.getDockerVDiskUsage()

	// Log filesystem usage
	result["logs"] = s.getLogFilesystemUsage()

	// Boot filesystem usage
	result["boot"] = s.getBootUsage()

	result["last_updated"] = time.Now().UTC().Format(time.RFC3339)

	return result
}

// getDockerVDiskUsage gets Docker vDisk usage information
func (s *SystemService) getDockerVDiskUsage() map[string]interface{} {
	// Check common Docker vDisk locations
	dockerPaths := []string{"/var/lib/docker", "/mnt/user/system/docker/docker.img"}

	for _, path := range dockerPaths {
		if usage := s.getPathUsage(path); usage["total"].(int64) > 0 {
			return usage
		}
	}

	return map[string]interface{}{
		"total": int64(0),
		"used":  int64(0),
		"free":  int64(0),
		"usage": 0.0,
	}
}

// getLogFilesystemUsage gets log filesystem usage
func (s *SystemService) getLogFilesystemUsage() map[string]interface{} {
	return s.getPathUsage("/var/log")
}

// getBootUsage gets boot filesystem usage
func (s *SystemService) getBootUsage() map[string]interface{} {
	return s.getPathUsage("/boot")
}

// getPathUsage gets filesystem usage for a specific path
func (s *SystemService) getPathUsage(path string) map[string]interface{} {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return map[string]interface{}{
			"total": int64(0),
			"used":  int64(0),
			"free":  int64(0),
			"usage": 0.0,
		}
	}

	total := int64(stat.Blocks) * int64(stat.Bsize)
	free := int64(stat.Bavail) * int64(stat.Bsize)
	used := total - free
	usage := 0.0
	if total > 0 {
		usage = float64(used) / float64(total) * 100
	}

	return map[string]interface{}{
		"total": total,
		"used":  used,
		"free":  free,
		"usage": usage,
	}
}
