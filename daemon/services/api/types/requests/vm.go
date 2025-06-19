package requests

// VM-related request types

// VMActionRequest represents a request to perform an action on a VM
type VMActionRequest struct {
	Action string `json:"action"` // "start", "stop", "restart", "pause", "resume", "reset"
	Force  bool   `json:"force,omitempty"`
}

// VMCreateRequest represents a request to create a new VM
type VMCreateRequest struct {
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Template    string            `json:"template,omitempty"`
	CPUs        int               `json:"cpus"`
	Memory      int               `json:"memory"` // Memory in MB
	Storage     []VMStorageConfig `json:"storage,omitempty"`
	Networks    []VMNetworkConfig `json:"networks,omitempty"`
	Autostart   bool              `json:"autostart,omitempty"`
}

// VMUpdateRequest represents a request to update VM configuration
type VMUpdateRequest struct {
	Description string            `json:"description,omitempty"`
	CPUs        int               `json:"cpus,omitempty"`
	Memory      int               `json:"memory,omitempty"` // Memory in MB
	Storage     []VMStorageConfig `json:"storage,omitempty"`
	Networks    []VMNetworkConfig `json:"networks,omitempty"`
	Autostart   bool              `json:"autostart,omitempty"`
}

// VMStorageConfig represents VM storage configuration
type VMStorageConfig struct {
	Type     string `json:"type"`     // "disk", "cdrom", "floppy"
	Source   string `json:"source"`   // Path to disk image or ISO
	Target   string `json:"target"`   // Target device (e.g., "hda", "sda")
	Bus      string `json:"bus"`      // "ide", "scsi", "virtio"
	Format   string `json:"format"`   // "raw", "qcow2", "vmdk"
	Size     int    `json:"size"`     // Size in GB (for new disks)
	ReadOnly bool   `json:"readonly"` // Read-only flag
}

// VMNetworkConfig represents VM network configuration
type VMNetworkConfig struct {
	Type   string `json:"type"`   // "bridge", "nat", "host"
	Source string `json:"source"` // Bridge name or network name
	MAC    string `json:"mac"`    // MAC address (optional)
	Model  string `json:"model"`  // Network model (e.g., "virtio", "e1000")
}

// VMSnapshotRequest represents a request to create a VM snapshot
type VMSnapshotRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// VMCloneRequest represents a request to clone a VM
type VMCloneRequest struct {
	NewName     string `json:"new_name"`
	Description string `json:"description,omitempty"`
	LinkedClone bool   `json:"linked_clone,omitempty"`
}
