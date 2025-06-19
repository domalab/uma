package requests

// Storage-related request types

// ArrayStartRequest represents a request to start the Unraid array
type ArrayStartRequest struct {
	MaintenanceMode bool `json:"maintenance_mode"`
	CheckFilesystem bool `json:"check_filesystem"`
}

// ArrayStopRequest represents a request to stop the Unraid array
type ArrayStopRequest struct {
	Force          bool `json:"force"`
	UnmountShares  bool `json:"unmount_shares"`
	StopContainers bool `json:"stop_containers"`
	StopVMs        bool `json:"stop_vms"`
}

// ParityCheckRequest represents a request to start a parity check
type ParityCheckRequest struct {
	Type     string `json:"type"`     // "check" or "correct"
	Priority string `json:"priority"` // "low", "normal", "high"
}

// DiskAddRequest represents a request to add a disk to the array
type DiskAddRequest struct {
	Device   string `json:"device"`   // e.g., "/dev/sdc"
	Position string `json:"position"` // e.g., "disk1", "parity2"
}

// DiskRemoveRequest represents a request to remove a disk from the array
type DiskRemoveRequest struct {
	Position string `json:"position"` // e.g., "disk1", "parity2"
}

// ShareCreateRequest represents a request to create a new share
type ShareCreateRequest struct {
	Name             string   `json:"name" validate:"required,min=1,max=40"`
	Comment          string   `json:"comment,omitempty"`
	AllocatorMethod  string   `json:"allocator_method,omitempty"` // "high-water", "most-free", "fill-up"
	MinimumFreeSpace string   `json:"minimum_free_space,omitempty"`
	SplitLevel       int      `json:"split_level,omitempty"`
	IncludedDisks    []string `json:"included_disks,omitempty"`
	ExcludedDisks    []string `json:"excluded_disks,omitempty"`
	UseCache         string   `json:"use_cache,omitempty"`  // "yes", "no", "only", "prefer"
	CachePool        string   `json:"cache_pool,omitempty"` // "cache", "cache2", etc.
	SMBEnabled       bool     `json:"smb_enabled,omitempty"`
	SMBSecurity      string   `json:"smb_security,omitempty"` // "public", "secure", "private"
	SMBGuests        bool     `json:"smb_guests,omitempty"`
	NFSEnabled       bool     `json:"nfs_enabled,omitempty"`
	NFSSecurity      string   `json:"nfs_security,omitempty"` // "public", "secure", "private"
	AFPEnabled       bool     `json:"afp_enabled,omitempty"`
	FTPEnabled       bool     `json:"ftp_enabled,omitempty"`
}

// ShareUpdateRequest represents a request to update an existing share
type ShareUpdateRequest struct {
	Comment          string   `json:"comment,omitempty"`
	AllocatorMethod  string   `json:"allocator_method,omitempty"`
	MinimumFreeSpace string   `json:"minimum_free_space,omitempty"`
	SplitLevel       int      `json:"split_level,omitempty"`
	IncludedDisks    []string `json:"included_disks,omitempty"`
	ExcludedDisks    []string `json:"excluded_disks,omitempty"`
	UseCache         string   `json:"use_cache,omitempty"`
	CachePool        string   `json:"cache_pool,omitempty"`
	SMBEnabled       bool     `json:"smb_enabled,omitempty"`
	SMBSecurity      string   `json:"smb_security,omitempty"`
	SMBGuests        bool     `json:"smb_guests,omitempty"`
	NFSEnabled       bool     `json:"nfs_enabled,omitempty"`
	NFSSecurity      string   `json:"nfs_security,omitempty"`
	AFPEnabled       bool     `json:"afp_enabled,omitempty"`
	FTPEnabled       bool     `json:"ftp_enabled,omitempty"`
}
