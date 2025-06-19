package requests

// Docker-related request types

// DockerContainerActionRequest represents a request to perform an action on a Docker container
type DockerContainerActionRequest struct {
	Action string `json:"action"` // "start", "stop", "restart", "pause", "unpause", "remove"
	Force  bool   `json:"force,omitempty"`
}

// DockerBulkActionRequest represents a request to perform bulk actions on Docker containers
type DockerBulkActionRequest struct {
	ContainerIDs []string `json:"container_ids"`
	Action       string   `json:"action"` // "start", "stop", "restart"
	Force        bool     `json:"force,omitempty"`
}

// DockerContainerCreateRequest represents a request to create a new Docker container
type DockerContainerCreateRequest struct {
	Name          string            `json:"name"`
	Image         string            `json:"image"`
	Command       []string          `json:"command,omitempty"`
	Environment   map[string]string `json:"environment,omitempty"`
	Ports         map[string]string `json:"ports,omitempty"`
	Volumes       map[string]string `json:"volumes,omitempty"`
	Networks      []string          `json:"networks,omitempty"`
	RestartPolicy string            `json:"restart_policy,omitempty"`
	Privileged    bool              `json:"privileged,omitempty"`
	AutoRemove    bool              `json:"auto_remove,omitempty"`
}

// DockerImagePullRequest represents a request to pull a Docker image
type DockerImagePullRequest struct {
	Image string `json:"image"`
	Tag   string `json:"tag,omitempty"`
}

// DockerImageRemoveRequest represents a request to remove a Docker image
type DockerImageRemoveRequest struct {
	ImageID string `json:"image_id"`
	Force   bool   `json:"force,omitempty"`
}

// DockerNetworkCreateRequest represents a request to create a Docker network
type DockerNetworkCreateRequest struct {
	Name    string            `json:"name"`
	Driver  string            `json:"driver,omitempty"`
	Options map[string]string `json:"options,omitempty"`
}

// DockerVolumeCreateRequest represents a request to create a Docker volume
type DockerVolumeCreateRequest struct {
	Name    string            `json:"name"`
	Driver  string            `json:"driver,omitempty"`
	Options map[string]string `json:"options,omitempty"`
}
