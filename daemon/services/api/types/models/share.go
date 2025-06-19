package models

import "time"

// Share represents a Unraid share configuration
type Share struct {
	Name             string    `json:"name"`
	Comment          string    `json:"comment"`
	Path             string    `json:"path"`
	AllocatorMethod  string    `json:"allocator_method"` // "high-water", "most-free", "fill-up"
	MinimumFreeSpace string    `json:"minimum_free_space"`
	SplitLevel       int       `json:"split_level"`
	IncludedDisks    []string  `json:"included_disks"`
	ExcludedDisks    []string  `json:"excluded_disks"`
	UseCache         string    `json:"use_cache"`  // "yes", "no", "only", "prefer"
	CachePool        string    `json:"cache_pool"` // "cache", "cache2", etc.
	SMBEnabled       bool      `json:"smb_enabled"`
	SMBSecurity      string    `json:"smb_security"` // "public", "secure", "private"
	NFSEnabled       bool      `json:"nfs_enabled"`
	AFPEnabled       bool      `json:"afp_enabled"`
	FTPEnabled       bool      `json:"ftp_enabled"`
	Created          time.Time `json:"created"`
	Modified         time.Time `json:"modified"`
}

// ShareAccess represents share access permissions
type ShareAccess struct {
	ShareName   string   `json:"share_name"`
	Users       []string `json:"users"`
	Groups      []string `json:"groups"`
	ReadOnly    bool     `json:"read_only"`
	GuestAccess bool     `json:"guest_access"`
}

// ShareUsage represents share usage statistics
type ShareUsage struct {
	ShareName   string    `json:"share_name"`
	TotalSize   int64     `json:"total_size"`
	UsedSize    int64     `json:"used_size"`
	FreeSize    int64     `json:"free_size"`
	FileCount   int64     `json:"file_count"`
	DirCount    int64     `json:"dir_count"`
	LastUpdated time.Time `json:"last_updated"`
}

// ShareBackup represents share backup configuration
type ShareBackup struct {
	ShareName   string     `json:"share_name"`
	Enabled     bool       `json:"enabled"`
	Schedule    string     `json:"schedule"`    // Cron expression
	Destination string     `json:"destination"` // Backup destination path
	Retention   int        `json:"retention"`   // Days to keep backups
	Compression bool       `json:"compression"` // Enable compression
	Encryption  bool       `json:"encryption"`  // Enable encryption
	LastBackup  *time.Time `json:"last_backup,omitempty"`
	NextBackup  *time.Time `json:"next_backup,omitempty"`
	BackupSize  int64      `json:"backup_size"` // Size of last backup
	Status      string     `json:"status"`      // "success", "failed", "running"
	LastUpdated time.Time  `json:"last_updated"`
}

// LegacyShareUsage represents share usage statistics (legacy format from http_server.go)
type LegacyShareUsage struct {
	Name           string  `json:"name"`
	TotalSize      int64   `json:"total_size"`   // bytes
	UsedSize       int64   `json:"used_size"`    // bytes
	FreeSize       int64   `json:"free_size"`    // bytes
	UsedPercent    float64 `json:"used_percent"` // 0-100
	FileCount      int64   `json:"file_count"`
	DirectoryCount int64   `json:"directory_count"`
	LastAccessed   string  `json:"last_accessed"` // ISO 8601 format
}
