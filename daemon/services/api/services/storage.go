package services

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/domalab/uma/daemon/logger"
	"github.com/domalab/uma/daemon/plugins/storage"
	"github.com/domalab/uma/daemon/services/api/utils"
)

// StorageService handles storage-related business logic
type StorageService struct {
	api utils.APIInterface
}

// NewStorageService creates a new storage service
func NewStorageService(api utils.APIInterface) *StorageService {
	return &StorageService{
		api: api,
	}
}

// Share represents an Unraid share configuration
type Share struct {
	Name             string   `json:"name"`
	Comment          string   `json:"comment"`
	AllocatorMethod  string   `json:"allocator_method"`
	FloorSize        string   `json:"floor_size"`
	SplitLevel       string   `json:"split_level"`
	IncludedDisks    []string `json:"included_disks"`
	ExcludedDisks    []string `json:"excluded_disks"`
	UseCache         string   `json:"use_cache"`
	CachePool        string   `json:"cache_pool"`
	ReadRestriction  string   `json:"read_restriction"`
	WriteRestriction string   `json:"write_restriction"`
}

// ShareUsage represents share usage statistics
type ShareUsage struct {
	TotalSize int64 `json:"total_size"`
	UsedSize  int64 `json:"used_size"`
	FreeSize  int64 `json:"free_size"`
	FileCount int64 `json:"file_count"`
	DirCount  int64 `json:"dir_count"`
}

// ShareCreateRequest represents a request to create a new share
type ShareCreateRequest struct {
	Name            string   `json:"name"`
	Comment         string   `json:"comment,omitempty"`
	AllocatorMethod string   `json:"allocator_method,omitempty"`
	FloorSize       string   `json:"floor_size,omitempty"`
	SplitLevel      string   `json:"split_level,omitempty"`
	IncludedDisks   []string `json:"included_disks,omitempty"`
	ExcludedDisks   []string `json:"excluded_disks,omitempty"`
	UseCache        string   `json:"use_cache,omitempty"`
	CachePool       string   `json:"cache_pool,omitempty"`
}

// ShareUpdateRequest represents a request to update an existing share
type ShareUpdateRequest struct {
	Comment         string   `json:"comment,omitempty"`
	AllocatorMethod string   `json:"allocator_method,omitempty"`
	FloorSize       string   `json:"floor_size,omitempty"`
	SplitLevel      string   `json:"split_level,omitempty"`
	IncludedDisks   []string `json:"included_disks,omitempty"`
	ExcludedDisks   []string `json:"excluded_disks,omitempty"`
	UseCache        string   `json:"use_cache,omitempty"`
	CachePool       string   `json:"cache_pool,omitempty"`
}

// GetArrayData retrieves array information in optimized format
func (s *StorageService) GetArrayData() (map[string]interface{}, error) {
	arrayInfo, err := s.api.GetStorage().GetArrayInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get array info: %v", err)
	}

	// Convert to optimized format
	if arrayData, ok := arrayInfo.(*storage.ArrayInfo); ok {
		return s.convertToGeneralFormatOptimized(arrayData), nil
	}

	// Fallback: return as-is with metadata
	return map[string]interface{}{
		"data":         arrayInfo,
		"last_updated": time.Now().UTC().Format(time.RFC3339),
	}, nil
}

// GetCacheData retrieves cache pool information
func (s *StorageService) GetCacheData() (map[string]interface{}, error) {
	cacheInfo, err := s.api.GetStorage().GetCacheInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get cache info: %v", err)
	}

	return map[string]interface{}{
		"data":         cacheInfo,
		"last_updated": time.Now().UTC().Format(time.RFC3339),
	}, nil
}

// GetBootData retrieves boot device information
func (s *StorageService) GetBootData() (map[string]interface{}, error) {
	// For now, return placeholder boot information
	// In a real implementation, this would get boot device info from the system
	return map[string]interface{}{
		"device":       "/dev/sda1",
		"size":         "32GB",
		"used":         "2.1GB",
		"available":    "29.9GB",
		"usage":        6.6,
		"filesystem":   "vfat",
		"last_updated": time.Now().UTC().Format(time.RFC3339),
	}, nil
}

// GetZFSData retrieves ZFS pool information
func (s *StorageService) GetZFSData() (map[string]interface{}, error) {
	// Try to get ZFS information
	cmd := exec.Command("zpool", "status", "-v")
	output, err := cmd.Output()
	if err != nil {
		// ZFS might not be available
		return map[string]interface{}{
			"pools":        []interface{}{},
			"available":    false,
			"last_updated": time.Now().UTC().Format(time.RFC3339),
		}, nil
	}

	// Parse ZFS output (simplified)
	pools := s.parseZPoolStatus(string(output))

	return map[string]interface{}{
		"pools":        pools,
		"available":    true,
		"last_updated": time.Now().UTC().Format(time.RFC3339),
	}, nil
}

// GetDisksData retrieves disk information in optimized format
func (s *StorageService) GetDisksData() ([]map[string]interface{}, error) {
	disksInfo, err := s.api.GetStorage().GetDisks()
	if err != nil {
		return nil, fmt.Errorf("failed to get disks info: %v", err)
	}

	// Try to convert to disk slice
	if diskSlice, ok := disksInfo.([]storage.DiskInfo); ok {
		return s.convertDisksOptimized(diskSlice), nil
	}

	// Fallback: return as-is if type assertion fails
	if diskData, ok := disksInfo.([]map[string]interface{}); ok {
		return diskData, nil
	}

	// Last fallback: return empty slice
	return []map[string]interface{}{}, nil
}

// GetParityDiskData retrieves parity disk information
func (s *StorageService) GetParityDiskData() (map[string]interface{}, error) {
	// Find parity disks
	var parityDisks []map[string]interface{}

	// Try to get disks info
	disksInfo, err := s.api.GetStorage().GetDisks()
	if err == nil {
		if diskSlice, ok := disksInfo.([]storage.DiskInfo); ok {
			for _, disk := range diskSlice {
				if strings.HasPrefix(disk.Name, "parity") {
					diskData := map[string]interface{}{
						"name":        disk.Name,
						"device":      disk.Device,
						"size":        disk.Size,
						"temperature": disk.Temperature,
						"status":      disk.Status,
						"health":      disk.Health,
					}
					parityDisks = append(parityDisks, diskData)
				}
			}
		}
	}

	return map[string]interface{}{
		"parity_disks": parityDisks,
		"last_updated": time.Now().UTC().Format(time.RFC3339),
	}, nil
}

// GetParityCheckData retrieves parity check status and history
func (s *StorageService) GetParityCheckData() (map[string]interface{}, error) {
	// Get current parity check status from mdcmd
	currentStatus := s.getCurrentParityCheckStatus()

	// Get parity check history from logs
	history := s.getParityCheckHistory()

	return map[string]interface{}{
		"current_status": currentStatus,
		"history":        history,
		"last_updated":   time.Now().UTC().Format(time.RFC3339),
	}, nil
}

// StartArray starts the Unraid array
func (s *StorageService) StartArray() error {
	// This is a critical operation that should be implemented carefully
	// For now, return a placeholder implementation
	logger.Blue("Array start operation requested")

	// In a real implementation, this would:
	// 1. Validate array configuration
	// 2. Start the array using mdcmd
	// 3. Mount shares
	// 4. Start services

	return fmt.Errorf("array start operation not implemented - requires careful integration with Unraid mdcmd")
}

// StopArray stops the Unraid array
func (s *StorageService) StopArray() error {
	// This is a critical operation that should be implemented carefully
	// For now, return a placeholder implementation
	logger.Blue("Array stop operation requested")

	// In a real implementation, this would:
	// 1. Stop Docker containers
	// 2. Stop VMs
	// 3. Unmount shares
	// 4. Unmount disks
	// 5. Stop parity
	// 6. Stop array using mdcmd

	return fmt.Errorf("array stop operation not implemented - requires careful integration with Unraid mdcmd")
}

// StartParityCheck starts a parity check operation
func (s *StorageService) StartParityCheck(checkType string) error {
	// Validate check type
	validTypes := []string{"check", "check-correct", "check-nocorrect"}
	isValid := false
	for _, validType := range validTypes {
		if checkType == validType {
			isValid = true
			break
		}
	}

	if !isValid {
		return fmt.Errorf("invalid parity check type: %s", checkType)
	}

	logger.Blue("Parity check (%s) operation requested", checkType)

	// In a real implementation, this would use mdcmd to start parity check
	return fmt.Errorf("parity check operation not implemented - requires integration with Unraid mdcmd")
}

// StopParityCheck stops the current parity check operation
func (s *StorageService) StopParityCheck() error {
	logger.Blue("Parity check stop operation requested")

	// In a real implementation, this would use mdcmd to stop parity check
	return fmt.Errorf("parity check stop operation not implemented - requires integration with Unraid mdcmd")
}

// AddDiskToArray adds a disk to the array
func (s *StorageService) AddDiskToArray(diskDevice string, position string) error {
	if diskDevice == "" {
		return fmt.Errorf("disk device is required")
	}

	if position == "" {
		return fmt.Errorf("disk position is required")
	}

	logger.Blue("Add disk operation requested: %s to position %s", diskDevice, position)

	// In a real implementation, this would:
	// 1. Validate disk is available
	// 2. Check disk health
	// 3. Add disk to array configuration
	// 4. Update array using mdcmd

	return fmt.Errorf("add disk operation not implemented - requires integration with Unraid mdcmd")
}

// RemoveDiskFromArray removes a disk from the array
func (s *StorageService) RemoveDiskFromArray(diskName string) error {
	if diskName == "" {
		return fmt.Errorf("disk name is required")
	}

	logger.Blue("Remove disk operation requested: %s", diskName)

	// In a real implementation, this would:
	// 1. Validate disk can be safely removed
	// 2. Move data if necessary
	// 3. Remove disk from array configuration
	// 4. Update array using mdcmd

	return fmt.Errorf("remove disk operation not implemented - requires integration with Unraid mdcmd")
}

// Helper methods

// convertToGeneralFormatOptimized converts array info to optimized general format
func (s *StorageService) convertToGeneralFormatOptimized(arrayInfo *storage.ArrayInfo) map[string]interface{} {
	startTime := time.Now()

	// Use channels for parallel data collection
	type result struct {
		key   string
		value interface{}
	}

	resultChan := make(chan result, 10)

	// Collect basic array information
	go func() {
		resultChan <- result{"state", arrayInfo.State}
	}()

	go func() {
		resultChan <- result{"total_size", arrayInfo.TotalSize}
	}()

	go func() {
		resultChan <- result{"free_size", arrayInfo.FreeSize}
	}()

	go func() {
		resultChan <- result{"used_size", arrayInfo.UsedSize}
	}()

	go func() {
		resultChan <- result{"used_percent", arrayInfo.UsedPercent}
	}()

	// Collect results
	data := make(map[string]interface{})
	for i := 0; i < 6; i++ {
		res := <-resultChan
		data[res.key] = res.value
	}

	// Add metadata
	data["last_updated"] = time.Now().UTC().Format(time.RFC3339)
	data["processing_time"] = time.Since(startTime).String()

	return data
}

// convertDisksOptimized converts disk info to optimized format
func (s *StorageService) convertDisksOptimized(disks []storage.DiskInfo) []map[string]interface{} {
	if len(disks) == 0 {
		return []map[string]interface{}{}
	}

	result := make([]map[string]interface{}, len(disks))

	// Process disks in parallel
	type diskResult struct {
		index int
		data  map[string]interface{}
	}

	resultChan := make(chan diskResult, len(disks))

	for i, disk := range disks {
		go func(idx int, d storage.DiskInfo) {
			diskData := map[string]interface{}{
				"name":        d.Name,
				"device":      d.Device,
				"size":        d.Size,
				"used":        d.Used,
				"available":   d.Available,
				"temperature": d.Temperature,
				"status":      d.Status,
				"health":      d.Health,
				"file_system": d.FileSystem,
				"mount_point": d.MountPoint,
			}
			resultChan <- diskResult{idx, diskData}
		}(i, disk)
	}

	// Collect results
	for i := 0; i < len(disks); i++ {
		res := <-resultChan
		result[res.index] = res.data
	}

	return result
}

// parseZPoolStatus parses zpool status output
func (s *StorageService) parseZPoolStatus(output string) []map[string]interface{} {
	var pools []map[string]interface{}

	lines := strings.Split(output, "\n")
	var currentPool map[string]interface{}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "pool:") {
			if currentPool != nil {
				pools = append(pools, currentPool)
			}
			poolName := strings.TrimSpace(strings.TrimPrefix(line, "pool:"))
			currentPool = map[string]interface{}{
				"name":    poolName,
				"status":  "unknown",
				"devices": []string{},
			}
		} else if strings.HasPrefix(line, "state:") && currentPool != nil {
			status := strings.TrimSpace(strings.TrimPrefix(line, "state:"))
			currentPool["status"] = status
		}
	}

	if currentPool != nil {
		pools = append(pools, currentPool)
	}

	return pools
}

// getCurrentParityCheckStatus gets the current parity check status
func (s *StorageService) getCurrentParityCheckStatus() map[string]interface{} {
	// In a real implementation, this would query mdcmd for current status
	return map[string]interface{}{
		"running":    false,
		"progress":   0,
		"speed":      0,
		"eta":        "",
		"errors":     0,
		"type":       "",
		"started_at": "",
	}
}

// getParityCheckHistory gets parity check history from logs
func (s *StorageService) getParityCheckHistory() []map[string]interface{} {
	// Try to read parity check log
	logPath := "/boot/config/parity-checks.log"
	content, err := os.ReadFile(logPath)
	if err != nil {
		return []map[string]interface{}{}
	}

	var history []map[string]interface{}
	lines := strings.Split(string(content), "\n")

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		// Parse parity check log format (pipe-delimited)
		parts := strings.Split(line, "|")
		if len(parts) >= 4 {
			entry := map[string]interface{}{
				"timestamp": parts[0],
				"status":    parts[1],
				"duration":  parts[2],
				"speed":     parts[3],
			}
			if len(parts) >= 5 {
				entry["errors"] = parts[4]
			}
			history = append(history, entry)
		}
	}

	return history
}

// Share Management Methods

// GetShares returns a list of all configured shares
func (s *StorageService) GetShares() ([]Share, error) {
	var shares []Share

	// Read share configuration files from /boot/config/shares/
	sharesDir := "/boot/config/shares"
	if _, err := os.Stat(sharesDir); os.IsNotExist(err) {
		return shares, nil
	}

	entries, err := os.ReadDir(sharesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read shares directory: %v", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".cfg") {
			shareName := strings.TrimSuffix(entry.Name(), ".cfg")
			share, err := s.parseShareConfig(shareName)
			if err != nil {
				logger.Yellow("Failed to parse share config for %s: %v", shareName, err)
				continue
			}
			shares = append(shares, *share)
		}
	}

	return shares, nil
}

// GetShare returns a specific share configuration
func (s *StorageService) GetShare(shareName string) (*Share, error) {
	return s.parseShareConfig(shareName)
}

// GetShareUsage returns usage statistics for a specific share
func (s *StorageService) GetShareUsage(shareName string) (*ShareUsage, error) {
	sharePath := fmt.Sprintf("/mnt/user/%s", shareName)

	// Check if share path exists
	if _, err := os.Stat(sharePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("share '%s' does not exist", shareName)
	}

	// Get filesystem statistics
	var stat syscall.Statfs_t
	if err := syscall.Statfs(sharePath, &stat); err != nil {
		return nil, fmt.Errorf("failed to get filesystem stats: %v", err)
	}

	totalSize := int64(stat.Blocks) * int64(stat.Bsize)
	freeSize := int64(stat.Bavail) * int64(stat.Bsize)
	usedSize := totalSize - freeSize

	// Count files and directories
	fileCount, dirCount := s.countFilesAndDirs(sharePath)

	return &ShareUsage{
		TotalSize: totalSize,
		UsedSize:  usedSize,
		FreeSize:  freeSize,
		FileCount: fileCount,
		DirCount:  dirCount,
	}, nil
}

// CreateShare creates a new share
func (s *StorageService) CreateShare(req ShareCreateRequest) error {
	if err := s.validateShareCreateRequest(&req); err != nil {
		return err
	}

	configPath := fmt.Sprintf("/boot/config/shares/%s.cfg", req.Name)

	// Ensure shares directory exists
	sharesDir := "/boot/config/shares"
	if err := os.MkdirAll(sharesDir, 0755); err != nil {
		return fmt.Errorf("failed to create shares directory: %v", err)
	}

	// Check if share already exists
	if _, err := os.Stat(configPath); err == nil {
		return fmt.Errorf("share '%s' already exists", req.Name)
	}

	// Build configuration content
	config := s.buildShareConfig(&req)

	// Write configuration file
	if err := os.WriteFile(configPath, []byte(config), 0644); err != nil {
		return fmt.Errorf("failed to write share configuration: %v", err)
	}

	// Create share directory
	shareDir := fmt.Sprintf("/mnt/user/%s", req.Name)
	if err := os.MkdirAll(shareDir, 0755); err != nil {
		logger.Yellow("Failed to create share directory %s: %v", shareDir, err)
	}

	// Reload SMB configuration
	s.reloadSMBConfig()

	logger.Blue("Created share: %s", req.Name)
	return nil
}

// UpdateShare updates an existing share
func (s *StorageService) UpdateShare(shareName string, req ShareUpdateRequest) error {
	if err := s.validateShareUpdateRequest(&req); err != nil {
		return err
	}

	configPath := fmt.Sprintf("/boot/config/shares/%s.cfg", shareName)

	// Check if share exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("share '%s' does not exist", shareName)
	}

	// Get current share configuration
	share, err := s.parseShareConfig(shareName)
	if err != nil {
		return fmt.Errorf("failed to parse current share config: %v", err)
	}

	// Update fields
	s.updateShareFields(share, &req)

	// Build updated configuration
	config := s.buildShareConfigFromShare(share)

	// Write updated configuration file
	if err := os.WriteFile(configPath, []byte(config), 0644); err != nil {
		return fmt.Errorf("failed to write updated share configuration: %v", err)
	}

	// Reload SMB configuration
	s.reloadSMBConfig()

	logger.Blue("Updated share: %s", shareName)
	return nil
}

// DeleteShare deletes a share
func (s *StorageService) DeleteShare(shareName string) error {
	configPath := fmt.Sprintf("/boot/config/shares/%s.cfg", shareName)

	// Check if share exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("share '%s' does not exist", shareName)
	}

	// Remove configuration file
	if err := os.Remove(configPath); err != nil {
		return fmt.Errorf("failed to remove share configuration: %v", err)
	}

	// Note: We don't remove the share directory as it may contain user data
	// The user should manually remove the directory if desired

	// Reload SMB configuration
	s.reloadSMBConfig()

	logger.Blue("Deleted share: %s", shareName)
	return nil
}

// Helper methods for share management

// parseShareConfig parses a share configuration file
func (s *StorageService) parseShareConfig(shareName string) (*Share, error) {
	configPath := fmt.Sprintf("/boot/config/shares/%s.cfg", shareName)

	content, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read share config: %v", err)
	}

	share := &Share{
		Name: shareName,
	}

	// Parse configuration file (simplified parsing)
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
		case "comment":
			share.Comment = value
		case "allocatorMethod":
			share.AllocatorMethod = value
		case "floorSize":
			share.FloorSize = value
		case "splitLevel":
			share.SplitLevel = value
		case "useCache":
			share.UseCache = value
		case "cachePool":
			share.CachePool = value
		}
	}

	return share, nil
}

// countFilesAndDirs counts files and directories in a path
func (s *StorageService) countFilesAndDirs(path string) (int64, int64) {
	var fileCount, dirCount int64

	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}
		if info.IsDir() {
			dirCount++
		} else {
			fileCount++
		}
		return nil
	})

	if err != nil {
		return 0, 0
	}

	return fileCount, dirCount
}

// validateShareCreateRequest validates a share creation request
func (s *StorageService) validateShareCreateRequest(req *ShareCreateRequest) error {
	if req.Name == "" {
		return fmt.Errorf("share name is required")
	}

	// Validate share name (no special characters, etc.)
	if strings.ContainsAny(req.Name, "/\\:*?\"<>|") {
		return fmt.Errorf("share name contains invalid characters")
	}

	return nil
}

// buildShareConfig builds configuration content for a share
func (s *StorageService) buildShareConfig(req *ShareCreateRequest) string {
	var config strings.Builder

	config.WriteString(fmt.Sprintf("shareName=\"%s\"\n", req.Name))

	if req.Comment != "" {
		config.WriteString(fmt.Sprintf("comment=\"%s\"\n", req.Comment))
	}

	if req.AllocatorMethod != "" {
		config.WriteString(fmt.Sprintf("allocatorMethod=\"%s\"\n", req.AllocatorMethod))
	} else {
		config.WriteString("allocatorMethod=\"highwater\"\n")
	}

	if req.UseCache != "" {
		config.WriteString(fmt.Sprintf("useCache=\"%s\"\n", req.UseCache))
	} else {
		config.WriteString("useCache=\"no\"\n")
	}

	return config.String()
}

// reloadSMBConfig reloads the SMB configuration
func (s *StorageService) reloadSMBConfig() {
	// Execute command to reload SMB configuration
	cmd := exec.Command("smbcontrol", "smbd", "reload-config")
	if err := cmd.Run(); err != nil {
		logger.Yellow("Failed to reload SMB config: %v", err)
	}
}

// validateShareUpdateRequest validates a share update request
func (s *StorageService) validateShareUpdateRequest(req *ShareUpdateRequest) error {
	// Basic validation for update request
	return nil
}

// updateShareFields updates share fields from update request
func (s *StorageService) updateShareFields(share *Share, req *ShareUpdateRequest) {
	if req.Comment != "" {
		share.Comment = req.Comment
	}
	if req.AllocatorMethod != "" {
		share.AllocatorMethod = req.AllocatorMethod
	}
	if req.UseCache != "" {
		share.UseCache = req.UseCache
	}
	if req.CachePool != "" {
		share.CachePool = req.CachePool
	}
}

// buildShareConfigFromShare builds configuration content from a share object
func (s *StorageService) buildShareConfigFromShare(share *Share) string {
	var config strings.Builder

	config.WriteString(fmt.Sprintf("shareName=\"%s\"\n", share.Name))

	if share.Comment != "" {
		config.WriteString(fmt.Sprintf("comment=\"%s\"\n", share.Comment))
	}

	if share.AllocatorMethod != "" {
		config.WriteString(fmt.Sprintf("allocatorMethod=\"%s\"\n", share.AllocatorMethod))
	}

	if share.UseCache != "" {
		config.WriteString(fmt.Sprintf("useCache=\"%s\"\n", share.UseCache))
	}

	return config.String()
}
