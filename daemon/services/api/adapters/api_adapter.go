package adapters

import (
	"fmt"
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

// GetNotifications returns the notification interface
func (a *APIAdapter) GetNotifications() utils.NotificationInterface {
	return &NotificationAdapter{api: a.api}
}

// GetUPSDetector returns the UPS detector interface
func (a *APIAdapter) GetUPSDetector() utils.UPSDetectorInterface {
	return &UPSDetectorAdapter{api: a.api}
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

			// Initialize slice fields for each container to prevent null values
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
				result[i] = container
			}
			logger.Blue("Successfully retrieved and processed %d Docker containers", len(containers))
			return result, nil
		}
	}

	// Fallback to empty array if Docker manager not available
	logger.Yellow("Docker manager not available, returning empty container list")
	return []interface{}{}, nil
}

func (d *DockerAdapter) GetContainer(id string) (interface{}, error) {
	// Try to get the Docker manager from the API
	if apiInstance, ok := d.api.(interface{ GetDockerManager() interface{} }); ok {
		if dockerManager := apiInstance.GetDockerManager(); dockerManager != nil {
			// Use reflection to call GetContainer method
			if dm, ok := dockerManager.(interface {
				GetContainer(string) (interface{}, error)
			}); ok {
				return dm.GetContainer(id)
			}
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
	// Try to get the Docker manager from the API
	if apiInstance, ok := d.api.(interface{ GetDockerManager() interface{} }); ok {
		if dockerManager := apiInstance.GetDockerManager(); dockerManager != nil {
			// Use reflection to call StartContainer method
			if dm, ok := dockerManager.(interface{ StartContainer(string) error }); ok {
				return dm.StartContainer(id)
			}
		}
	}

	return nil
}

func (d *DockerAdapter) StopContainer(id string) error {
	// Try to get the Docker manager from the API
	if apiInstance, ok := d.api.(interface{ GetDockerManager() interface{} }); ok {
		if dockerManager := apiInstance.GetDockerManager(); dockerManager != nil {
			// Use reflection to call StopContainer method
			if dm, ok := dockerManager.(interface{ StopContainer(string, int) error }); ok {
				return dm.StopContainer(id, 10) // 10 second timeout
			}
		}
	}

	return nil
}

func (d *DockerAdapter) RestartContainer(id string) error {
	// Try to get the Docker manager from the API
	if apiInstance, ok := d.api.(interface{ GetDockerManager() interface{} }); ok {
		if dockerManager := apiInstance.GetDockerManager(); dockerManager != nil {
			// Use reflection to call RestartContainer method
			if dm, ok := dockerManager.(interface{ RestartContainer(string, int) error }); ok {
				return dm.RestartContainer(id, 10) // 10 second timeout
			}
		}
	}

	return nil
}

func (d *DockerAdapter) GetImages() (interface{}, error) {
	// Try to get the Docker manager from the API
	if apiInstance, ok := d.api.(interface{ GetDockerManager() interface{} }); ok {
		if dockerManager := apiInstance.GetDockerManager(); dockerManager != nil {
			// Use reflection to call ListImages method
			if dm, ok := dockerManager.(interface{ ListImages() (interface{}, error) }); ok {
				return dm.ListImages()
			}
		}
	}

	return []interface{}{}, nil
}

func (d *DockerAdapter) GetNetworks() (interface{}, error) {
	// Try to get the Docker manager from the API
	if apiInstance, ok := d.api.(interface{ GetDockerManager() interface{} }); ok {
		if dockerManager := apiInstance.GetDockerManager(); dockerManager != nil {
			// Use reflection to call GetNetworks method
			if dm, ok := dockerManager.(interface{ GetNetworks() (interface{}, error) }); ok {
				return dm.GetNetworks()
			}
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
