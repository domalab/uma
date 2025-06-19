package responses

import (
	"time"
)

// Storage-related response types

// ArrayOperationResponse represents the response from array operations
type ArrayOperationResponse struct {
	Success       bool   `json:"success"`
	Message       string `json:"message"`
	OperationID   string `json:"operation_id,omitempty"`
	EstimatedTime int    `json:"estimated_time,omitempty"` // seconds
}

// ParityCheckStatus represents the status of a parity check operation
type ParityCheckStatus struct {
	Active        bool       `json:"active"`
	Type          string     `json:"type,omitempty"`           // "check" or "correct"
	Progress      float64    `json:"progress,omitempty"`       // 0-100
	Speed         string     `json:"speed,omitempty"`          // e.g., "45.2 MB/s"
	TimeRemaining string     `json:"time_remaining,omitempty"` // e.g., "2h 15m"
	Errors        int        `json:"errors,omitempty"`
	StartTime     *time.Time `json:"start_time,omitempty"`
	LastUpdated   time.Time  `json:"last_updated"`
}

// DiskInfo represents information about a disk
type DiskInfo struct {
	Device      string    `json:"device"`
	Name        string    `json:"name"`
	Size        int64     `json:"size"`
	Used        int64     `json:"used"`
	Free        int64     `json:"free"`
	Filesystem  string    `json:"filesystem"`
	MountPoint  string    `json:"mount_point"`
	Status      string    `json:"status"` // "active", "standby", "spun_down", "error"
	Temperature int       `json:"temperature,omitempty"`
	Health      string    `json:"health"` // "healthy", "warning", "critical"
	SMARTStatus string    `json:"smart_status"`
	LastUpdated time.Time `json:"last_updated"`
}

// ArrayStatus represents the status of the Unraid array
type ArrayStatus struct {
	State       string     `json:"state"`      // "started", "stopped", "starting", "stopping"
	Protection  string     `json:"protection"` // "protected", "unprotected", "invalid"
	Disks       []DiskInfo `json:"disks"`
	Parity      []DiskInfo `json:"parity"`
	Cache       []DiskInfo `json:"cache,omitempty"`
	LastUpdated time.Time  `json:"last_updated"`
}

// ShareInfo represents information about a share
type ShareInfo struct {
	Name             string     `json:"name"`
	Comment          string     `json:"comment"`
	Path             string     `json:"path"`
	Size             int64      `json:"size"`
	Used             int64      `json:"used"`
	Free             int64      `json:"free"`
	AllocatorMethod  string     `json:"allocator_method"`
	MinimumFreeSpace string     `json:"minimum_free_space"`
	SplitLevel       int        `json:"split_level"`
	IncludedDisks    []string   `json:"included_disks"`
	ExcludedDisks    []string   `json:"excluded_disks"`
	UseCache         string     `json:"use_cache"`
	CachePool        string     `json:"cache_pool"`
	SMBEnabled       bool       `json:"smb_enabled"`
	SMBSecurity      string     `json:"smb_security"`
	NFSEnabled       bool       `json:"nfs_enabled"`
	AFPEnabled       bool       `json:"afp_enabled"`
	FTPEnabled       bool       `json:"ftp_enabled"`
	LastAccessed     *time.Time `json:"last_accessed,omitempty"`
	LastModified     *time.Time `json:"last_modified,omitempty"`
	LastUpdated      time.Time  `json:"last_updated"`
}

// ZFSPoolInfo represents information about a ZFS pool
type ZFSPoolInfo struct {
	Name        string     `json:"name"`
	State       string     `json:"state"`
	Size        int64      `json:"size"`
	Used        int64      `json:"used"`
	Available   int64      `json:"available"`
	Health      string     `json:"health"`
	Devices     []string   `json:"devices"`
	LastScrub   *time.Time `json:"last_scrub,omitempty"`
	LastUpdated time.Time  `json:"last_updated"`
}

// CacheInfo represents information about cache pools
type CacheInfo struct {
	Name        string    `json:"name"`
	Size        int64     `json:"size"`
	Used        int64     `json:"used"`
	Free        int64     `json:"free"`
	Devices     []string  `json:"devices"`
	Status      string    `json:"status"`
	LastUpdated time.Time `json:"last_updated"`
}

// ShareOperationResponse represents a response from share operations
type ShareOperationResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	ShareName string `json:"share_name,omitempty"`
}

// ShareListResponse represents a list of shares
type ShareListResponse struct {
	Shares interface{} `json:"shares"` // Can be []Share or []models.Share depending on context
}
