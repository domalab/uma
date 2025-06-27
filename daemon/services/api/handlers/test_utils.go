package handlers

import (
	"fmt"

	"github.com/domalab/uma/daemon/services/api/utils"
)

// MockAPIInterface provides a mock implementation of the API interface for testing
type MockAPIInterface struct{}

// GetInfo returns mock system info
func (m *MockAPIInterface) GetInfo() interface{} {
	return map[string]interface{}{
		"version": "test-version",
		"status":  "healthy",
	}
}

// GetSystem returns a mock system interface
func (m *MockAPIInterface) GetSystem() utils.SystemInterface {
	return &MockSystemInterface{}
}

// GetStorage returns a mock storage interface
func (m *MockAPIInterface) GetStorage() utils.StorageInterface {
	return &MockStorageInterface{}
}

// GetUPSDetector returns a mock UPS detector interface
func (m *MockAPIInterface) GetUPSDetector() utils.UPSDetectorInterface {
	return &MockUPSDetectorInterface{}
}

// GetDocker returns a mock Docker interface
func (m *MockAPIInterface) GetDocker() utils.DockerInterface {
	return &MockDockerInterface{}
}

// GetVM returns a mock VM interface
func (m *MockAPIInterface) GetVM() utils.VMInterface {
	return &MockVMInterface{}
}

// GetNotifications returns a mock notification interface
func (m *MockAPIInterface) GetNotifications() utils.NotificationInterface {
	return &MockNotificationInterface{}
}

// GetAuth returns a mock authentication interface
func (m *MockAPIInterface) GetAuth() utils.AuthInterface {
	return &MockAuthInterface{}
}

// GetConfigManager returns a mock configuration manager
func (m *MockAPIInterface) GetConfigManager() interface{} {
	return &MockConfigManager{}
}

// GetMCPServer returns a mock MCP server
func (m *MockAPIInterface) GetMCPServer() interface{} {
	return nil // Return nil to simulate disabled MCP server
}

// MockConfigManager provides mock configuration management
type MockConfigManager struct {
	mcpEnabled        bool
	mcpMaxConnections int
}

func (m *MockConfigManager) SetMCPEnabled(enabled bool) error {
	m.mcpEnabled = enabled
	return nil
}

func (m *MockConfigManager) SetMCPMaxConnections(maxConnections int) error {
	if maxConnections <= 0 {
		return fmt.Errorf("invalid max connections: %d", maxConnections)
	}
	m.mcpMaxConnections = maxConnections
	return nil
}

// MockSystemInterface provides mock system functionality
type MockSystemInterface struct{}

func (m *MockSystemInterface) GetCPUInfo() (interface{}, error) {
	return map[string]interface{}{
		"cores":        4,
		"usage":        25.5,
		"temperature":  65.0,
		"last_updated": "2024-01-01T00:00:00Z",
	}, nil
}

func (m *MockSystemInterface) GetMemoryInfo() (interface{}, error) {
	return map[string]interface{}{
		"total":        8192,
		"used":         4096,
		"available":    4096,
		"last_updated": "2024-01-01T00:00:00Z",
	}, nil
}

func (m *MockSystemInterface) GetLoadInfo() (interface{}, error) {
	return map[string]interface{}{"load1": 1.5, "load5": 1.2, "load15": 1.0}, nil
}

func (m *MockSystemInterface) GetUptimeInfo() (interface{}, error) {
	return map[string]interface{}{"uptime": 86400}, nil
}

func (m *MockSystemInterface) GetNetworkInfo() (interface{}, error) {
	return []interface{}{}, nil
}

func (m *MockSystemInterface) GetEnhancedTemperatureData() (interface{}, error) {
	return map[string]interface{}{}, nil
}

func (m *MockSystemInterface) GetGPUInfo() (interface{}, error) {
	return map[string]interface{}{
		"gpus": []interface{}{
			map[string]interface{}{
				"name":        "Intel UHD Graphics 630",
				"vendor":      "Intel",
				"type":        "integrated",
				"driver":      "i915",
				"temperature": 45.0,
			},
		},
	}, nil
}

func (m *MockSystemInterface) GetSystemLogs() (interface{}, error) {
	return map[string]interface{}{
		"logs": []interface{}{
			map[string]interface{}{
				"name":    "syslog",
				"path":    "/var/log/syslog",
				"entries": []interface{}{},
			},
		},
	}, nil
}

func (m *MockSystemInterface) GetRealArrayInfo() (interface{}, error) {
	return map[string]interface{}{
		"state":         "started",
		"protection":    "parity",
		"disks":         []interface{}{},
		"parity":        []interface{}{},
		"sync_action":   "none",
		"sync_progress": 0.0,
		"parity_history": map[string]interface{}{
			"last_check":     nil,
			"last_duration":  nil,
			"last_speed":     nil,
			"last_errors":    0,
			"last_action":    "unknown",
			"next_scheduled": nil,
			"checks":         []interface{}{},
		},
	}, nil
}

func (m *MockSystemInterface) GetRealDisks() (interface{}, error) {
	return []interface{}{
		map[string]interface{}{
			"name":          "sda",
			"size":          "1TB",
			"used":          500000000000,
			"available":     500000000000,
			"usage_percent": 50.0,
			"mount_point":   "/mnt/disk1",
			"health":        "healthy",
			"temperature":   35,
		},
	}, nil
}

// MockStorageInterface provides mock storage functionality
type MockStorageInterface struct{}

func (m *MockStorageInterface) GetArrayInfo() (interface{}, error) {
	return map[string]interface{}{"status": "started"}, nil
}

func (m *MockStorageInterface) GetDisks() (interface{}, error) {
	return []interface{}{}, nil
}

func (m *MockStorageInterface) GetConsolidatedDisksInfo() (interface{}, error) {
	return map[string]interface{}{
		"array_disks":  []interface{}{},
		"parity_disks": []interface{}{},
		"cache_disks":  []interface{}{},
		"boot_disk":    nil,
		"summary": map[string]interface{}{
			"total_disks": 0,
			"total_size":  0,
		},
	}, nil
}

func (m *MockStorageInterface) GetZFSPools() (interface{}, error) {
	return []interface{}{}, nil
}

func (m *MockStorageInterface) GetCacheInfo() (interface{}, error) {
	return map[string]interface{}{}, nil
}

func (m *MockStorageInterface) StartArray(request interface{}) error {
	return nil
}

func (m *MockStorageInterface) StopArray(request interface{}) error {
	return nil
}

// MockDockerInterface provides mock Docker functionality
type MockDockerInterface struct{}

func (m *MockDockerInterface) GetContainers() (interface{}, error) {
	return []interface{}{}, nil
}

func (m *MockDockerInterface) GetContainersWithStats() (interface{}, error) {
	return []interface{}{}, nil
}

func (m *MockDockerInterface) GetContainer(id string) (interface{}, error) {
	// Return error for nonexistent containers
	if id == "nonexistent-container" {
		return nil, fmt.Errorf("container not found: %s", id)
	}

	return map[string]interface{}{
		"id":       id,
		"name":     "test-container",
		"state":    "running", // Default state that allows stop operations
		"status":   "Up 2 hours",
		"image":    "nginx:latest",
		"ports":    []interface{}{},
		"mounts":   []interface{}{},
		"networks": []interface{}{},
		"labels":   map[string]interface{}{},
	}, nil
}

func (m *MockDockerInterface) StartContainer(id string) error {
	return nil
}

func (m *MockDockerInterface) StopContainer(id string, timeout int) error {
	return nil
}

func (m *MockDockerInterface) RestartContainer(id string, timeout int) error {
	return nil
}

func (m *MockDockerInterface) GetImages() (interface{}, error) {
	return []interface{}{}, nil
}

func (m *MockDockerInterface) GetNetworks() (interface{}, error) {
	return []interface{}{}, nil
}

func (m *MockDockerInterface) GetContainerStats(id string) (interface{}, error) {
	return map[string]interface{}{
		"container_id": id,
		"cpu_percent":  0.0,
		"memory_usage": 0,
	}, nil
}

func (m *MockDockerInterface) GetSystemInfo() (interface{}, error) {
	return map[string]interface{}{}, nil
}

// MockVMInterface provides mock VM functionality
type MockVMInterface struct{}

func (m *MockVMInterface) GetVMs() (interface{}, error) {
	return []interface{}{}, nil
}

func (m *MockVMInterface) GetVM(name string) (interface{}, error) {
	return map[string]interface{}{"name": name}, nil
}

func (m *MockVMInterface) StartVM(name string) error {
	return nil
}

func (m *MockVMInterface) StopVM(name string) error {
	return nil
}

func (m *MockVMInterface) RestartVM(name string) error {
	return nil
}

func (m *MockVMInterface) GetVMStats(name string) (interface{}, error) {
	return map[string]interface{}{}, nil
}

func (m *MockVMInterface) GetVMConsole(name string) (interface{}, error) {
	return map[string]interface{}{}, nil
}

func (m *MockVMInterface) SetVMAutostart(name string, autostart bool) error {
	return nil
}

// MockNotificationInterface provides mock notification functionality
type MockNotificationInterface struct{}

func (m *MockNotificationInterface) GetNotifications(level, category string, unreadOnly bool) (interface{}, error) {
	return []interface{}{}, nil
}

func (m *MockNotificationInterface) GetNotificationsPaginated(page, limit int, level, category string, unreadOnly bool) (interface{}, error) {
	return map[string]interface{}{"notifications": []interface{}{}, "total": 0}, nil
}

func (m *MockNotificationInterface) GetNotification(id string) (interface{}, error) {
	return map[string]interface{}{"id": id}, nil
}

func (m *MockNotificationInterface) CreateNotification(title, message string, level interface{}, category interface{}, metadata map[string]interface{}) (interface{}, error) {
	return map[string]interface{}{"id": "test-id"}, nil
}

func (m *MockNotificationInterface) UpdateNotification(id string, updates map[string]interface{}) (interface{}, error) {
	return map[string]interface{}{"id": id}, nil
}

func (m *MockNotificationInterface) DeleteNotification(id string) error {
	return nil
}

func (m *MockNotificationInterface) ClearAllNotifications() error {
	return nil
}

func (m *MockNotificationInterface) MarkAllAsRead() error {
	return nil
}

func (m *MockNotificationInterface) GetNotificationStats() (interface{}, error) {
	return map[string]interface{}{"total": 0, "unread": 0}, nil
}

func (m *MockNotificationInterface) GetNotificationCount(level, category string, unreadOnly bool) (int, error) {
	return 0, nil
}

// MockUPSDetectorInterface provides mock UPS detector functionality
type MockUPSDetectorInterface struct{}

func (m *MockUPSDetectorInterface) IsAvailable() bool {
	return false // Mock UPS as not available by default
}

func (m *MockUPSDetectorInterface) GetStatus() interface{} {
	return map[string]interface{}{
		"available":  false,
		"type":       "none",
		"last_check": "2024-01-01T00:00:00Z",
		"error":      "No UPS detected (mock)",
	}
}

// MockAuthInterface provides mock authentication functionality
type MockAuthInterface struct{}

func (m *MockAuthInterface) Login(username, password string) (interface{}, error) {
	return nil, fmt.Errorf("authentication is not implemented in UMA")
}

func (m *MockAuthInterface) GetUsers() (interface{}, error) {
	return []interface{}{}, nil
}

func (m *MockAuthInterface) GetStats() (interface{}, error) {
	return map[string]interface{}{
		"enabled":        false,
		"total_users":    0,
		"active_users":   0,
		"total_sessions": 0,
		"message":        "Authentication is not implemented in UMA",
	}, nil
}

func (m *MockAuthInterface) IsEnabled() bool {
	return false
}
