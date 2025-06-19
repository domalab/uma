package responses

import "time"

// VM-related response types

// VMInfo represents information about a virtual machine
type VMInfo struct {
	Name        string      `json:"name"`
	UUID        string      `json:"uuid"`
	Description string      `json:"description"`
	State       string      `json:"state"` // "running", "stopped", "paused", "suspended"
	CPUs        int         `json:"cpus"`
	Memory      int         `json:"memory"` // Memory in MB
	Storage     []VMStorage `json:"storage,omitempty"`
	Networks    []VMNetwork `json:"networks,omitempty"`
	Autostart   bool        `json:"autostart"`
	Created     time.Time   `json:"created"`
	LastStarted *time.Time  `json:"last_started,omitempty"`
	LastStopped *time.Time  `json:"last_stopped,omitempty"`
	LastUpdated time.Time   `json:"last_updated"`
}

// VMStorage represents VM storage information
type VMStorage struct {
	Type     string `json:"type"`     // "disk", "cdrom", "floppy"
	Source   string `json:"source"`   // Path to disk image or ISO
	Target   string `json:"target"`   // Target device (e.g., "hda", "sda")
	Bus      string `json:"bus"`      // "ide", "scsi", "virtio"
	Format   string `json:"format"`   // "raw", "qcow2", "vmdk"
	Size     int64  `json:"size"`     // Size in bytes
	Used     int64  `json:"used"`     // Used space in bytes
	ReadOnly bool   `json:"readonly"` // Read-only flag
	Bootable bool   `json:"bootable"` // Bootable flag
}

// VMNetwork represents VM network information
type VMNetwork struct {
	Type      string `json:"type"`       // "bridge", "nat", "host"
	Source    string `json:"source"`     // Bridge name or network name
	MAC       string `json:"mac"`        // MAC address
	Model     string `json:"model"`      // Network model (e.g., "virtio", "e1000")
	IPAddress string `json:"ip_address"` // Current IP address (if available)
	Status    string `json:"status"`     // "up", "down"
}

// VMStats represents VM performance statistics
type VMStats struct {
	Name          string    `json:"name"`
	CPUTime       int64     `json:"cpu_time"`       // CPU time in nanoseconds
	CPUPercent    float64   `json:"cpu_percent"`    // CPU usage percentage
	MemoryUsage   int64     `json:"memory_usage"`   // Memory usage in bytes
	MemoryPercent float64   `json:"memory_percent"` // Memory usage percentage
	DiskRead      int64     `json:"disk_read"`      // Disk read bytes
	DiskWrite     int64     `json:"disk_write"`     // Disk write bytes
	NetworkRx     int64     `json:"network_rx"`     // Network received bytes
	NetworkTx     int64     `json:"network_tx"`     // Network transmitted bytes
	LastUpdated   time.Time `json:"last_updated"`
}

// VMConsoleInfo represents VM console information
type VMConsoleInfo struct {
	Type      string `json:"type"`      // "vnc", "spice", "serial"
	Host      string `json:"host"`      // Console host
	Port      int    `json:"port"`      // Console port
	Password  string `json:"password"`  // Console password (if any)
	WebSocket bool   `json:"websocket"` // WebSocket support
	SSL       bool   `json:"ssl"`       // SSL/TLS support
}

// VMSnapshotInfo represents VM snapshot information
type VMSnapshotInfo struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	State       string    `json:"state"`  // "disk-snapshot", "internal"
	Parent      string    `json:"parent"` // Parent snapshot name
	Created     time.Time `json:"created"`
	Size        int64     `json:"size"`    // Snapshot size in bytes
	Current     bool      `json:"current"` // Is current snapshot
}

// VMTemplateInfo represents VM template information
type VMTemplateInfo struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	OS          string     `json:"os"`        // Operating system
	Version     string     `json:"version"`   // Template version
	CPUs        int        `json:"cpus"`      // Default CPU count
	Memory      int        `json:"memory"`    // Default memory in MB
	DiskSize    int64      `json:"disk_size"` // Default disk size in bytes
	Created     time.Time  `json:"created"`
	LastUsed    *time.Time `json:"last_used,omitempty"`
}

// VMOperationResponse represents the response from VM operations
type VMOperationResponse struct {
	Success     bool   `json:"success"`
	Message     string `json:"message"`
	OperationID string `json:"operation_id,omitempty"`
	VMName      string `json:"vm_name"`
}

// VMListResponse represents a list of VMs
type VMListResponse struct {
	VMs         []VMInfo  `json:"vms"`
	Total       int       `json:"total"`
	Running     int       `json:"running"`
	Stopped     int       `json:"stopped"`
	Paused      int       `json:"paused"`
	LastUpdated time.Time `json:"last_updated"`
}
