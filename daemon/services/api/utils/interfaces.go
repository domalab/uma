package utils

// APIInterface defines the interface for API dependencies
// This allows handlers to access the main API functionality without tight coupling
type APIInterface interface {
	GetInfo() interface{}
	GetSystem() SystemInterface
	GetStorage() StorageInterface
	GetDocker() DockerInterface
	GetVM() VMInterface
	GetAuth() AuthInterface

	GetNotifications() NotificationInterface
	GetUPSDetector() UPSDetectorInterface
	GetConfigManager() interface{}
	GetMCPServer() interface{}
}

// SystemInterface defines the interface for system operations
type SystemInterface interface {
	GetCPUInfo() (interface{}, error)
	GetMemoryInfo() (interface{}, error)
	GetLoadInfo() (interface{}, error)
	GetUptimeInfo() (interface{}, error)
	GetNetworkInfo() (interface{}, error)
	GetEnhancedTemperatureData() (interface{}, error)
	GetGPUInfo() (interface{}, error)
	GetSystemLogs() (interface{}, error)
	GetRealArrayInfo() (interface{}, error)
	GetRealDisks() (interface{}, error)
}

// StorageInterface defines the interface for storage operations
type StorageInterface interface {
	GetArrayInfo() (interface{}, error)
	GetDisks() (interface{}, error)
	GetConsolidatedDisksInfo() (interface{}, error)
	GetZFSPools() (interface{}, error)
	GetCacheInfo() (interface{}, error)
	StartArray(request interface{}) error
	StopArray(request interface{}) error
}

// DockerInterface defines the interface for Docker operations
type DockerInterface interface {
	GetContainers() (interface{}, error)
	GetContainersWithStats() (interface{}, error)
	GetContainer(id string) (interface{}, error)
	GetContainerStats(id string) (interface{}, error)
	StartContainer(id string) error
	StopContainer(id string, timeout int) error
	RestartContainer(id string, timeout int) error
	GetImages() (interface{}, error)
	GetNetworks() (interface{}, error)
	GetSystemInfo() (interface{}, error)
}

// VMInterface defines the interface for VM operations
type VMInterface interface {
	GetVMs() (interface{}, error)
	GetVM(name string) (interface{}, error)
	StartVM(name string) error
	StopVM(name string) error
	RestartVM(name string) error
	GetVMStats(name string) (interface{}, error)
	GetVMConsole(name string) (interface{}, error)
	SetVMAutostart(name string, autostart bool) error
}

// NotificationInterface defines the interface for notification operations
type NotificationInterface interface {
	GetNotifications(level, category string, unreadOnly bool) (interface{}, error)
	GetNotificationsPaginated(page, limit int, level, category string, unreadOnly bool) (interface{}, error)
	GetNotification(id string) (interface{}, error)
	CreateNotification(title, message string, level interface{}, category interface{}, metadata map[string]interface{}) (interface{}, error)
	UpdateNotification(id string, updates map[string]interface{}) (interface{}, error)
	DeleteNotification(id string) error
	ClearAllNotifications() error
	MarkAllAsRead() error
	GetNotificationStats() (interface{}, error)
	GetNotificationCount(level, category string, unreadOnly bool) (int, error)
}

// UPSDetectorInterface defines the interface for UPS detection operations
type UPSDetectorInterface interface {
	IsAvailable() bool
	GetStatus() interface{}
}

// AuthInterface defines the interface for authentication operations
type AuthInterface interface {
	Login(username, password string) (interface{}, error)
	GetUsers() (interface{}, error)
	GetStats() (interface{}, error)
	IsEnabled() bool
}
