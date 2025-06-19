package services

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/domalab/uma/daemon/services/api/utils"
)

// HealthService handles health check and diagnostic business logic
type HealthService struct {
	api     utils.APIInterface
	version string
}

// NewHealthService creates a new health service
func NewHealthService(api utils.APIInterface, version string) *HealthService {
	return &HealthService{
		api:     api,
		version: version,
	}
}

// HealthStatus represents the overall health status
type HealthStatus struct {
	Status       string                 `json:"status"`
	Version      string                 `json:"version"`
	Timestamp    string                 `json:"timestamp"`
	Uptime       string                 `json:"uptime"`
	Dependencies map[string]interface{} `json:"dependencies"`
	System       map[string]interface{} `json:"system"`
}

// DependencyStatus represents the status of a dependency
type DependencyStatus struct {
	Name      string `json:"name"`
	Status    string `json:"status"`
	Message   string `json:"message,omitempty"`
	Timestamp string `json:"timestamp"`
}

// GetHealthStatus retrieves comprehensive health status
func (h *HealthService) GetHealthStatus() *HealthStatus {
	timestamp := time.Now().UTC().Format(time.RFC3339)

	// Get dependency statuses
	dependencies := h.checkAllDependencies()

	// Get system health
	systemHealth := h.getSystemHealth()

	// Determine overall status
	overallStatus := h.determineOverallStatus(dependencies, systemHealth)

	return &HealthStatus{
		Status:       overallStatus,
		Version:      h.version,
		Timestamp:    timestamp,
		Uptime:       h.getUptime(),
		Dependencies: dependencies,
		System:       systemHealth,
	}
}

// GetBasicHealthStatus retrieves basic health status
func (h *HealthService) GetBasicHealthStatus() map[string]interface{} {
	return map[string]interface{}{
		"status":    "healthy",
		"version":   h.version,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"uptime":    h.getUptime(),
	}
}

// checkAllDependencies checks the status of all dependencies
func (h *HealthService) checkAllDependencies() map[string]interface{} {
	dependencies := make(map[string]interface{})

	// Check Docker
	dependencies["docker"] = h.checkDockerHealth()

	// Check libvirt/VMs
	dependencies["libvirt"] = h.checkLibvirtHealth()

	// Check storage/array
	dependencies["storage"] = h.checkStorageHealth()

	// Check system services
	dependencies["system_services"] = h.checkSystemServicesHealth()

	return dependencies
}

// checkDockerHealth checks Docker daemon health
func (h *HealthService) checkDockerHealth() DependencyStatus {
	status := DependencyStatus{
		Name:      "docker",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	// Try to get Docker info through API
	if h.api.GetDocker() != nil {
		_, err := h.api.GetDocker().GetSystemInfo()
		if err != nil {
			status.Status = "unhealthy"
			status.Message = fmt.Sprintf("Docker API error: %v", err)
		} else {
			status.Status = "healthy"
			status.Message = "Docker daemon is running and accessible"
		}
	} else {
		status.Status = "unavailable"
		status.Message = "Docker interface not available"
	}

	return status
}

// checkLibvirtHealth checks libvirt/VM health
func (h *HealthService) checkLibvirtHealth() DependencyStatus {
	status := DependencyStatus{
		Name:      "libvirt",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	if h.api.GetVM() == nil {
		status.Status = "unavailable"
		status.Message = "VM interface not available"
		return status
	}

	// Try to get VMs to test connection
	_, err := h.api.GetVM().GetVMs()
	if err != nil {
		if strings.Contains(err.Error(), "connection") {
			status.Status = "disconnected"
			status.Message = "libvirt connection failed"
		} else {
			status.Status = "error"
			status.Message = fmt.Sprintf("libvirt error: %v", err)
		}
	} else {
		status.Status = "healthy"
		status.Message = "libvirt is running and accessible"
	}

	return status
}

// checkStorageHealth checks storage/array health
func (h *HealthService) checkStorageHealth() DependencyStatus {
	status := DependencyStatus{
		Name:      "storage",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	if h.api.GetStorage() == nil {
		status.Status = "unavailable"
		status.Message = "Storage interface not available"
		return status
	}

	// Try to get array info
	_, err := h.api.GetStorage().GetArrayInfo()
	if err != nil {
		status.Status = "error"
		status.Message = fmt.Sprintf("Storage error: %v", err)
	} else {
		status.Status = "healthy"
		status.Message = "Storage system is accessible"
	}

	return status
}

// checkSystemServicesHealth checks critical system services
func (h *HealthService) checkSystemServicesHealth() map[string]interface{} {
	services := map[string]interface{}{}

	// Check critical services
	criticalServices := []string{
		"sshd",
		"nginx",
		"smbd",
		"nmbd",
	}

	for _, service := range criticalServices {
		services[service] = h.checkSystemService(service)
	}

	return services
}

// checkSystemService checks if a system service is running
func (h *HealthService) checkSystemService(serviceName string) DependencyStatus {
	status := DependencyStatus{
		Name:      serviceName,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	// Check service status using systemctl
	cmd := exec.Command("systemctl", "is-active", serviceName)
	output, err := cmd.Output()

	if err != nil {
		status.Status = "unknown"
		status.Message = "Unable to check service status"
	} else {
		serviceStatus := strings.TrimSpace(string(output))
		if serviceStatus == "active" {
			status.Status = "healthy"
			status.Message = "Service is running"
		} else {
			status.Status = "unhealthy"
			status.Message = fmt.Sprintf("Service status: %s", serviceStatus)
		}
	}

	return status
}

// getSystemHealth retrieves system-level health information
func (h *HealthService) getSystemHealth() map[string]interface{} {
	return map[string]interface{}{
		"load_average":   h.getLoadAverage(),
		"memory_usage":   h.getMemoryUsage(),
		"disk_usage":     h.getDiskUsage(),
		"network_status": h.getNetworkStatus(),
		"temperature":    h.getTemperatureStatus(),
	}
}

// getLoadAverage gets system load average
func (h *HealthService) getLoadAverage() map[string]interface{} {
	// Read load average from /proc/loadavg
	content, err := os.ReadFile("/proc/loadavg")
	if err != nil {
		return map[string]interface{}{
			"status": "error",
			"error":  err.Error(),
		}
	}

	fields := strings.Fields(string(content))
	if len(fields) < 3 {
		return map[string]interface{}{
			"status": "error",
			"error":  "invalid loadavg format",
		}
	}

	return map[string]interface{}{
		"status": "ok",
		"1min":   fields[0],
		"5min":   fields[1],
		"15min":  fields[2],
	}
}

// getMemoryUsage gets memory usage information
func (h *HealthService) getMemoryUsage() map[string]interface{} {
	if h.api.GetSystem() != nil {
		memInfo, err := h.api.GetSystem().GetMemoryInfo()
		if err == nil {
			return map[string]interface{}{
				"status": "ok",
				"data":   memInfo,
			}
		}
	}

	return map[string]interface{}{
		"status": "error",
		"error":  "unable to get memory info",
	}
}

// getDiskUsage gets disk usage information
func (h *HealthService) getDiskUsage() map[string]interface{} {
	// Check critical mount points
	mountPoints := []string{"/", "/boot", "/var/log"}
	usage := make(map[string]interface{})

	for _, mount := range mountPoints {
		cmd := exec.Command("df", "-h", mount)
		output, err := cmd.Output()
		if err != nil {
			usage[mount] = map[string]interface{}{
				"status": "error",
				"error":  err.Error(),
			}
		} else {
			lines := strings.Split(string(output), "\n")
			if len(lines) >= 2 {
				fields := strings.Fields(lines[1])
				if len(fields) >= 5 {
					usage[mount] = map[string]interface{}{
						"status":      "ok",
						"filesystem":  fields[0],
						"size":        fields[1],
						"used":        fields[2],
						"available":   fields[3],
						"use_percent": fields[4],
					}
				}
			}
		}
	}

	return usage
}

// getNetworkStatus gets network connectivity status
func (h *HealthService) getNetworkStatus() map[string]interface{} {
	if h.api.GetSystem() != nil {
		netInfo, err := h.api.GetSystem().GetNetworkInfo()
		if err == nil {
			return map[string]interface{}{
				"status": "ok",
				"data":   netInfo,
			}
		}
	}

	return map[string]interface{}{
		"status": "error",
		"error":  "unable to get network info",
	}
}

// getTemperatureStatus gets temperature status
func (h *HealthService) getTemperatureStatus() map[string]interface{} {
	if h.api.GetSystem() != nil {
		tempInfo, err := h.api.GetSystem().GetEnhancedTemperatureData()
		if err == nil {
			return map[string]interface{}{
				"status": "ok",
				"data":   tempInfo,
			}
		}
	}

	return map[string]interface{}{
		"status": "error",
		"error":  "unable to get temperature info",
	}
}

// determineOverallStatus determines the overall health status
func (h *HealthService) determineOverallStatus(dependencies map[string]interface{}, system map[string]interface{}) string {
	// Check for any critical failures
	for _, dep := range dependencies {
		if depStatus, ok := dep.(DependencyStatus); ok {
			if depStatus.Status == "unhealthy" || depStatus.Status == "error" {
				return "unhealthy"
			}
		}
	}

	// Check system health
	for _, sysHealth := range system {
		if healthMap, ok := sysHealth.(map[string]interface{}); ok {
			if status, ok := healthMap["status"].(string); ok && status == "error" {
				return "degraded"
			}
		}
	}

	return "healthy"
}

// getUptime gets system uptime
func (h *HealthService) getUptime() string {
	if h.api.GetSystem() != nil {
		uptimeInfo, err := h.api.GetSystem().GetUptimeInfo()
		if err == nil {
			if uptimeMap, ok := uptimeInfo.(map[string]interface{}); ok {
				if uptime, ok := uptimeMap["uptime"].(string); ok {
					return uptime
				}
			}
		}
	}

	// Fallback: read from /proc/uptime
	content, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return "unknown"
	}

	fields := strings.Fields(string(content))
	if len(fields) > 0 {
		return fields[0] + " seconds"
	}

	return "unknown"
}
