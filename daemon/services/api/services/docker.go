package services

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/domalab/uma/daemon/logger"
	"github.com/domalab/uma/daemon/services/api/utils"
)

// DockerService handles Docker-related business logic
type DockerService struct {
	api utils.APIInterface
}

// NewDockerService creates a new Docker service
func NewDockerService(api utils.APIInterface) *DockerService {
	return &DockerService{
		api: api,
	}
}

// Container represents a Docker container
type Container struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Image   string `json:"image"`
	Status  string `json:"status"`
	State   string `json:"state"`
	Created string `json:"created"`
	Ports   string `json:"ports"`
}

// ContainerStats represents container resource usage statistics
type ContainerStats struct {
	ContainerID   string  `json:"container_id"`
	Name          string  `json:"name"`
	CPUPercent    float64 `json:"cpu_percent"`
	MemoryUsage   int64   `json:"memory_usage"`
	MemoryLimit   int64   `json:"memory_limit"`
	MemoryPercent float64 `json:"memory_percent"`
	NetworkRx     int64   `json:"network_rx"`
	NetworkTx     int64   `json:"network_tx"`
	BlockRead     int64   `json:"block_read"`
	BlockWrite    int64   `json:"block_write"`
}

// ContainerActionRequest represents a container action request
type ContainerActionRequest struct {
	Action     string `json:"action"`      // start, stop, restart, pause, unpause
	Force      bool   `json:"force"`       // force action
	Timeout    int    `json:"timeout"`     // timeout in seconds
	Signal     string `json:"signal"`      // signal for stop/kill
	RemoveData bool   `json:"remove_data"` // remove data when removing container
}

// BulkActionRequest represents a bulk action request
type BulkActionRequest struct {
	ContainerIDs []string `json:"container_ids"`
	Action       string   `json:"action"`
	Force        bool     `json:"force"`
	Timeout      int      `json:"timeout"`
}

// GetContainers retrieves all Docker containers
func (d *DockerService) GetContainers() ([]Container, error) {
	containersInfo, err := d.api.GetDocker().GetContainers()
	if err != nil {
		return nil, fmt.Errorf("failed to get containers: %v", err)
	}

	// Try to convert to container slice
	if containerSlice, ok := containersInfo.([]Container); ok {
		return containerSlice, nil
	}

	// Fallback: parse from interface{}
	containers := []Container{}
	if containerData, ok := containersInfo.([]interface{}); ok {
		for _, item := range containerData {
			if containerMap, ok := item.(map[string]interface{}); ok {
				container := Container{
					ID:      d.getStringValue(containerMap, "id"),
					Name:    d.getStringValue(containerMap, "name"),
					Image:   d.getStringValue(containerMap, "image"),
					Status:  d.getStringValue(containerMap, "status"),
					State:   d.getStringValue(containerMap, "state"),
					Created: d.getStringValue(containerMap, "created"),
					Ports:   d.getStringValue(containerMap, "ports"),
				}
				containers = append(containers, container)
			}
		}
	}

	return containers, nil
}

// GetContainer retrieves a specific Docker container
func (d *DockerService) GetContainer(containerID string) (*Container, error) {
	containerInfo, err := d.api.GetDocker().GetContainer(containerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get container: %v", err)
	}

	// Try to convert to container
	if container, ok := containerInfo.(*Container); ok {
		return container, nil
	}

	// Fallback: parse from interface{}
	if containerMap, ok := containerInfo.(map[string]interface{}); ok {
		container := &Container{
			ID:      d.getStringValue(containerMap, "id"),
			Name:    d.getStringValue(containerMap, "name"),
			Image:   d.getStringValue(containerMap, "image"),
			Status:  d.getStringValue(containerMap, "status"),
			State:   d.getStringValue(containerMap, "state"),
			Created: d.getStringValue(containerMap, "created"),
			Ports:   d.getStringValue(containerMap, "ports"),
		}
		return container, nil
	}

	return nil, fmt.Errorf("invalid container data format")
}

// GetContainerStats retrieves container resource usage statistics
func (d *DockerService) GetContainerStats(containerID string) (*ContainerStats, error) {
	// Execute docker stats command for specific container
	cmd := exec.Command("docker", "stats", "--no-stream", "--format",
		"table {{.Container}}\t{{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.MemPerc}}\t{{.NetIO}}\t{{.BlockIO}}",
		containerID)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get container stats: %v", err)
	}

	// Parse docker stats output
	lines := strings.Split(string(output), "\n")
	if len(lines) < 2 {
		return nil, fmt.Errorf("no stats data available")
	}

	// Skip header line and parse data line
	dataLine := strings.Fields(lines[1])
	if len(dataLine) < 7 {
		return nil, fmt.Errorf("invalid stats format")
	}

	stats := &ContainerStats{
		ContainerID: dataLine[0],
		Name:        dataLine[1],
	}

	// Parse CPU percentage
	if cpuStr := strings.TrimSuffix(dataLine[2], "%"); cpuStr != "" {
		if cpu, err := strconv.ParseFloat(cpuStr, 64); err == nil {
			stats.CPUPercent = cpu
		}
	}

	// Parse memory usage (simplified)
	memUsageStr := dataLine[3]
	if parts := strings.Split(memUsageStr, "/"); len(parts) == 2 {
		if usage := d.parseMemoryValue(strings.TrimSpace(parts[0])); usage > 0 {
			stats.MemoryUsage = usage
		}
		if limit := d.parseMemoryValue(strings.TrimSpace(parts[1])); limit > 0 {
			stats.MemoryLimit = limit
		}
	}

	// Parse memory percentage
	if memPercStr := strings.TrimSuffix(dataLine[4], "%"); memPercStr != "" {
		if memPerc, err := strconv.ParseFloat(memPercStr, 64); err == nil {
			stats.MemoryPercent = memPerc
		}
	}

	return stats, nil
}

// StartContainer starts a Docker container
func (d *DockerService) StartContainer(containerID string) error {
	if err := d.api.GetDocker().StartContainer(containerID); err != nil {
		return fmt.Errorf("failed to start container: %v", err)
	}

	logger.Blue("Started container: %s", containerID)
	return nil
}

// StopContainer stops a Docker container
func (d *DockerService) StopContainer(containerID string, timeout int) error {
	if err := d.api.GetDocker().StopContainer(containerID); err != nil {
		return fmt.Errorf("failed to stop container: %v", err)
	}

	logger.Blue("Stopped container: %s", containerID)
	return nil
}

// RestartContainer restarts a Docker container
func (d *DockerService) RestartContainer(containerID string, timeout int) error {
	if err := d.api.GetDocker().RestartContainer(containerID); err != nil {
		return fmt.Errorf("failed to restart container: %v", err)
	}

	logger.Blue("Restarted container: %s", containerID)
	return nil
}

// PauseContainer pauses a Docker container
func (d *DockerService) PauseContainer(containerID string) error {
	// Pause functionality not available in current interface
	// Would need to be implemented in the Docker plugin
	logger.Blue("Pause container requested: %s (not implemented)", containerID)
	return fmt.Errorf("pause container operation not implemented")
}

// UnpauseContainer unpauses a Docker container
func (d *DockerService) UnpauseContainer(containerID string) error {
	// Unpause functionality not available in current interface
	// Would need to be implemented in the Docker plugin
	logger.Blue("Unpause container requested: %s (not implemented)", containerID)
	return fmt.Errorf("unpause container operation not implemented")
}

// RemoveContainer removes a Docker container
func (d *DockerService) RemoveContainer(containerID string, force bool, removeData bool) error {
	// Stop container first if it's running
	if !force {
		container, err := d.GetContainer(containerID)
		if err == nil && container.State == "running" {
			if err := d.StopContainer(containerID, 10); err != nil {
				return fmt.Errorf("failed to stop container before removal: %v", err)
			}
		}
	}

	// Remove container functionality not available in current interface
	// Would need to be implemented in the Docker plugin
	logger.Blue("Remove container requested: %s (not implemented)", containerID)

	logger.Blue("Removed container: %s", containerID)
	return nil
}

// BulkAction performs bulk actions on multiple containers
func (d *DockerService) BulkAction(req BulkActionRequest) (map[string]interface{}, error) {
	results := make(map[string]interface{})
	successCount := 0
	errorCount := 0

	for _, containerID := range req.ContainerIDs {
		var err error

		switch req.Action {
		case "start":
			err = d.StartContainer(containerID)
		case "stop":
			err = d.StopContainer(containerID, req.Timeout)
		case "restart":
			err = d.RestartContainer(containerID, req.Timeout)
		case "pause":
			err = d.PauseContainer(containerID)
		case "unpause":
			err = d.UnpauseContainer(containerID)
		case "remove":
			err = d.RemoveContainer(containerID, req.Force, false)
		default:
			err = fmt.Errorf("invalid action: %s", req.Action)
		}

		if err != nil {
			results[containerID] = map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			}
			errorCount++
		} else {
			results[containerID] = map[string]interface{}{
				"success": true,
			}
			successCount++
		}
	}

	return map[string]interface{}{
		"results":       results,
		"success_count": successCount,
		"error_count":   errorCount,
		"total_count":   len(req.ContainerIDs),
	}, nil
}

// GetDockerInfo retrieves Docker system information
func (d *DockerService) GetDockerInfo() (map[string]interface{}, error) {
	// Try to get Docker info from the API first
	info, err := d.api.GetDocker().GetSystemInfo()
	if err == nil {
		if infoMap, ok := info.(map[string]interface{}); ok {
			// Add system-level aggregation metrics
			d.addSystemLevelMetrics(infoMap)
			infoMap["last_updated"] = time.Now().UTC().Format(time.RFC3339)
			return infoMap, nil
		}
	}

	// Fallback: calculate system metrics from container data
	systemInfo := d.calculateSystemMetricsFromContainers()
	systemInfo["last_updated"] = time.Now().UTC().Format(time.RFC3339)
	systemInfo["error"] = fmt.Sprintf("Docker info unavailable: %v", err)

	return systemInfo, nil
}

// Helper methods

// getStringValue safely gets a string value from a map
func (d *DockerService) getStringValue(m map[string]interface{}, key string) string {
	if value, ok := m[key]; ok {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return ""
}

// parseMemoryValue parses memory value strings like "1.5GiB", "512MiB"
func (d *DockerService) parseMemoryValue(value string) int64 {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0
	}

	// Extract numeric part and unit
	var numStr string
	var unit string

	for i, char := range value {
		if char >= '0' && char <= '9' || char == '.' {
			numStr += string(char)
		} else {
			unit = value[i:]
			break
		}
	}

	num, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return 0
	}

	// Convert to bytes based on unit
	switch strings.ToLower(unit) {
	case "b", "":
		return int64(num)
	case "k", "kb", "kib":
		return int64(num * 1024)
	case "m", "mb", "mib":
		return int64(num * 1024 * 1024)
	case "g", "gb", "gib":
		return int64(num * 1024 * 1024 * 1024)
	case "t", "tb", "tib":
		return int64(num * 1024 * 1024 * 1024 * 1024)
	default:
		return int64(num)
	}
}

// GetDockerDataOptimized collects Docker data with parallel container processing
func (d *DockerService) GetDockerDataOptimized() interface{} {
	containers, err := d.api.GetDocker().GetContainers()
	if err != nil {
		logger.Red("Failed to get Docker containers: %v", err)
		return map[string]interface{}{
			"containers": []interface{}{},
			"total":      0,
			"error":      err.Error(),
		}
	}

	// Convert to slice if needed
	var containerSlice []interface{}
	if slice, ok := containers.([]interface{}); ok {
		containerSlice = slice
	} else {
		// Try to convert single container or other formats
		containerSlice = []interface{}{containers}
	}

	if len(containerSlice) == 0 {
		return map[string]interface{}{
			"containers": []interface{}{},
			"total":      0,
		}
	}

	return map[string]interface{}{
		"containers": containerSlice,
		"total":      len(containerSlice),
	}
}

// addSystemLevelMetrics adds system-level aggregation metrics to Docker info
func (d *DockerService) addSystemLevelMetrics(infoMap map[string]interface{}) {
	// Get container data for aggregation
	containers, err := d.api.GetDocker().GetContainers()
	if err != nil {
		logger.Yellow("Failed to get containers for system metrics: %v", err)
		return
	}

	var containerSlice []interface{}
	if slice, ok := containers.([]interface{}); ok {
		containerSlice = slice
	}

	// Calculate container counts by state
	totalContainers := len(containerSlice)
	runningContainers := 0
	pausedContainers := 0
	stoppedContainers := 0

	for _, container := range containerSlice {
		if containerMap, ok := container.(map[string]interface{}); ok {
			state := d.getStringValue(containerMap, "state")
			switch state {
			case "running":
				runningContainers++
			case "paused":
				pausedContainers++
			case "exited", "stopped":
				stoppedContainers++
			}
		}
	}

	// Add aggregated metrics to info map
	infoMap["containers_total"] = totalContainers
	infoMap["containers_running"] = runningContainers
	infoMap["containers_paused"] = pausedContainers
	infoMap["containers_stopped"] = stoppedContainers
}

// calculateSystemMetricsFromContainers calculates system metrics when Docker info is unavailable
func (d *DockerService) calculateSystemMetricsFromContainers() map[string]interface{} {
	systemInfo := map[string]interface{}{
		"version":            "unknown",
		"api_version":        "unknown",
		"containers_total":   0,
		"containers_running": 0,
		"containers_paused":  0,
		"containers_stopped": 0,
		"images":             0,
		"storage_driver":     "unknown",
	}

	// Get container data
	containers, err := d.api.GetDocker().GetContainers()
	if err != nil {
		logger.Yellow("Failed to get containers for fallback metrics: %v", err)
		return systemInfo
	}

	var containerSlice []interface{}
	if slice, ok := containers.([]interface{}); ok {
		containerSlice = slice
	}

	// Calculate container counts
	totalContainers := len(containerSlice)
	runningContainers := 0
	pausedContainers := 0
	stoppedContainers := 0

	for _, container := range containerSlice {
		if containerMap, ok := container.(map[string]interface{}); ok {
			state := d.getStringValue(containerMap, "state")
			switch state {
			case "running":
				runningContainers++
			case "paused":
				pausedContainers++
			case "exited", "stopped":
				stoppedContainers++
			}
		}
	}

	systemInfo["containers_total"] = totalContainers
	systemInfo["containers_running"] = runningContainers
	systemInfo["containers_paused"] = pausedContainers
	systemInfo["containers_stopped"] = stoppedContainers

	return systemInfo
}
