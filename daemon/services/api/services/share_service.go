package services

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/domalab/uma/daemon/services/api/types/models"
	"github.com/domalab/uma/daemon/services/api/types/requests"
)

// ShareService handles share management operations
type ShareService struct{}

// NewShareService creates a new share service instance
func NewShareService() *ShareService {
	return &ShareService{}
}

// LegacyShare represents a share with legacy string timestamps for backward compatibility
type LegacyShare struct {
	Name             string   `json:"name"`
	Comment          string   `json:"comment"`
	Path             string   `json:"path"`
	AllocatorMethod  string   `json:"allocator_method"` // "high-water", "most-free", "fill-up"
	MinimumFreeSpace string   `json:"minimum_free_space"`
	SplitLevel       int      `json:"split_level"`
	IncludedDisks    []string `json:"included_disks"`
	ExcludedDisks    []string `json:"excluded_disks"`
	UseCache         string   `json:"use_cache"`  // "yes", "no", "only", "prefer"
	CachePool        string   `json:"cache_pool"` // "cache", "cache2", etc.
	SMBEnabled       bool     `json:"smb_enabled"`
	SMBSecurity      string   `json:"smb_security"` // "public", "secure", "private"
	SMBGuests        bool     `json:"smb_guests"`
	NFSEnabled       bool     `json:"nfs_enabled"`
	NFSSecurity      string   `json:"nfs_security"` // "public", "secure", "private"
	AFPEnabled       bool     `json:"afp_enabled"`
	FTPEnabled       bool     `json:"ftp_enabled"`
	CreatedAt        string   `json:"created_at"`
	ModifiedAt       string   `json:"modified_at"`
}

// GetShares returns a list of all configured shares
func (s *ShareService) GetShares() ([]LegacyShare, error) {
	var shares []LegacyShare

	// Read share configuration files from /boot/config/shares/
	sharesDir := "/boot/config/shares"
	if _, err := os.Stat(sharesDir); os.IsNotExist(err) {
		// No shares directory, return empty list
		return shares, nil
	}

	entries, err := os.ReadDir(sharesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read shares directory: %v", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".cfg") {
			shareName := strings.TrimSuffix(entry.Name(), ".cfg")
			share, err := s.ParseShareConfig(shareName)
			if err != nil {
				// Log error but continue with other shares
				continue
			}
			shares = append(shares, *share)
		}
	}

	return shares, nil
}

// GetShare returns detailed information for a specific share
func (s *ShareService) GetShare(shareName string) (*LegacyShare, error) {
	return s.ParseShareConfig(shareName)
}

// ParseShareConfig parses a share configuration file
func (s *ShareService) ParseShareConfig(shareName string) (*LegacyShare, error) {
	configPath := fmt.Sprintf("/boot/config/shares/%s.cfg", shareName)

	content, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("share '%s' not found", shareName)
	}

	share := &LegacyShare{
		Name: shareName,
		Path: fmt.Sprintf("/mnt/user/%s", shareName),
	}

	// Parse the configuration file (simple key=value format)
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.Trim(strings.TrimSpace(parts[1]), "\"")

		switch key {
		case "shareComment":
			share.Comment = value
		case "shareAllocator":
			share.AllocatorMethod = value
		case "shareFloor":
			share.MinimumFreeSpace = value
		case "shareSplitLevel":
			if level, err := strconv.Atoi(value); err == nil {
				share.SplitLevel = level
			}
		case "shareInclude":
			if value != "" {
				share.IncludedDisks = strings.Split(value, ",")
			}
		case "shareExclude":
			if value != "" {
				share.ExcludedDisks = strings.Split(value, ",")
			}
		case "shareUseCache":
			share.UseCache = value
		case "shareCachePool":
			share.CachePool = value
		case "shareExport":
			share.SMBEnabled = (value == "yes")
		case "shareSecurity":
			share.SMBSecurity = value
		case "shareGuest":
			share.SMBGuests = (value == "yes")
		case "shareNFSExport":
			share.NFSEnabled = (value == "yes")
		case "shareNFSSecurity":
			share.NFSSecurity = value
		case "shareAFPExport":
			share.AFPEnabled = (value == "yes")
		case "shareFTPExport":
			share.FTPEnabled = (value == "yes")
		}
	}

	// Get file timestamps
	if stat, err := os.Stat(configPath); err == nil {
		share.ModifiedAt = stat.ModTime().Format(time.RFC3339)
		// For creation time, we'll use the same as modified time
		share.CreatedAt = share.ModifiedAt
	}

	return share, nil
}

// GetShareUsage calculates usage statistics for a share
func (s *ShareService) GetShareUsage(shareName string) (*models.LegacyShareUsage, error) {
	sharePath := fmt.Sprintf("/mnt/user/%s", shareName)

	// Check if share path exists
	if _, err := os.Stat(sharePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("share '%s' not found", shareName)
	}

	usage := &models.LegacyShareUsage{
		Name: shareName,
	}

	// Get filesystem statistics using statvfs
	var stat syscall.Statfs_t
	if err := syscall.Statfs(sharePath, &stat); err != nil {
		return nil, fmt.Errorf("failed to get filesystem stats: %v", err)
	}

	// Calculate sizes
	blockSize := int64(stat.Bsize)
	usage.TotalSize = int64(stat.Blocks) * blockSize
	usage.FreeSize = int64(stat.Bavail) * blockSize
	usage.UsedSize = usage.TotalSize - usage.FreeSize

	if usage.TotalSize > 0 {
		usage.UsedPercent = float64(usage.UsedSize) / float64(usage.TotalSize) * 100
	}

	// Count files and directories (this can be expensive for large shares)
	go func() {
		fileCount, dirCount := s.countFilesAndDirs(sharePath)
		usage.FileCount = fileCount
		usage.DirectoryCount = dirCount
	}()

	// Get last access time from directory stat
	if stat, err := os.Stat(sharePath); err == nil {
		usage.LastAccessed = stat.ModTime().Format(time.RFC3339)
	}

	return usage, nil
}

// countFilesAndDirs counts files and directories in a path (runs in background)
func (s *ShareService) countFilesAndDirs(path string) (int64, int64) {
	var fileCount, dirCount int64

	filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Continue on errors
		}

		if info.IsDir() {
			dirCount++
		} else {
			fileCount++
		}

		return nil
	})

	return fileCount, dirCount
}

// ValidateShareCreateRequest validates a share creation request
func (s *ShareService) ValidateShareCreateRequest(req *requests.ShareCreateRequest) error {
	if req.Name == "" {
		return fmt.Errorf("share name is required")
	}

	// Validate share name (alphanumeric, underscore, hyphen only)
	if !s.isValidShareName(req.Name) {
		return fmt.Errorf("invalid share name: must contain only letters, numbers, underscore, and hyphen")
	}

	// Check if share already exists
	if _, err := s.GetShare(req.Name); err == nil {
		return fmt.Errorf("share '%s' already exists", req.Name)
	}

	// Validate allocator method
	if req.AllocatorMethod != "" {
		validMethods := []string{"high-water", "most-free", "fill-up"}
		if !s.contains(validMethods, req.AllocatorMethod) {
			return fmt.Errorf("invalid allocator method: must be one of %v", validMethods)
		}
	}

	// Validate cache usage
	if req.UseCache != "" {
		validCache := []string{"yes", "no", "only", "prefer"}
		if !s.contains(validCache, req.UseCache) {
			return fmt.Errorf("invalid cache usage: must be one of %v", validCache)
		}
	}

	// Validate security settings
	if req.SMBSecurity != "" {
		validSecurity := []string{"public", "secure", "private"}
		if !s.contains(validSecurity, req.SMBSecurity) {
			return fmt.Errorf("invalid SMB security: must be one of %v", validSecurity)
		}
	}

	if req.NFSSecurity != "" {
		validSecurity := []string{"public", "secure", "private"}
		if !s.contains(validSecurity, req.NFSSecurity) {
			return fmt.Errorf("invalid NFS security: must be one of %v", validSecurity)
		}
	}

	return nil
}

// ValidateShareUpdateRequest validates a share update request
func (s *ShareService) ValidateShareUpdateRequest(req *requests.ShareUpdateRequest) error {
	// Similar validation as create, but name is not required
	if req.AllocatorMethod != "" {
		validMethods := []string{"high-water", "most-free", "fill-up"}
		if !s.contains(validMethods, req.AllocatorMethod) {
			return fmt.Errorf("invalid allocator method: must be one of %v", validMethods)
		}
	}

	if req.UseCache != "" {
		validCache := []string{"yes", "no", "only", "prefer"}
		if !s.contains(validCache, req.UseCache) {
			return fmt.Errorf("invalid cache usage: must be one of %v", validCache)
		}
	}

	if req.SMBSecurity != "" {
		validSecurity := []string{"public", "secure", "private"}
		if !s.contains(validSecurity, req.SMBSecurity) {
			return fmt.Errorf("invalid SMB security: must be one of %v", validSecurity)
		}
	}

	if req.NFSSecurity != "" {
		validSecurity := []string{"public", "secure", "private"}
		if !s.contains(validSecurity, req.NFSSecurity) {
			return fmt.Errorf("invalid NFS security: must be one of %v", validSecurity)
		}
	}

	return nil
}

// CreateShare creates a new share
func (s *ShareService) CreateShare(req *requests.ShareCreateRequest) error {
	configPath := fmt.Sprintf("/boot/config/shares/%s.cfg", req.Name)

	// Ensure shares directory exists
	sharesDir := "/boot/config/shares"
	if err := os.MkdirAll(sharesDir, 0755); err != nil {
		return fmt.Errorf("failed to create shares directory: %v", err)
	}

	// Create share configuration content
	config := s.buildShareConfig(req)

	// Write configuration file
	if err := os.WriteFile(configPath, []byte(config), 0644); err != nil {
		return fmt.Errorf("failed to write share config: %v", err)
	}

	// Create share directory
	sharePath := fmt.Sprintf("/mnt/user/%s", req.Name)
	if err := os.MkdirAll(sharePath, 0755); err != nil {
		// Clean up config file if directory creation fails
		os.Remove(configPath)
		return fmt.Errorf("failed to create share directory: %v", err)
	}

	// Reload SMB configuration if SMB is enabled
	if req.SMBEnabled {
		s.reloadSMBConfig()
	}

	return nil
}

// UpdateShare updates an existing share
func (s *ShareService) UpdateShare(shareName string, req *requests.ShareUpdateRequest) error {
	configPath := fmt.Sprintf("/boot/config/shares/%s.cfg", shareName)

	// Check if share exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("share '%s' not found", shareName)
	}

	// Read existing configuration
	existingShare, err := s.ParseShareConfig(shareName)
	if err != nil {
		return fmt.Errorf("failed to read existing share config: %v", err)
	}

	// Update fields that are provided
	s.updateShareFields(existingShare, req)

	// Build new configuration
	config := s.buildShareConfigFromShare(existingShare)

	// Write updated configuration
	if err := os.WriteFile(configPath, []byte(config), 0644); err != nil {
		return fmt.Errorf("failed to update share config: %v", err)
	}

	// Reload SMB configuration if SMB settings changed
	if req.SMBEnabled || req.SMBSecurity != "" {
		s.reloadSMBConfig()
	}

	return nil
}

// DeleteShare deletes a share
func (s *ShareService) DeleteShare(shareName string) error {
	configPath := fmt.Sprintf("/boot/config/shares/%s.cfg", shareName)

	// Check if share exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("share '%s' not found", shareName)
	}

	// Remove configuration file
	if err := os.Remove(configPath); err != nil {
		return fmt.Errorf("failed to remove share config: %v", err)
	}

	// Note: We don't remove the share directory as it may contain user data
	// The directory will remain at /mnt/user/{shareName} but won't be shared

	// Reload SMB configuration
	s.reloadSMBConfig()

	return nil
}

// Helper functions

// isValidShareName validates a share name
func (s *ShareService) isValidShareName(name string) bool {
	if name == "" || len(name) > 64 {
		return false
	}

	for _, char := range name {
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '_' || char == '-') {
			return false
		}
	}

	return true
}

// contains checks if a slice contains a string
func (s *ShareService) contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// buildShareConfig builds configuration content for a new share
func (s *ShareService) buildShareConfig(req *requests.ShareCreateRequest) string {
	var config strings.Builder

	config.WriteString(fmt.Sprintf("shareName=\"%s\"\n", req.Name))

	if req.Comment != "" {
		config.WriteString(fmt.Sprintf("shareComment=\"%s\"\n", req.Comment))
	}

	// Set defaults if not provided
	allocator := req.AllocatorMethod
	if allocator == "" {
		allocator = "high-water"
	}
	config.WriteString(fmt.Sprintf("shareAllocator=\"%s\"\n", allocator))

	if req.MinimumFreeSpace != "" {
		config.WriteString(fmt.Sprintf("shareFloor=\"%s\"\n", req.MinimumFreeSpace))
	}

	config.WriteString(fmt.Sprintf("shareSplitLevel=\"%d\"\n", req.SplitLevel))

	if len(req.IncludedDisks) > 0 {
		config.WriteString(fmt.Sprintf("shareInclude=\"%s\"\n", strings.Join(req.IncludedDisks, ",")))
	}

	if len(req.ExcludedDisks) > 0 {
		config.WriteString(fmt.Sprintf("shareExclude=\"%s\"\n", strings.Join(req.ExcludedDisks, ",")))
	}

	useCache := req.UseCache
	if useCache == "" {
		useCache = "yes"
	}
	config.WriteString(fmt.Sprintf("shareUseCache=\"%s\"\n", useCache))

	if req.CachePool != "" {
		config.WriteString(fmt.Sprintf("shareCachePool=\"%s\"\n", req.CachePool))
	}

	// SMB settings
	smbExport := "no"
	if req.SMBEnabled {
		smbExport = "yes"
	}
	config.WriteString(fmt.Sprintf("shareExport=\"%s\"\n", smbExport))

	smbSecurity := req.SMBSecurity
	if smbSecurity == "" {
		smbSecurity = "private"
	}
	config.WriteString(fmt.Sprintf("shareSecurity=\"%s\"\n", smbSecurity))

	smbGuests := "no"
	if req.SMBGuests {
		smbGuests = "yes"
	}
	config.WriteString(fmt.Sprintf("shareGuest=\"%s\"\n", smbGuests))

	// NFS settings
	nfsExport := "no"
	if req.NFSEnabled {
		nfsExport = "yes"
	}
	config.WriteString(fmt.Sprintf("shareNFSExport=\"%s\"\n", nfsExport))

	if req.NFSSecurity != "" {
		config.WriteString(fmt.Sprintf("shareNFSSecurity=\"%s\"\n", req.NFSSecurity))
	}

	// AFP settings
	afpExport := "no"
	if req.AFPEnabled {
		afpExport = "yes"
	}
	config.WriteString(fmt.Sprintf("shareAFPExport=\"%s\"\n", afpExport))

	// FTP settings
	ftpExport := "no"
	if req.FTPEnabled {
		ftpExport = "yes"
	}
	config.WriteString(fmt.Sprintf("shareFTPExport=\"%s\"\n", ftpExport))

	return config.String()
}

// buildShareConfigFromShare builds configuration content from a LegacyShare struct
func (s *ShareService) buildShareConfigFromShare(share *LegacyShare) string {
	var config strings.Builder

	config.WriteString(fmt.Sprintf("shareName=\"%s\"\n", share.Name))

	if share.Comment != "" {
		config.WriteString(fmt.Sprintf("shareComment=\"%s\"\n", share.Comment))
	}

	config.WriteString(fmt.Sprintf("shareAllocator=\"%s\"\n", share.AllocatorMethod))

	if share.MinimumFreeSpace != "" {
		config.WriteString(fmt.Sprintf("shareFloor=\"%s\"\n", share.MinimumFreeSpace))
	}

	config.WriteString(fmt.Sprintf("shareSplitLevel=\"%d\"\n", share.SplitLevel))

	if len(share.IncludedDisks) > 0 {
		config.WriteString(fmt.Sprintf("shareInclude=\"%s\"\n", strings.Join(share.IncludedDisks, ",")))
	}

	if len(share.ExcludedDisks) > 0 {
		config.WriteString(fmt.Sprintf("shareExclude=\"%s\"\n", strings.Join(share.ExcludedDisks, ",")))
	}

	config.WriteString(fmt.Sprintf("shareUseCache=\"%s\"\n", share.UseCache))

	if share.CachePool != "" {
		config.WriteString(fmt.Sprintf("shareCachePool=\"%s\"\n", share.CachePool))
	}

	// SMB settings
	smbExport := "no"
	if share.SMBEnabled {
		smbExport = "yes"
	}
	config.WriteString(fmt.Sprintf("shareExport=\"%s\"\n", smbExport))

	config.WriteString(fmt.Sprintf("shareSecurity=\"%s\"\n", share.SMBSecurity))

	smbGuests := "no"
	if share.SMBGuests {
		smbGuests = "yes"
	}
	config.WriteString(fmt.Sprintf("shareGuest=\"%s\"\n", smbGuests))

	// NFS settings
	nfsExport := "no"
	if share.NFSEnabled {
		nfsExport = "yes"
	}
	config.WriteString(fmt.Sprintf("shareNFSExport=\"%s\"\n", nfsExport))

	if share.NFSSecurity != "" {
		config.WriteString(fmt.Sprintf("shareNFSSecurity=\"%s\"\n", share.NFSSecurity))
	}

	// AFP settings
	afpExport := "no"
	if share.AFPEnabled {
		afpExport = "yes"
	}
	config.WriteString(fmt.Sprintf("shareAFPExport=\"%s\"\n", afpExport))

	// FTP settings
	ftpExport := "no"
	if share.FTPEnabled {
		ftpExport = "yes"
	}
	config.WriteString(fmt.Sprintf("shareFTPExport=\"%s\"\n", ftpExport))

	return config.String()
}

// updateShareFields updates share fields from update request
func (s *ShareService) updateShareFields(share *LegacyShare, req *requests.ShareUpdateRequest) {
	if req.Comment != "" {
		share.Comment = req.Comment
	}

	if req.AllocatorMethod != "" {
		share.AllocatorMethod = req.AllocatorMethod
	}

	if req.MinimumFreeSpace != "" {
		share.MinimumFreeSpace = req.MinimumFreeSpace
	}

	if req.SplitLevel > 0 {
		share.SplitLevel = req.SplitLevel
	}

	if len(req.IncludedDisks) > 0 {
		share.IncludedDisks = req.IncludedDisks
	}

	if len(req.ExcludedDisks) > 0 {
		share.ExcludedDisks = req.ExcludedDisks
	}

	if req.UseCache != "" {
		share.UseCache = req.UseCache
	}

	if req.CachePool != "" {
		share.CachePool = req.CachePool
	}

	// SMB settings
	if req.SMBEnabled {
		share.SMBEnabled = req.SMBEnabled
	}

	if req.SMBSecurity != "" {
		share.SMBSecurity = req.SMBSecurity
	}

	if req.SMBGuests {
		share.SMBGuests = req.SMBGuests
	}

	// NFS settings
	if req.NFSEnabled {
		share.NFSEnabled = req.NFSEnabled
	}

	if req.NFSSecurity != "" {
		share.NFSSecurity = req.NFSSecurity
	}

	// AFP settings
	if req.AFPEnabled {
		share.AFPEnabled = req.AFPEnabled
	}

	// FTP settings
	if req.FTPEnabled {
		share.FTPEnabled = req.FTPEnabled
	}

	// Update modification time
	share.ModifiedAt = time.Now().Format(time.RFC3339)
}

// reloadSMBConfig reloads the SMB configuration
func (s *ShareService) reloadSMBConfig() {
	// Execute command to reload SMB configuration
	// This is typically done by restarting the SMB service or reloading config
	exec.Command("/etc/rc.d/rc.samba", "reload").Run()
}
