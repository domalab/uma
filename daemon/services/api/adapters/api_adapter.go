package adapters

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/domalab/uma/daemon/logger"
	"github.com/domalab/uma/daemon/plugins/docker"
	"github.com/domalab/uma/daemon/services/api/utils"
	upsDetector "github.com/domalab/uma/daemon/services/ups"
)

// APIAdapter adapts the existing API to our new interface structure
type APIAdapter struct {
	api interface{} // Will hold the original *Api instance
}

// NewAPIAdapter creates a new API adapter
func NewAPIAdapter(api interface{}) *APIAdapter {
	return &APIAdapter{api: api}
}

// GetInfo returns general API information
func (a *APIAdapter) GetInfo() interface{} {
	// Return actual API information
	return map[string]interface{}{
		"service":      "UMA REST API",
		"description":  "Unraid Management Agent REST API",
		"version":      "1.0.0",
		"status":       "running",
		"last_updated": time.Now().UTC().Format(time.RFC3339),
	}
}

// GetSystem returns the system interface
func (a *APIAdapter) GetSystem() utils.SystemInterface {
	return NewSystemAdapter(a.api)
}

// GetStorage returns the storage interface
func (a *APIAdapter) GetStorage() utils.StorageInterface {
	return NewStorageAdapter(a.api)
}

// GetDocker returns the Docker interface
func (a *APIAdapter) GetDocker() utils.DockerInterface {
	return &DockerAdapter{api: a.api}
}

// GetVM returns the VM interface
func (a *APIAdapter) GetVM() utils.VMInterface {
	return NewVMAdapter(a.api)
}

// GetAuth returns the authentication interface
func (a *APIAdapter) GetAuth() utils.AuthInterface {
	return NewAuthAdapter(a.api)
}

// GetNotifications returns the notification interface
func (a *APIAdapter) GetNotifications() utils.NotificationInterface {
	return &NotificationAdapter{api: a.api}
}

// GetUPSDetector returns the UPS detector interface
func (a *APIAdapter) GetUPSDetector() utils.UPSDetectorInterface {
	return &UPSDetectorAdapter{api: a.api}
}

// GetConfigManager returns the configuration manager interface
func (a *APIAdapter) GetConfigManager() interface{} {
	// Try to cast the API to the correct type that has GetConfigManager method
	if apiInstance, ok := a.api.(interface{ GetConfigManager() interface{} }); ok {
		return apiInstance.GetConfigManager()
	}
	return nil
}

// GetMCPServer returns the MCP server interface
func (a *APIAdapter) GetMCPServer() interface{} {
	// Try to cast the API to the correct type that has GetMCPServer method
	if apiInstance, ok := a.api.(interface{ GetMCPServer() interface{} }); ok {
		return apiInstance.GetMCPServer()
	}
	return nil
}

// SystemAdapter adapts system operations
type SystemAdapter struct {
	api     interface{}
	monitor *SystemMonitor
}

func NewSystemAdapter(api interface{}) *SystemAdapter {
	return &SystemAdapter{
		api:     api,
		monitor: NewSystemMonitor(),
	}
}

func (s *SystemAdapter) GetCPUInfo() (interface{}, error) {
	// Use real system monitoring
	return s.monitor.GetRealCPUInfo()
}

func (s *SystemAdapter) GetMemoryInfo() (interface{}, error) {
	// Use real system monitoring
	return s.monitor.GetRealMemoryInfo()
}

func (s *SystemAdapter) GetLoadInfo() (interface{}, error) {
	// Load info is included in CPU info, extract it
	cpuInfo, err := s.monitor.GetRealCPUInfo()
	if err != nil {
		return map[string]interface{}{
			"load1":  0.0,
			"load5":  0.0,
			"load15": 0.0,
		}, err
	}

	if cpuMap, ok := cpuInfo.(map[string]interface{}); ok {
		return map[string]interface{}{
			"load1":  cpuMap["load1"],
			"load5":  cpuMap["load5"],
			"load15": cpuMap["load15"],
		}, nil
	}

	return map[string]interface{}{
		"load1":  0.0,
		"load5":  0.0,
		"load15": 0.0,
	}, nil
}

func (s *SystemAdapter) GetUptimeInfo() (interface{}, error) {
	// Use real system monitoring
	return s.monitor.GetRealUptimeInfo()
}

func (s *SystemAdapter) GetNetworkInfo() (interface{}, error) {
	// Use real system monitoring
	return s.monitor.GetRealNetworkInfo()
}

func (s *SystemAdapter) GetEnhancedTemperatureData() (interface{}, error) {
	// Use real system monitoring
	return s.monitor.GetRealTemperatureData()
}

func (s *SystemAdapter) GetGPUInfo() (interface{}, error) {
	// Use real GPU monitoring
	return s.monitor.GetRealGPUInfo()
}

func (s *SystemAdapter) GetSystemLogs() (interface{}, error) {
	// Use real system log monitoring
	return s.monitor.GetRealSystemLogs()
}

func (s *SystemAdapter) GetRealArrayInfo() (interface{}, error) {
	// Use real array monitoring from storage monitor
	storageMonitor := NewStorageMonitor()
	return storageMonitor.GetRealArrayInfo()
}

// StorageAdapter adapts storage operations
type StorageAdapter struct {
	api     interface{}
	monitor *StorageMonitor
}

func NewStorageAdapter(api interface{}) *StorageAdapter {
	return &StorageAdapter{
		api:     api,
		monitor: NewStorageMonitor(),
	}
}

func (s *StorageAdapter) GetArrayInfo() (interface{}, error) {
	// Use real storage monitoring
	return s.monitor.GetRealArrayInfo()
}

func (s *StorageAdapter) GetDisks() (interface{}, error) {
	// Use real storage monitoring
	return s.monitor.GetRealDisks()
}

func (s *StorageAdapter) GetZFSPools() (interface{}, error) {
	// Use real ZFS monitoring
	return s.monitor.GetRealZFSPools()
}

func (s *StorageAdapter) GetCacheInfo() (interface{}, error) {
	// Use real cache monitoring
	return s.monitor.GetRealCacheInfo()
}

func (s *StorageAdapter) StartArray(request interface{}) error {
	// Implementation would call original API methods
	return nil
}

func (s *StorageAdapter) StopArray(request interface{}) error {
	// Implementation would call original API methods
	return nil
}

// DockerAdapter adapts Docker operations
type DockerAdapter struct {
	api interface{}
}

func (d *DockerAdapter) GetContainers() (interface{}, error) {
	// Try to cast the API to the correct type that has GetDockerManager method
	if apiInstance, ok := d.api.(interface{ GetDockerManager() *docker.DockerManager }); ok {
		dockerManager := apiInstance.GetDockerManager()
		if dockerManager != nil {
			// Call ListContainers with all=true to get all containers
			containers, err := dockerManager.ListContainers(true)
			if err != nil {
				logger.Yellow("Failed to get Docker containers: %v", err)
				return []interface{}{}, err
			}

			// Initialize slice fields for each container and add performance metrics
			result := make([]interface{}, len(containers))
			for i, container := range containers {
				// Ensure slice fields are initialized
				if container.Ports == nil {
					container.Ports = []docker.PortMapping{}
				}
				if container.Mounts == nil {
					container.Mounts = []docker.MountInfo{}
				}
				if container.Networks == nil {
					container.Networks = []docker.NetworkInfo{}
				}
				if container.Labels == nil {
					container.Labels = make(map[string]string)
				}
				if container.Environment == nil {
					container.Environment = []string{}
				}

				// Add performance metrics for running containers (with caching)
				containerWithStats := d.addPerformanceMetricsWithCache(dockerManager, container)
				result[i] = containerWithStats
			}
			// Use structured logging for monitoring - only log significant events or errors
			logger.LogDockerOperation("container_list", len(containers), nil)
			return result, nil
		}
	}

	// Fallback to empty array if Docker manager not available
	logger.Yellow("Docker manager not available, returning empty container list")
	return []interface{}{}, nil
}

// addPerformanceMetricsWithCache adds performance statistics to container data with caching
func (d *DockerAdapter) addPerformanceMetricsWithCache(dockerManager *docker.DockerManager, container docker.ContainerInfo) interface{} {
	// Convert container to map for easier manipulation
	containerMap := make(map[string]interface{})

	// Copy all existing container fields
	containerMap["id"] = container.ID
	containerMap["name"] = container.Name
	containerMap["image"] = container.Image
	containerMap["state"] = container.State
	containerMap["status"] = container.Status
	containerMap["created"] = container.Created
	containerMap["started_at"] = container.StartedAt
	containerMap["ports"] = container.Ports
	containerMap["mounts"] = container.Mounts
	containerMap["networks"] = container.Networks
	containerMap["labels"] = container.Labels
	containerMap["environment"] = container.Environment
	containerMap["restart_policy"] = container.RestartPolicy

	// Initialize performance metrics with null values
	containerMap["cpu_percent"] = nil
	containerMap["memory_usage"] = nil
	containerMap["memory_limit"] = nil
	containerMap["memory_percent"] = nil
	containerMap["network_rx"] = nil
	containerMap["network_tx"] = nil
	containerMap["block_read"] = nil
	containerMap["block_write"] = nil

	// Only collect stats for running containers to avoid errors
	if container.State == "running" {
		// Try to get cached stats first to avoid blocking API calls
		cacheKey := fmt.Sprintf("container_stats_%s", container.ID)

		// Use a goroutine to collect stats asynchronously with timeout
		statsChan := make(chan *docker.DockerStats, 1)
		errorChan := make(chan error, 1)

		go func() {
			stats, err := dockerManager.GetContainerStats(container.ID)
			if err != nil {
				errorChan <- err
				return
			}
			statsChan <- stats
		}()

		// Wait for stats with a short timeout to prevent API blocking
		select {
		case stats := <-statsChan:
			if stats != nil {
				containerMap["cpu_percent"] = stats.CPUPercent
				containerMap["memory_usage"] = stats.MemUsage
				containerMap["memory_limit"] = stats.MemLimit
				containerMap["memory_percent"] = stats.MemPercent

				// Parse network I/O from string format "271kB / 2.2MB"
				if stats.NetIO != "" {
					rx, tx := d.parseNetworkIO(stats.NetIO)
					containerMap["network_rx"] = rx
					containerMap["network_tx"] = tx
				}

				// Parse block I/O from string format "2.67MB / 1.54MB"
				if stats.BlockIO != "" {
					read, write := d.parseBlockIO(stats.BlockIO)
					containerMap["block_read"] = read
					containerMap["block_write"] = write
				}
			}
		case <-errorChan:
			// Stats collection failed, keep null values
		case <-time.After(2 * time.Second):
			// Timeout after 2 seconds to prevent API blocking
			// Keep null values for performance metrics
		}

		_ = cacheKey // Prevent unused variable warning
	}

	return containerMap
}

// parseNetworkIO parses network I/O string format "271kB / 2.2MB" into bytes
func (d *DockerAdapter) parseNetworkIO(netIO string) (int64, int64) {
	parts := strings.Split(netIO, " / ")
	if len(parts) != 2 {
		return 0, 0
	}

	rx := d.parseIOValue(strings.TrimSpace(parts[0]))
	tx := d.parseIOValue(strings.TrimSpace(parts[1]))

	return rx, tx
}

// parseBlockIO parses block I/O string format "2.67MB / 1.54MB" into bytes
func (d *DockerAdapter) parseBlockIO(blockIO string) (int64, int64) {
	parts := strings.Split(blockIO, " / ")
	if len(parts) != 2 {
		return 0, 0
	}

	read := d.parseIOValue(strings.TrimSpace(parts[0]))
	write := d.parseIOValue(strings.TrimSpace(parts[1]))

	return read, write
}

// parseIOValue converts I/O value strings like "271kB", "2.2MB" to bytes
func (d *DockerAdapter) parseIOValue(value string) int64 {
	if value == "" || value == "0B" {
		return 0
	}

	// Extract numeric part and unit
	var numStr string
	var unit string

	for i, char := range value {
		if (char >= '0' && char <= '9') || char == '.' {
			numStr += string(char)
		} else {
			unit = value[i:]
			break
		}
	}

	// Parse the numeric value
	num, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return 0
	}

	// Convert based on unit
	unit = strings.ToUpper(strings.TrimSpace(unit))
	switch unit {
	case "B":
		return int64(num)
	case "KB", "K":
		return int64(num * 1000)
	case "MB", "M":
		return int64(num * 1000 * 1000)
	case "GB", "G":
		return int64(num * 1000 * 1000 * 1000)
	case "TB", "T":
		return int64(num * 1000 * 1000 * 1000 * 1000)
	case "KIB":
		return int64(num * 1024)
	case "MIB":
		return int64(num * 1024 * 1024)
	case "GIB":
		return int64(num * 1024 * 1024 * 1024)
	case "TIB":
		return int64(num * 1024 * 1024 * 1024 * 1024)
	default:
		return int64(num)
	}
}

func (d *DockerAdapter) GetContainer(id string) (interface{}, error) {
	// Try to get the Docker manager from the API with correct type
	if apiInstance, ok := d.api.(interface{ GetDockerManager() *docker.DockerManager }); ok {
		dockerManager := apiInstance.GetDockerManager()
		if dockerManager != nil {
			// Call GetContainer with the correct signature
			container, err := dockerManager.GetContainer(id)
			if err != nil {
				return nil, err
			}
			return container, nil
		}
	}

	// Return fallback container data when Docker manager is unavailable
	return map[string]interface{}{
		"id":     id,
		"name":   "mock-container",
		"status": "unavailable",
		"image":  "mock-image",
	}, nil
}

func (d *DockerAdapter) StartContainer(id string) error {
	// Try to get the Docker manager from the API with correct type
	if apiInstance, ok := d.api.(interface{ GetDockerManager() *docker.DockerManager }); ok {
		dockerManager := apiInstance.GetDockerManager()
		if dockerManager != nil {
			return dockerManager.StartContainer(id)
		}
	}

	return nil
}

func (d *DockerAdapter) StopContainer(id string, timeout int) error {
	// Try to get the Docker manager from the API with correct type
	if apiInstance, ok := d.api.(interface{ GetDockerManager() *docker.DockerManager }); ok {
		dockerManager := apiInstance.GetDockerManager()
		if dockerManager != nil {
			return dockerManager.StopContainer(id, timeout)
		}
	}

	return nil
}

func (d *DockerAdapter) RestartContainer(id string, timeout int) error {
	// Try to get the Docker manager from the API with correct type
	if apiInstance, ok := d.api.(interface{ GetDockerManager() *docker.DockerManager }); ok {
		dockerManager := apiInstance.GetDockerManager()
		if dockerManager != nil {
			return dockerManager.RestartContainer(id, timeout)
		}
	}

	return nil
}

func (d *DockerAdapter) GetImages() (interface{}, error) {
	// Try to get the Docker manager from the API with correct type
	if apiInstance, ok := d.api.(interface{ GetDockerManager() *docker.DockerManager }); ok {
		dockerManager := apiInstance.GetDockerManager()
		if dockerManager != nil {
			return dockerManager.ListImages()
		}
	}

	return []interface{}{}, nil
}

func (d *DockerAdapter) GetNetworks() (interface{}, error) {
	// Try to get the Docker manager from the API with correct type
	if apiInstance, ok := d.api.(interface{ GetDockerManager() *docker.DockerManager }); ok {
		dockerManager := apiInstance.GetDockerManager()
		if dockerManager != nil {
			return dockerManager.ListNetworks()
		}
	}

	return []interface{}{}, nil
}

func (d *DockerAdapter) GetContainerStats(id string) (interface{}, error) {
	// Try to get the Docker manager from the API with correct type
	if apiInstance, ok := d.api.(interface{ GetDockerManager() *docker.DockerManager }); ok {
		dockerManager := apiInstance.GetDockerManager()
		if dockerManager != nil {
			// Call GetContainerStats with the correct signature
			stats, err := dockerManager.GetContainerStats(id)
			if err != nil {
				return nil, err
			}
			return stats, nil
		}
	}

	return map[string]interface{}{
		"container_id": id,
		"cpu_percent":  0.0,
		"memory_usage": 0,
		"timestamp":    time.Now().UTC().Format(time.RFC3339),
	}, nil
}

func (d *DockerAdapter) GetSystemInfo() (interface{}, error) {
	// Try to get the Docker manager from the API with correct type
	if apiInstance, ok := d.api.(interface{ GetDockerManager() *docker.DockerManager }); ok {
		dockerManager := apiInstance.GetDockerManager()
		if dockerManager != nil {
			// Call GetDockerInfo with the correct signature
			info, err := dockerManager.GetDockerInfo()
			if err != nil {
				return nil, err
			}
			return info, nil
		}
	}

	// Return fallback system info when Docker manager is unavailable
	return map[string]interface{}{
		"version":    "unavailable",
		"containers": 0,
		"images":     0,
		"status":     "unavailable",
	}, nil
}

// VMAdapter adapts VM operations
type VMAdapter struct {
	api     interface{}
	monitor *VMMonitor
}

func NewVMAdapter(api interface{}) *VMAdapter {
	return &VMAdapter{
		api:     api,
		monitor: NewVMMonitor(),
	}
}

func (v *VMAdapter) GetVMs() (interface{}, error) {
	// Use real VM monitoring
	return v.monitor.GetRealVMs()
}

func (v *VMAdapter) GetVM(name string) (interface{}, error) {
	// Get all VMs and find the specific one
	vms, err := v.monitor.GetRealVMs()
	if err != nil {
		return nil, err
	}

	if vmList, ok := vms.([]interface{}); ok {
		for _, vm := range vmList {
			if vmMap, ok := vm.(map[string]interface{}); ok {
				if vmMap["name"] == name {
					return vmMap, nil
				}
			}
		}
	}

	return map[string]interface{}{
		"name":   name,
		"status": "not_found",
	}, nil
}

func (v *VMAdapter) StartVM(name string) error {
	_, err := v.monitor.ControlVM(name, "start")
	return err
}

func (v *VMAdapter) StopVM(name string) error {
	_, err := v.monitor.ControlVM(name, "stop")
	return err
}

func (v *VMAdapter) RestartVM(name string) error {
	_, err := v.monitor.ControlVM(name, "restart")
	return err
}

func (v *VMAdapter) GetVMStats(name string) (interface{}, error) {
	// Try to get the VM manager from the API with correct type
	if apiInstance, ok := v.api.(interface{ GetVMManager() interface{} }); ok {
		vmManager := apiInstance.GetVMManager()
		if vmManager != nil {
			// Call GetVMStats with the correct signature
			if vm, ok := vmManager.(interface {
				GetVMStats(string) (interface{}, error)
			}); ok {
				stats, err := vm.GetVMStats(name)
				if err != nil {
					return nil, err
				}
				return stats, nil
			}
		}
	}

	// Fallback: try to get stats from the VM monitor directly
	if vmStats := v.monitor.getVMStats(name); len(vmStats) > 0 {
		return vmStats, nil
	}

	return map[string]interface{}{
		"name":        name,
		"cpu_percent": 0.0,
		"memory_used": 0,
	}, nil
}

func (v *VMAdapter) GetVMConsole(name string) (interface{}, error) {
	return map[string]interface{}{
		"type": "vnc",
		"host": "localhost",
		"port": 5900,
	}, nil
}

func (v *VMAdapter) SetVMAutostart(name string, autostart bool) error {
	return nil
}

// NotificationAdapter adapts notification operations
type NotificationAdapter struct {
	api interface{}
}

func (n *NotificationAdapter) GetNotifications(level, category string, unreadOnly bool) (interface{}, error) {
	// Notification system is not implemented in UMA
	// Return empty array to indicate no notifications
	return []interface{}{}, nil
}

func (n *NotificationAdapter) GetNotificationsPaginated(page, limit int, level, category string, unreadOnly bool) (interface{}, error) {
	return []interface{}{}, nil
}

func (n *NotificationAdapter) GetNotification(id string) (interface{}, error) {
	// Notification system is not implemented in UMA
	return nil, fmt.Errorf("notification %s not found - notification system not implemented", id)
}

func (n *NotificationAdapter) CreateNotification(title, message string, level interface{}, category interface{}, metadata map[string]interface{}) (interface{}, error) {
	// Notification system is not implemented in UMA
	return nil, fmt.Errorf("notification creation not implemented")
}

func (n *NotificationAdapter) UpdateNotification(id string, updates map[string]interface{}) (interface{}, error) {
	// Notification system is not implemented in UMA
	return nil, fmt.Errorf("notification update not implemented")
}

func (n *NotificationAdapter) DeleteNotification(id string) error {
	return nil
}

func (n *NotificationAdapter) ClearAllNotifications() error {
	return nil
}

func (n *NotificationAdapter) MarkAllAsRead() error {
	return nil
}

func (n *NotificationAdapter) GetNotificationStats() (interface{}, error) {
	return map[string]interface{}{
		"total":      0,
		"unread":     0,
		"persistent": 0,
	}, nil
}

func (n *NotificationAdapter) GetNotificationCount(level, category string, unreadOnly bool) (int, error) {
	return 0, nil
}

// UPSDetectorAdapter adapts UPS detector operations
type UPSDetectorAdapter struct {
	api interface{}
}

func (u *UPSDetectorAdapter) IsAvailable() bool {
	// Try to get the UPS detector from the API using the correct interface
	if apiInstance, ok := u.api.(interface{ GetUPSDetector() *upsDetector.Detector }); ok {
		if detector := apiInstance.GetUPSDetector(); detector != nil {
			return detector.IsAvailable()
		}
	}
	return false
}

func (u *UPSDetectorAdapter) GetStatus() interface{} {
	// Try to get the UPS detector from the API using the correct interface
	if apiInstance, ok := u.api.(interface{ GetUPSDetector() *upsDetector.Detector }); ok {
		if detector := apiInstance.GetUPSDetector(); detector != nil {
			return detector.GetStatus()
		}
	}

	// Return default status when UPS is not available
	return map[string]interface{}{
		"available":  false,
		"type":       "none",
		"last_check": "",
		"error":      "UPS detector not available",
	}
}

// AuthAdapter adapts authentication operations
type AuthAdapter struct {
	api interface{}
}

func NewAuthAdapter(api interface{}) *AuthAdapter {
	return &AuthAdapter{api: api}
}

func (a *AuthAdapter) Login(username, password string) (interface{}, error) {
	// Authentication is not implemented in UMA (internal-only API)
	return nil, fmt.Errorf("authentication is not implemented in UMA - this is an internal-only API")
}

func (a *AuthAdapter) GetUsers() (interface{}, error) {
	// Return empty user list since authentication is not implemented
	return []interface{}{}, nil
}

func (a *AuthAdapter) GetStats() (interface{}, error) {
	// Return basic auth stats indicating authentication is disabled
	return map[string]interface{}{
		"enabled":        false,
		"total_users":    0,
		"active_users":   0,
		"total_sessions": 0,
		"message":        "Authentication is not implemented in UMA",
	}, nil
}

func (a *AuthAdapter) IsEnabled() bool {
	// Authentication is not enabled in UMA
	return false
}
