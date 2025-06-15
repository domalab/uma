package docker

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/domalab/uma/daemon/lib"
	"github.com/domalab/uma/daemon/logger"
)

// DockerManager provides Docker container management capabilities
type DockerManager struct{}

// ContainerInfo represents information about a Docker container
type ContainerInfo struct {
	ID            string            `json:"id"`
	Name          string            `json:"name"`
	Image         string            `json:"image"`
	Status        string            `json:"status"`
	State         string            `json:"state"`
	Created       time.Time         `json:"created"`
	StartedAt     time.Time         `json:"started_at,omitempty"`
	Ports         []PortMapping     `json:"ports"`
	Mounts        []MountInfo       `json:"mounts"`
	Networks      []NetworkInfo     `json:"networks"`
	Labels        map[string]string `json:"labels"`
	Environment   []string          `json:"environment,omitempty"`
	RestartPolicy string            `json:"restart_policy"`
	CPUUsage      float64           `json:"cpu_usage_percent,omitempty"`
	MemoryUsage   uint64            `json:"memory_usage_bytes,omitempty"`
	MemoryLimit   uint64            `json:"memory_limit_bytes,omitempty"`
	NetworkRx     uint64            `json:"network_rx_bytes,omitempty"`
	NetworkTx     uint64            `json:"network_tx_bytes,omitempty"`
}

// PortMapping represents a port mapping
type PortMapping struct {
	HostIP        string `json:"host_ip"`
	HostPort      string `json:"host_port"`
	ContainerPort string `json:"container_port"`
	Protocol      string `json:"protocol"`
}

// MountInfo represents a mount point
type MountInfo struct {
	Type        string `json:"type"`
	Source      string `json:"source"`
	Destination string `json:"destination"`
	Mode        string `json:"mode"`
	ReadWrite   bool   `json:"read_write"`
}

// NetworkInfo represents network information
type NetworkInfo struct {
	Name      string `json:"name"`
	IPAddress string `json:"ip_address"`
	Gateway   string `json:"gateway"`
}

// DockerStats represents Docker container statistics
type DockerStats struct {
	ContainerID string  `json:"container_id"`
	Name        string  `json:"name"`
	CPUPercent  float64 `json:"cpu_percent"`
	MemUsage    uint64  `json:"memory_usage"`
	MemLimit    uint64  `json:"memory_limit"`
	MemPercent  float64 `json:"memory_percent"`
	NetIO       string  `json:"net_io"`
	BlockIO     string  `json:"block_io"`
}

// DockerNetwork represents a Docker network
type DockerNetwork struct {
	ID         string            `json:"id"`
	Name       string            `json:"name"`
	Driver     string            `json:"driver"`
	Scope      string            `json:"scope"`
	Internal   bool              `json:"internal"`
	Attachable bool              `json:"attachable"`
	Ingress    bool              `json:"ingress"`
	IPAM       DockerIPAM        `json:"ipam"`
	Created    time.Time         `json:"created"`
	Options    map[string]string `json:"options,omitempty"`
	Labels     map[string]string `json:"labels,omitempty"`
}

// DockerIPAM represents Docker network IPAM configuration
type DockerIPAM struct {
	Driver  string             `json:"driver"`
	Config  []DockerIPAMConfig `json:"config"`
	Options map[string]string  `json:"options,omitempty"`
}

// DockerIPAMConfig represents Docker IPAM configuration
type DockerIPAMConfig struct {
	Subnet  string `json:"subnet,omitempty"`
	Gateway string `json:"gateway,omitempty"`
}

// DockerImage represents a Docker image
type DockerImage struct {
	ID          string            `json:"id"`
	Repository  string            `json:"repository"`
	Tag         string            `json:"tag"`
	Digest      string            `json:"digest,omitempty"`
	Size        int64             `json:"size"`
	Created     time.Time         `json:"created"`
	Labels      map[string]string `json:"labels,omitempty"`
	RepoTags    []string          `json:"repo_tags,omitempty"`
	RepoDigests []string          `json:"repo_digests,omitempty"`
}

// NewDockerManager creates a new Docker manager
func NewDockerManager() *DockerManager {
	return &DockerManager{}
}

// IsDockerAvailable checks if Docker is available and running
func (d *DockerManager) IsDockerAvailable() bool {
	output := lib.GetCmdOutput("docker", "version", "--format", "{{.Server.Version}}")
	return len(output) > 0 && !strings.Contains(strings.Join(output, ""), "Cannot connect")
}

// ListContainers returns a list of all containers
func (d *DockerManager) ListContainers(all bool) ([]ContainerInfo, error) {
	containers := make([]ContainerInfo, 0)

	if !d.IsDockerAvailable() {
		return containers, fmt.Errorf("Docker is not available")
	}

	args := []string{"ps", "--format", "json", "--no-trunc"}
	if all {
		args = append(args, "--all")
	}

	output := lib.GetCmdOutput("docker", args...)
	logger.Blue("Docker ps output: %d lines", len(output))

	for i, line := range output {
		if strings.TrimSpace(line) == "" {
			continue
		}

		logger.Blue("Processing line %d: %s", i, line[:min(100, len(line))])

		// Parse the docker ps JSON format first
		var dockerPsData map[string]interface{}
		if err := json.Unmarshal([]byte(line), &dockerPsData); err != nil {
			logger.Yellow("Failed to parse container JSON: %v", err)
			continue
		}

		// Convert to our ContainerInfo format
		container := ContainerInfo{}
		if err := d.parseDockerPsData(&container, dockerPsData); err != nil {
			logger.Yellow("Failed to parse docker ps data: %v", err)
			continue
		}

		// Get detailed information
		if err := d.getContainerDetails(&container); err != nil {
			logger.Yellow("Failed to get container details for %s: %v", container.ID, err)
		}

		containers = append(containers, container)
	}

	logger.Blue("Found %d containers", len(containers))
	return containers, nil
}

// parseDockerPsData parses docker ps JSON output into ContainerInfo
func (d *DockerManager) parseDockerPsData(container *ContainerInfo, data map[string]interface{}) error {
	if id, ok := data["ID"].(string); ok {
		container.ID = id
	}

	if names, ok := data["Names"].(string); ok {
		container.Name = names
	}

	if image, ok := data["Image"].(string); ok {
		container.Image = image
	}

	if status, ok := data["Status"].(string); ok {
		container.Status = status
	}

	if state, ok := data["State"].(string); ok {
		container.State = state
	}

	if createdAt, ok := data["CreatedAt"].(string); ok {
		// Parse the created time - format: "2025-06-13 04:01:43 +1000 AEST"
		if t, err := time.Parse("2006-01-02 15:04:05 -0700 MST", createdAt); err == nil {
			container.Created = t
		}
	}

	return nil
}

// min helper function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// GetContainer returns information about a specific container
func (d *DockerManager) GetContainer(nameOrID string) (*ContainerInfo, error) {
	if !d.IsDockerAvailable() {
		return nil, fmt.Errorf("Docker is not available")
	}

	output := lib.GetCmdOutput("docker", "inspect", nameOrID)
	if len(output) == 0 {
		return nil, fmt.Errorf("container not found: %s", nameOrID)
	}

	// Parse the JSON output
	jsonStr := strings.Join(output, "")
	var inspectData []map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &inspectData); err != nil {
		return nil, fmt.Errorf("failed to parse inspect output: %w", err)
	}

	if len(inspectData) == 0 {
		return nil, fmt.Errorf("no container data found")
	}

	container := &ContainerInfo{}
	if err := d.parseInspectData(container, inspectData[0]); err != nil {
		return nil, err
	}

	return container, nil
}

// StartContainer starts a container
func (d *DockerManager) StartContainer(nameOrID string) error {
	if !d.IsDockerAvailable() {
		return fmt.Errorf("Docker is not available")
	}

	output := lib.GetCmdOutput("docker", "start", nameOrID)
	if len(output) == 0 {
		return fmt.Errorf("failed to start container: %s", nameOrID)
	}

	// Check if there were any errors
	for _, line := range output {
		if strings.Contains(line, "Error") {
			return fmt.Errorf("error starting container: %s", line)
		}
	}

	logger.Blue("Started container: %s", nameOrID)
	return nil
}

// StopContainer stops a container
func (d *DockerManager) StopContainer(nameOrID string, timeout int) error {
	if !d.IsDockerAvailable() {
		return fmt.Errorf("Docker is not available")
	}

	args := []string{"stop"}
	if timeout > 0 {
		args = append(args, "--time", strconv.Itoa(timeout))
	}
	args = append(args, nameOrID)

	output := lib.GetCmdOutput("docker", args...)

	// Check if there were any errors
	for _, line := range output {
		if strings.Contains(line, "Error") {
			return fmt.Errorf("error stopping container: %s", line)
		}
	}

	logger.Blue("Stopped container: %s", nameOrID)
	return nil
}

// RestartContainer restarts a container
func (d *DockerManager) RestartContainer(nameOrID string, timeout int) error {
	if !d.IsDockerAvailable() {
		return fmt.Errorf("Docker is not available")
	}

	args := []string{"restart"}
	if timeout > 0 {
		args = append(args, "--time", strconv.Itoa(timeout))
	}
	args = append(args, nameOrID)

	output := lib.GetCmdOutput("docker", args...)

	// Check if there were any errors
	for _, line := range output {
		if strings.Contains(line, "Error") {
			return fmt.Errorf("error restarting container: %s", line)
		}
	}

	logger.Blue("Restarted container: %s", nameOrID)
	return nil
}

// GetContainerLogs returns logs for a container
func (d *DockerManager) GetContainerLogs(nameOrID string, lines int, follow bool) ([]string, error) {
	if !d.IsDockerAvailable() {
		return nil, fmt.Errorf("Docker is not available")
	}

	args := []string{"logs"}
	if lines > 0 {
		args = append(args, "--tail", strconv.Itoa(lines))
	}
	if follow {
		args = append(args, "--follow")
	}
	args = append(args, nameOrID)

	output := lib.GetCmdOutput("docker", args...)
	return output, nil
}

// GetContainerStats returns real-time statistics for containers
func (d *DockerManager) GetContainerStats(nameOrID string) (*DockerStats, error) {
	if !d.IsDockerAvailable() {
		return nil, fmt.Errorf("Docker is not available")
	}

	args := []string{"stats", "--no-stream", "--format", "json"}
	if nameOrID != "" {
		args = append(args, nameOrID)
	}

	output := lib.GetCmdOutput("docker", args...)
	if len(output) == 0 {
		return nil, fmt.Errorf("no stats available")
	}

	var stats DockerStats
	if err := json.Unmarshal([]byte(output[0]), &stats); err != nil {
		return nil, fmt.Errorf("failed to parse stats: %w", err)
	}

	return &stats, nil
}

// PauseContainer pauses a container
func (d *DockerManager) PauseContainer(nameOrID string) error {
	if !d.IsDockerAvailable() {
		return fmt.Errorf("Docker is not available")
	}

	output := lib.GetCmdOutput("docker", "pause", nameOrID)

	// Check if there were any errors
	for _, line := range output {
		if strings.Contains(line, "Error") {
			return fmt.Errorf("error pausing container: %s", line)
		}
	}

	logger.Blue("Paused container: %s", nameOrID)
	return nil
}

// UnpauseContainer unpauses a container
func (d *DockerManager) UnpauseContainer(nameOrID string) error {
	if !d.IsDockerAvailable() {
		return fmt.Errorf("Docker is not available")
	}

	output := lib.GetCmdOutput("docker", "unpause", nameOrID)

	// Check if there were any errors
	for _, line := range output {
		if strings.Contains(line, "Error") {
			return fmt.Errorf("error unpausing container: %s", line)
		}
	}

	logger.Blue("Unpaused container: %s", nameOrID)
	return nil
}

// RemoveContainer removes a container
func (d *DockerManager) RemoveContainer(nameOrID string, force bool) error {
	if !d.IsDockerAvailable() {
		return fmt.Errorf("Docker is not available")
	}

	args := []string{"rm"}
	if force {
		args = append(args, "--force")
	}
	args = append(args, nameOrID)

	output := lib.GetCmdOutput("docker", args...)

	// Check if there were any errors
	for _, line := range output {
		if strings.Contains(line, "Error") {
			return fmt.Errorf("error removing container: %s", line)
		}
	}

	logger.Blue("Removed container: %s", nameOrID)
	return nil
}

// GetDockerInfo returns Docker system information
func (d *DockerManager) GetDockerInfo() (map[string]interface{}, error) {
	if !d.IsDockerAvailable() {
		return nil, fmt.Errorf("Docker is not available")
	}

	output := lib.GetCmdOutput("docker", "info", "--format", "json")
	if len(output) == 0 {
		return nil, fmt.Errorf("failed to get Docker info")
	}

	var info map[string]interface{}
	jsonStr := strings.Join(output, "")
	if err := json.Unmarshal([]byte(jsonStr), &info); err != nil {
		return nil, fmt.Errorf("failed to parse Docker info: %w", err)
	}

	return info, nil
}

// ListNetworks returns a list of Docker networks
func (d *DockerManager) ListNetworks() ([]DockerNetwork, error) {
	networks := make([]DockerNetwork, 0)

	if !d.IsDockerAvailable() {
		return networks, fmt.Errorf("Docker is not available")
	}

	output := lib.GetCmdOutput("docker", "network", "ls", "--format", "json", "--no-trunc")
	logger.Blue("Docker network ls output: %d lines", len(output))

	for i, line := range output {
		if strings.TrimSpace(line) == "" {
			continue
		}

		logger.Blue("Processing network line %d: %s", i, line[:min(100, len(line))])

		// Parse the docker network ls JSON format
		var networkData map[string]interface{}
		if err := json.Unmarshal([]byte(line), &networkData); err != nil {
			logger.Yellow("Failed to parse network JSON: %v", err)
			continue
		}

		// Convert to our DockerNetwork format
		network := DockerNetwork{}
		if err := d.parseNetworkData(&network, networkData); err != nil {
			logger.Yellow("Failed to parse network data: %v", err)
			continue
		}

		// Get detailed network information
		if err := d.getNetworkDetails(&network); err != nil {
			logger.Yellow("Failed to get network details for %s: %v", network.ID, err)
		}

		networks = append(networks, network)
	}

	return networks, nil
}

// ListImages returns a list of Docker images
func (d *DockerManager) ListImages() ([]DockerImage, error) {
	images := make([]DockerImage, 0)

	if !d.IsDockerAvailable() {
		return images, fmt.Errorf("Docker is not available")
	}

	output := lib.GetCmdOutput("docker", "images", "--format", "json", "--no-trunc")
	logger.Blue("Docker images output: %d lines", len(output))

	for i, line := range output {
		if strings.TrimSpace(line) == "" {
			continue
		}

		logger.Blue("Processing image line %d: %s", i, line[:min(100, len(line))])

		// Parse the docker images JSON format
		var imageData map[string]interface{}
		if err := json.Unmarshal([]byte(line), &imageData); err != nil {
			logger.Yellow("Failed to parse image JSON: %v", err)
			continue
		}

		// Convert to our DockerImage format
		image := DockerImage{}
		if err := d.parseImageData(&image, imageData); err != nil {
			logger.Yellow("Failed to parse image data: %v", err)
			continue
		}

		// Get detailed image information
		if err := d.getImageDetails(&image); err != nil {
			logger.Yellow("Failed to get image details for %s: %v", image.ID, err)
		}

		images = append(images, image)
	}

	return images, nil
}

// parseImageData parses docker images JSON output into DockerImage
func (d *DockerManager) parseImageData(image *DockerImage, data map[string]interface{}) error {
	if id, ok := data["ID"].(string); ok {
		image.ID = id
	}

	if repository, ok := data["Repository"].(string); ok {
		image.Repository = repository
	}

	if tag, ok := data["Tag"].(string); ok {
		image.Tag = tag
	}

	if digest, ok := data["Digest"].(string); ok {
		image.Digest = digest
	}

	if size, ok := data["Size"].(string); ok {
		// Parse size string (e.g., "1.23GB", "456MB")
		if sizeBytes, err := d.parseSizeString(size); err == nil {
			image.Size = sizeBytes
		}
	}

	if createdAt, ok := data["CreatedAt"].(string); ok {
		// Parse the created time - format: "2025-06-13 04:01:43 +1000 AEST"
		if t, err := time.Parse("2006-01-02 15:04:05 -0700 MST", createdAt); err == nil {
			image.Created = t
		}
	}

	return nil
}

// getImageDetails gets detailed information about an image using docker image inspect
func (d *DockerManager) getImageDetails(image *DockerImage) error {
	output := lib.GetCmdOutput("docker", "image", "inspect", image.ID)
	if len(output) == 0 {
		return fmt.Errorf("failed to inspect image %s", image.ID)
	}

	// Parse the JSON output
	jsonStr := strings.Join(output, "")
	var inspectData []map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &inspectData); err != nil {
		return fmt.Errorf("failed to parse image inspect output: %w", err)
	}

	if len(inspectData) == 0 {
		return fmt.Errorf("no image data found")
	}

	return d.parseImageInspectData(image, inspectData[0])
}

// parseImageInspectData parses Docker image inspect data into DockerImage
func (d *DockerManager) parseImageInspectData(image *DockerImage, data map[string]interface{}) error {
	// Parse basic information
	if id, ok := data["Id"].(string); ok {
		image.ID = id
	}

	if repoTags, ok := data["RepoTags"].([]interface{}); ok {
		image.RepoTags = make([]string, 0, len(repoTags))
		for _, tag := range repoTags {
			if str, ok := tag.(string); ok {
				image.RepoTags = append(image.RepoTags, str)
			}
		}
	}

	if repoDigests, ok := data["RepoDigests"].([]interface{}); ok {
		image.RepoDigests = make([]string, 0, len(repoDigests))
		for _, digest := range repoDigests {
			if str, ok := digest.(string); ok {
				image.RepoDigests = append(image.RepoDigests, str)
			}
		}
	}

	// Parse created time
	if created, ok := data["Created"].(string); ok {
		if t, err := time.Parse(time.RFC3339Nano, created); err == nil {
			image.Created = t
		}
	}

	// Parse size
	if size, ok := data["Size"].(float64); ok {
		image.Size = int64(size)
	}

	// Parse config for labels
	if config, ok := data["Config"].(map[string]interface{}); ok {
		if labels, ok := config["Labels"].(map[string]interface{}); ok {
			image.Labels = make(map[string]string)
			for k, v := range labels {
				if str, ok := v.(string); ok {
					image.Labels[k] = str
				}
			}
		}
	}

	return nil
}

// parseSizeString parses Docker size strings like "1.23GB", "456MB" into bytes
func (d *DockerManager) parseSizeString(sizeStr string) (int64, error) {
	sizeStr = strings.TrimSpace(sizeStr)
	if sizeStr == "" {
		return 0, nil
	}

	// Extract numeric part and unit
	var numStr string
	var unit string

	for i, r := range sizeStr {
		if r >= '0' && r <= '9' || r == '.' {
			numStr += string(r)
		} else {
			unit = sizeStr[i:]
			break
		}
	}

	if numStr == "" {
		return 0, fmt.Errorf("invalid size format: %s", sizeStr)
	}

	size, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid size number: %s", numStr)
	}

	// Convert to bytes based on unit
	unit = strings.ToUpper(strings.TrimSpace(unit))
	switch unit {
	case "B", "BYTES":
		return int64(size), nil
	case "KB", "K":
		return int64(size * 1024), nil
	case "MB", "M":
		return int64(size * 1024 * 1024), nil
	case "GB", "G":
		return int64(size * 1024 * 1024 * 1024), nil
	case "TB", "T":
		return int64(size * 1024 * 1024 * 1024 * 1024), nil
	default:
		// If no unit, assume bytes
		return int64(size), nil
	}
}

// parseNetworkData parses docker network ls JSON output into DockerNetwork
func (d *DockerManager) parseNetworkData(network *DockerNetwork, data map[string]interface{}) error {
	if id, ok := data["ID"].(string); ok {
		network.ID = id
	}

	if name, ok := data["Name"].(string); ok {
		network.Name = name
	}

	if driver, ok := data["Driver"].(string); ok {
		network.Driver = driver
	}

	if scope, ok := data["Scope"].(string); ok {
		network.Scope = scope
	}

	if createdAt, ok := data["CreatedAt"].(string); ok {
		// Parse the created time - format: "2025-06-13 04:01:43 +1000 AEST"
		if t, err := time.Parse("2006-01-02 15:04:05 -0700 MST", createdAt); err == nil {
			network.Created = t
		}
	}

	return nil
}

// getNetworkDetails gets detailed information about a network using docker network inspect
func (d *DockerManager) getNetworkDetails(network *DockerNetwork) error {
	output := lib.GetCmdOutput("docker", "network", "inspect", network.ID)
	if len(output) == 0 {
		return fmt.Errorf("failed to inspect network %s", network.ID)
	}

	// Parse the JSON output
	jsonStr := strings.Join(output, "")
	var inspectData []map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &inspectData); err != nil {
		return fmt.Errorf("failed to parse network inspect output: %w", err)
	}

	if len(inspectData) == 0 {
		return fmt.Errorf("no network data found")
	}

	return d.parseNetworkInspectData(network, inspectData[0])
}

// parseNetworkInspectData parses Docker network inspect data into DockerNetwork
func (d *DockerManager) parseNetworkInspectData(network *DockerNetwork, data map[string]interface{}) error {
	// Parse basic information
	if id, ok := data["Id"].(string); ok {
		network.ID = id
	}

	if name, ok := data["Name"].(string); ok {
		network.Name = name
	}

	if driver, ok := data["Driver"].(string); ok {
		network.Driver = driver
	}

	if scope, ok := data["Scope"].(string); ok {
		network.Scope = scope
	}

	if internal, ok := data["Internal"].(bool); ok {
		network.Internal = internal
	}

	if attachable, ok := data["Attachable"].(bool); ok {
		network.Attachable = attachable
	}

	if ingress, ok := data["Ingress"].(bool); ok {
		network.Ingress = ingress
	}

	// Parse created time
	if created, ok := data["Created"].(string); ok {
		if t, err := time.Parse(time.RFC3339Nano, created); err == nil {
			network.Created = t
		}
	}

	// Parse IPAM configuration
	if ipamData, ok := data["IPAM"].(map[string]interface{}); ok {
		network.IPAM = DockerIPAM{}
		if driver, ok := ipamData["Driver"].(string); ok {
			network.IPAM.Driver = driver
		}

		if configData, ok := ipamData["Config"].([]interface{}); ok {
			network.IPAM.Config = make([]DockerIPAMConfig, 0)
			for _, configItem := range configData {
				if configMap, ok := configItem.(map[string]interface{}); ok {
					config := DockerIPAMConfig{}
					if subnet, ok := configMap["Subnet"].(string); ok {
						config.Subnet = subnet
					}
					if gateway, ok := configMap["Gateway"].(string); ok {
						config.Gateway = gateway
					}
					network.IPAM.Config = append(network.IPAM.Config, config)
				}
			}
		}

		if options, ok := ipamData["Options"].(map[string]interface{}); ok {
			network.IPAM.Options = make(map[string]string)
			for k, v := range options {
				if str, ok := v.(string); ok {
					network.IPAM.Options[k] = str
				}
			}
		}
	}

	// Parse options
	if options, ok := data["Options"].(map[string]interface{}); ok {
		network.Options = make(map[string]string)
		for k, v := range options {
			if str, ok := v.(string); ok {
				network.Options[k] = str
			}
		}
	}

	// Parse labels
	if labels, ok := data["Labels"].(map[string]interface{}); ok {
		network.Labels = make(map[string]string)
		for k, v := range labels {
			if str, ok := v.(string); ok {
				network.Labels[k] = str
			}
		}
	}

	return nil
}

// getContainerDetails gets detailed information about a container
func (d *DockerManager) getContainerDetails(container *ContainerInfo) error {
	// Get detailed inspect information
	output := lib.GetCmdOutput("docker", "inspect", container.ID)
	if len(output) == 0 {
		return fmt.Errorf("failed to inspect container")
	}

	// Parse the JSON output
	jsonStr := strings.Join(output, "")
	var inspectData []map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &inspectData); err != nil {
		return err
	}

	if len(inspectData) > 0 {
		return d.parseInspectData(container, inspectData[0])
	}

	return nil
}

// parseInspectData parses Docker inspect data into ContainerInfo
func (d *DockerManager) parseInspectData(container *ContainerInfo, data map[string]interface{}) error {
	// Parse basic information
	if id, ok := data["Id"].(string); ok {
		container.ID = id
	}

	if name, ok := data["Name"].(string); ok {
		container.Name = strings.TrimPrefix(name, "/")
	}

	// Parse config
	if config, ok := data["Config"].(map[string]interface{}); ok {
		if image, ok := config["Image"].(string); ok {
			container.Image = image
		}

		if labels, ok := config["Labels"].(map[string]interface{}); ok {
			container.Labels = make(map[string]string)
			for k, v := range labels {
				if str, ok := v.(string); ok {
					container.Labels[k] = str
				}
			}
		}

		if env, ok := config["Env"].([]interface{}); ok {
			container.Environment = make([]string, 0, len(env))
			for _, e := range env {
				if str, ok := e.(string); ok {
					container.Environment = append(container.Environment, str)
				}
			}
		}
	}

	// Parse state
	if state, ok := data["State"].(map[string]interface{}); ok {
		if status, ok := state["Status"].(string); ok {
			container.Status = status
			container.State = status
		}

		if startedAt, ok := state["StartedAt"].(string); ok {
			if t, err := time.Parse(time.RFC3339Nano, startedAt); err == nil {
				container.StartedAt = t
			}
		}
	}

	// Parse created time
	if created, ok := data["Created"].(string); ok {
		if t, err := time.Parse(time.RFC3339Nano, created); err == nil {
			container.Created = t
		}
	}

	// Parse host config
	if hostConfig, ok := data["HostConfig"].(map[string]interface{}); ok {
		if restartPolicy, ok := hostConfig["RestartPolicy"].(map[string]interface{}); ok {
			if name, ok := restartPolicy["Name"].(string); ok {
				container.RestartPolicy = name
			}
		}

		// Parse port bindings
		if portBindings, ok := hostConfig["PortBindings"].(map[string]interface{}); ok {
			container.Ports = d.parsePortBindings(portBindings)
		}

		// Parse mounts
		if mounts, ok := hostConfig["Mounts"].([]interface{}); ok {
			container.Mounts = d.parseMounts(mounts)
		}
	}

	// Parse network settings
	if networkSettings, ok := data["NetworkSettings"].(map[string]interface{}); ok {
		if networks, ok := networkSettings["Networks"].(map[string]interface{}); ok {
			container.Networks = d.parseNetworks(networks)
		}
	}

	return nil
}

// parsePortBindings parses port binding information
func (d *DockerManager) parsePortBindings(portBindings map[string]interface{}) []PortMapping {
	ports := make([]PortMapping, 0)

	for containerPort, bindings := range portBindings {
		if bindingList, ok := bindings.([]interface{}); ok {
			for _, binding := range bindingList {
				if bindingMap, ok := binding.(map[string]interface{}); ok {
					port := PortMapping{
						ContainerPort: containerPort,
					}

					if hostIP, ok := bindingMap["HostIp"].(string); ok {
						port.HostIP = hostIP
					}
					if hostPort, ok := bindingMap["HostPort"].(string); ok {
						port.HostPort = hostPort
					}

					// Extract protocol from container port (e.g., "80/tcp")
					if parts := strings.Split(containerPort, "/"); len(parts) == 2 {
						port.ContainerPort = parts[0]
						port.Protocol = parts[1]
					}

					ports = append(ports, port)
				}
			}
		}
	}

	return ports
}

// parseMounts parses mount information
func (d *DockerManager) parseMounts(mounts []interface{}) []MountInfo {
	mountList := make([]MountInfo, 0)

	for _, mount := range mounts {
		if mountMap, ok := mount.(map[string]interface{}); ok {
			mountInfo := MountInfo{}

			if mountType, ok := mountMap["Type"].(string); ok {
				mountInfo.Type = mountType
			}
			if source, ok := mountMap["Source"].(string); ok {
				mountInfo.Source = source
			}
			if destination, ok := mountMap["Destination"].(string); ok {
				mountInfo.Destination = destination
			}
			if mode, ok := mountMap["Mode"].(string); ok {
				mountInfo.Mode = mode
			}
			if rw, ok := mountMap["RW"].(bool); ok {
				mountInfo.ReadWrite = rw
			}

			mountList = append(mountList, mountInfo)
		}
	}

	return mountList
}

// parseNetworks parses network information
func (d *DockerManager) parseNetworks(networks map[string]interface{}) []NetworkInfo {
	networkList := make([]NetworkInfo, 0)

	for name, network := range networks {
		if networkMap, ok := network.(map[string]interface{}); ok {
			networkInfo := NetworkInfo{
				Name: name,
			}

			if ipAddress, ok := networkMap["IPAddress"].(string); ok {
				networkInfo.IPAddress = ipAddress
			}
			if gateway, ok := networkMap["Gateway"].(string); ok {
				networkInfo.Gateway = gateway
			}

			networkList = append(networkList, networkInfo)
		}
	}

	return networkList
}
