package responses

import "time"

// Docker-related response types

// DockerContainerInfo represents information about a Docker container
type DockerContainerInfo struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Image       string            `json:"image"`
	Status      string            `json:"status"` // "running", "stopped", "paused", "restarting"
	State       string            `json:"state"`  // "created", "running", "paused", "restarting", "removing", "exited", "dead"
	Created     time.Time         `json:"created"`
	Started     *time.Time        `json:"started,omitempty"`
	Finished    *time.Time        `json:"finished,omitempty"`
	Ports       []PortMapping     `json:"ports,omitempty"`
	Volumes     []VolumeMapping   `json:"volumes,omitempty"`
	Networks    []NetworkMapping  `json:"networks,omitempty"`
	Environment map[string]string `json:"environment,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	CPUUsage    float64           `json:"cpu_usage,omitempty"`    // Percentage
	MemoryUsage int64             `json:"memory_usage,omitempty"` // Bytes
	MemoryLimit int64             `json:"memory_limit,omitempty"` // Bytes
	LastUpdated time.Time         `json:"last_updated"`
}

// PortMapping represents a port mapping for a container
type PortMapping struct {
	PrivatePort int    `json:"private_port"`
	PublicPort  int    `json:"public_port,omitempty"`
	Type        string `json:"type"` // "tcp", "udp"
	IP          string `json:"ip,omitempty"`
}

// VolumeMapping represents a volume mapping for a container
type VolumeMapping struct {
	Source      string `json:"source"`
	Destination string `json:"destination"`
	Mode        string `json:"mode"` // "rw", "ro"
}

// NetworkMapping represents a network mapping for a container
type NetworkMapping struct {
	NetworkID   string `json:"network_id"`
	NetworkName string `json:"network_name"`
	IPAddress   string `json:"ip_address,omitempty"`
	Gateway     string `json:"gateway,omitempty"`
}

// DockerImageInfo represents information about a Docker image
type DockerImageInfo struct {
	ID          string            `json:"id"`
	Repository  string            `json:"repository"`
	Tag         string            `json:"tag"`
	Size        int64             `json:"size"`
	Created     time.Time         `json:"created"`
	Labels      map[string]string `json:"labels,omitempty"`
	LastUpdated time.Time         `json:"last_updated"`
}

// DockerNetworkInfo represents information about a Docker network
type DockerNetworkInfo struct {
	ID          string             `json:"id"`
	Name        string             `json:"name"`
	Driver      string             `json:"driver"`
	Scope       string             `json:"scope"`
	Internal    bool               `json:"internal"`
	Attachable  bool               `json:"attachable"`
	Containers  []NetworkContainer `json:"containers,omitempty"`
	Options     map[string]string  `json:"options,omitempty"`
	Labels      map[string]string  `json:"labels,omitempty"`
	LastUpdated time.Time          `json:"last_updated"`
}

// NetworkContainer represents a container attached to a network
type NetworkContainer struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	IPAddress string `json:"ip_address"`
}

// DockerVolumeInfo represents information about a Docker volume
type DockerVolumeInfo struct {
	Name        string            `json:"name"`
	Driver      string            `json:"driver"`
	Mountpoint  string            `json:"mountpoint"`
	Size        int64             `json:"size,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	Options     map[string]string `json:"options,omitempty"`
	LastUpdated time.Time         `json:"last_updated"`
}

// DockerSystemInfo represents Docker system information
type DockerSystemInfo struct {
	Version           string    `json:"version"`
	APIVersion        string    `json:"api_version"`
	KernelVersion     string    `json:"kernel_version"`
	OS                string    `json:"os"`
	Architecture      string    `json:"architecture"`
	Containers        int       `json:"containers"`
	ContainersRunning int       `json:"containers_running"`
	ContainersPaused  int       `json:"containers_paused"`
	ContainersStopped int       `json:"containers_stopped"`
	Images            int       `json:"images"`
	StorageDriver     string    `json:"storage_driver"`
	LoggingDriver     string    `json:"logging_driver"`
	MemoryLimit       bool      `json:"memory_limit"`
	SwapLimit         bool      `json:"swap_limit"`
	CPUShares         bool      `json:"cpu_shares"`
	LastUpdated       time.Time `json:"last_updated"`
}

// DockerStats represents Docker container statistics
type DockerStats struct {
	ContainerID   string    `json:"container_id"`
	Name          string    `json:"name"`
	CPUPercent    float64   `json:"cpu_percent"`
	MemoryUsage   int64     `json:"memory_usage"`
	MemoryLimit   int64     `json:"memory_limit"`
	MemoryPercent float64   `json:"memory_percent"`
	NetworkRx     int64     `json:"network_rx"`
	NetworkTx     int64     `json:"network_tx"`
	BlockRead     int64     `json:"block_read"`
	BlockWrite    int64     `json:"block_write"`
	PIDs          int       `json:"pids"`
	LastUpdated   time.Time `json:"last_updated"`
}

// ContainerOperationResult represents the result of an operation on a single container
type ContainerOperationResult struct {
	ContainerID   string `json:"container_id"`
	ContainerName string `json:"container_name,omitempty"`
	Success       bool   `json:"success"`
	Error         string `json:"error,omitempty"`
	Duration      string `json:"duration,omitempty"`
}
