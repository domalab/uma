package async

import (
	"github.com/domalab/uma/daemon/plugins/docker"
	"github.com/domalab/uma/daemon/plugins/storage"
)

// StorageMonitorAdapter adapts the storage.StorageMonitor to the async interface
type StorageMonitorAdapter struct {
	monitor *storage.StorageMonitor
}

// NewStorageMonitorAdapter creates a new storage monitor adapter
func NewStorageMonitorAdapter(monitor *storage.StorageMonitor) *StorageMonitorAdapter {
	return &StorageMonitorAdapter{
		monitor: monitor,
	}
}

// StartParityCheck starts a parity check operation
func (a *StorageMonitorAdapter) StartParityCheck(checkType string, priority string) error {
	// Convert to the storage monitor's expected format
	return a.monitor.StartParityCheck(checkType, priority)
}

// GetParityCheckStatus gets the current parity check status
func (a *StorageMonitorAdapter) GetParityCheckStatus() (map[string]interface{}, error) {
	status, err := a.monitor.GetParityCheckStatus()
	if err != nil {
		return nil, err
	}

	// Convert the storage.ParityCheckStatus to map[string]interface{}
	result := map[string]interface{}{
		"active":         status.Active,
		"progress":       status.Progress,
		"speed":          status.Speed,
		"time_remaining": status.TimeRemaining,
		"errors":         status.Errors,
		"type":           status.Type,
	}

	return result, nil
}

// CancelParityCheck cancels the current parity check
func (a *StorageMonitorAdapter) CancelParityCheck() error {
	return a.monitor.CancelParityCheck()
}

// DockerManagerAdapter adapts the docker.DockerManager to the async interface
type DockerManagerAdapter struct {
	manager *docker.DockerManager
}

// NewDockerManagerAdapter creates a new docker manager adapter
func NewDockerManagerAdapter(manager *docker.DockerManager) *DockerManagerAdapter {
	return &DockerManagerAdapter{
		manager: manager,
	}
}

// StartContainer starts a Docker container
func (a *DockerManagerAdapter) StartContainer(containerID string) error {
	return a.manager.StartContainer(containerID)
}

// StopContainer stops a Docker container
func (a *DockerManagerAdapter) StopContainer(containerID string, timeout int) error {
	return a.manager.StopContainer(containerID, timeout)
}

// RestartContainer restarts a Docker container
func (a *DockerManagerAdapter) RestartContainer(containerID string, timeout int) error {
	return a.manager.RestartContainer(containerID, timeout)
}
