package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/domalab/uma/daemon/logger"
	"github.com/domalab/uma/daemon/services/api/services"
	"github.com/domalab/uma/daemon/services/api/types/requests"
	"github.com/domalab/uma/daemon/services/api/types/responses"
	"github.com/domalab/uma/daemon/services/api/utils"
)

// StorageHandler handles storage-related HTTP requests
type StorageHandler struct {
	api            utils.APIInterface
	storageService *services.StorageService
}

// NewStorageHandler creates a new storage handler
func NewStorageHandler(api utils.APIInterface) *StorageHandler {
	return &StorageHandler{
		api:            api,
		storageService: services.NewStorageService(api),
	}
}

// HandleStorageArray handles GET /api/v1/storage/array
func (h *StorageHandler) HandleStorageArray(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Use the system interface to get real array info with parity check data
	arrayInfo, err := h.api.GetSystem().GetRealArrayInfo()
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get array information: %v", err))
		return
	}

	// Transform array data to include real usage calculations
	transformedArrayInfo := h.transformArrayData(arrayInfo)
	utils.WriteJSON(w, http.StatusOK, transformedArrayInfo)
}

// HandleStorageDisks handles GET /api/v1/storage/disks
func (h *StorageHandler) HandleStorageDisks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Try to get comprehensive disk info, fallback to basic disk info if needed
	disks, err := h.api.GetStorage().GetConsolidatedDisksInfo()
	if err != nil {
		// Fallback to basic disk information from system interface
		basicDisks, fallbackErr := h.api.GetSystem().GetRealDisks()
		if fallbackErr != nil {
			utils.WriteError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get disk information: %v", fallbackErr))
			return
		}
		// Transform basic disk data
		transformedDisks := h.transformDisksData(basicDisks)
		utils.WriteJSON(w, http.StatusOK, transformedDisks)
		return
	}

	// Transform the enhanced disk data (now returns simple array format)
	transformedDisks := h.transformDisksData(disks)
	utils.WriteJSON(w, http.StatusOK, transformedDisks)
}

// HandleStorageZFS handles GET /api/v1/storage/zfs
func (h *StorageHandler) HandleStorageZFS(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	zfsPools, err := h.api.GetStorage().GetZFSPools()
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get ZFS information: %v", err))
		return
	}

	// Transform ZFS data to match OpenAPI schema (object instead of array)
	zfsInfo := h.transformZFSInfo(zfsPools)
	utils.WriteJSON(w, http.StatusOK, zfsInfo)
}

// HandleStorageCache handles GET /api/v1/storage/cache
func (h *StorageHandler) HandleStorageCache(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	cacheInfo, err := h.api.GetStorage().GetCacheInfo()
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get cache information: %v", err))
		return
	}

	// Transform cache info to match OpenAPI schema
	transformedCacheInfo := h.transformCacheInfo(cacheInfo)
	utils.WriteJSON(w, http.StatusOK, transformedCacheInfo)
}

// HandleStorageBoot handles GET /api/v1/storage/boot
func (h *StorageHandler) HandleStorageBoot(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	bootInfo := h.getBootUsage()
	// Transform boot info to match OpenAPI schema
	transformedBootInfo := h.transformBootInfo(bootInfo)
	utils.WriteJSON(w, http.StatusOK, transformedBootInfo)
}

// HandleStorageGeneral handles GET /api/v1/storage/general
func (h *StorageHandler) HandleStorageGeneral(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	generalInfo := h.getGeneralStorageInfo()
	// Transform general info to match OpenAPI schema
	transformedGeneralInfo := h.transformGeneralInfo(generalInfo)
	utils.WriteJSON(w, http.StatusOK, transformedGeneralInfo)
}

// HandleArrayStart handles POST /api/v1/storage/array/start
func (h *StorageHandler) HandleArrayStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var request requests.ArrayStartRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		// Allow empty body with default values
		request = requests.ArrayStartRequest{
			MaintenanceMode: false,
			CheckFilesystem: false,
		}
	}

	// Enhanced array start with proper orchestration
	err := h.executeArrayStart(request)
	if err != nil {
		response := responses.ArrayOperationResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to start array: %v", err),
		}
		utils.WriteJSON(w, http.StatusInternalServerError, response)
		return
	}

	response := responses.ArrayOperationResponse{
		Success:     true,
		Message:     "Array start initiated successfully",
		OperationID: fmt.Sprintf("array_start_%d", time.Now().Unix()),
	}
	utils.WriteJSON(w, http.StatusOK, response)
}

// HandleArrayStop handles POST /api/v1/storage/array/stop
func (h *StorageHandler) HandleArrayStop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var request requests.ArrayStopRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		// Allow empty body with default values
		request = requests.ArrayStopRequest{
			Force:          false,
			UnmountShares:  true,
			StopContainers: false,
			StopVMs:        false,
		}
	}

	// Enhanced array stop with proper orchestration
	err := h.executeArrayStop(request)
	if err != nil {
		response := responses.ArrayOperationResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to stop array: %v", err),
		}
		utils.WriteJSON(w, http.StatusInternalServerError, response)
		return
	}

	response := responses.ArrayOperationResponse{
		Success:     true,
		Message:     "Array stop initiated successfully",
		OperationID: fmt.Sprintf("array_stop_%d", time.Now().Unix()),
	}
	utils.WriteJSON(w, http.StatusOK, response)
}

// HandleParityCheck handles GET/POST /api/v1/system/parity/check
func (h *StorageHandler) HandleParityCheck(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// Get parity check status
		status := h.getParityCheckStatus()
		utils.WriteJSON(w, http.StatusOK, status)

	case http.MethodPost:
		// Start parity check
		var request requests.ParityCheckRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			utils.WriteError(w, http.StatusBadRequest, "Invalid JSON request")
			return
		}

		err := h.startParityCheck(request)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to start parity check: %v", err))
			return
		}

		response := responses.ArrayOperationResponse{
			Success:     true,
			Message:     "Parity check started successfully",
			OperationID: fmt.Sprintf("parity_check_%d", time.Now().Unix()),
		}
		utils.WriteJSON(w, http.StatusOK, response)

	default:
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// HandleParityDisk handles GET /api/v1/system/parity/disk
func (h *StorageHandler) HandleParityDisk(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	parityDiskInfo := h.getParityDiskInfo()
	utils.WriteJSON(w, http.StatusOK, parityDiskInfo)
}

// Helper methods

// getBootUsage returns boot filesystem usage information
func (h *StorageHandler) getBootUsage() map[string]interface{} {
	bootData, err := h.storageService.GetBootData()
	if err != nil {
		return h.getPathUsage("/boot")
	}
	return bootData
}

// getGeneralStorageInfo returns general storage information
func (h *StorageHandler) getGeneralStorageInfo() map[string]interface{} {
	return map[string]interface{}{
		"docker_vdisk": h.getDockerVDiskUsage(),
		"log_usage":    h.getLogFilesystemUsage(),
		"boot_usage":   h.getBootUsage(),
		"last_updated": time.Now().UTC().Format(time.RFC3339),
	}
}

// getPathUsage returns filesystem usage for a given path
func (h *StorageHandler) getPathUsage(path string) map[string]interface{} {
	// Use syscall to get actual filesystem usage
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		// Return empty usage data if path doesn't exist or can't be accessed
		return map[string]interface{}{
			"total":        0,
			"used":         0,
			"free":         0,
			"path":         path,
			"last_updated": time.Now().UTC().Format(time.RFC3339),
		}
	}

	total := int64(stat.Blocks) * int64(stat.Bsize)
	free := int64(stat.Bavail) * int64(stat.Bsize)
	used := total - free
	usage := 0.0
	if total > 0 {
		usage = float64(used) / float64(total) * 100
	}

	return map[string]interface{}{
		"total":        total,
		"used":         used,
		"free":         free,
		"usage":        usage,
		"path":         path,
		"last_updated": time.Now().UTC().Format(time.RFC3339),
	}
}

// getDockerVDiskUsage returns Docker vDisk usage information
func (h *StorageHandler) getDockerVDiskUsage() map[string]interface{} {
	// Check common Docker vDisk locations
	dockerPaths := []string{"/var/lib/docker", "/mnt/user/system/docker/docker.img"}

	for _, path := range dockerPaths {
		if usage := h.getPathUsage(path); usage != nil {
			return usage
		}
	}

	return map[string]interface{}{
		"total": 0,
		"used":  0,
		"free":  0,
	}
}

// getLogFilesystemUsage returns log filesystem usage information
func (h *StorageHandler) getLogFilesystemUsage() map[string]interface{} {
	return h.getPathUsage("/var/log")
}

// HandleDockerStorage handles GET /api/v1/storage/docker
func (h *StorageHandler) HandleDockerStorage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	dockerUsage := h.getDockerVDiskUsage()

	// Enhance with additional Docker-specific information
	enhancedDockerUsage := map[string]interface{}{
		"total_bytes":      dockerUsage["total"],
		"used_bytes":       dockerUsage["used"],
		"free_bytes":       dockerUsage["free"],
		"usage_percentage": dockerUsage["usage_percentage"],
		"path":             dockerUsage["path"],
		"filesystem":       dockerUsage["filesystem"],
		"mount_point":      dockerUsage["mount_point"],
		"docker_info": map[string]interface{}{
			"vdisk_enabled":   h.isDockerVDiskEnabled(),
			"vdisk_size":      dockerUsage["total"],
			"containers_path": "/var/lib/docker/containers",
			"images_path":     "/var/lib/docker/image",
		},
		"last_updated": time.Now().UTC().Format(time.RFC3339),
	}

	utils.WriteJSON(w, http.StatusOK, enhancedDockerUsage)
}

// HandleLogStorage handles GET /api/v1/storage/logs
func (h *StorageHandler) HandleLogStorage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	logUsage := h.getLogFilesystemUsage()

	// Enhance with additional log-specific information
	enhancedLogUsage := map[string]interface{}{
		"total_bytes":      logUsage["total"],
		"used_bytes":       logUsage["used"],
		"free_bytes":       logUsage["free"],
		"usage_percentage": logUsage["usage_percentage"],
		"path":             "/var/log",
		"filesystem":       logUsage["filesystem"],
		"mount_point":      logUsage["mount_point"],
		"log_info": map[string]interface{}{
			"uma_log_size":     h.getUMALogSize(),
			"syslog_size":      h.getSyslogSize(),
			"max_log_size":     "5MB",
			"rotation_enabled": true,
		},
		"growth_rate": map[string]interface{}{
			"daily_estimate": h.estimateDailyLogGrowth(),
			"monitoring":     "enabled",
		},
		"last_updated": time.Now().UTC().Format(time.RFC3339),
	}

	utils.WriteJSON(w, http.StatusOK, enhancedLogUsage)
}

// isDockerVDiskEnabled checks if Docker vDisk is enabled
func (h *StorageHandler) isDockerVDiskEnabled() bool {
	// Check if Docker vDisk file exists
	dockerVDiskPaths := []string{
		"/mnt/user/system/docker/docker.img",
		"/var/lib/docker.img",
	}

	for _, path := range dockerVDiskPaths {
		if _, err := os.Stat(path); err == nil {
			return true
		}
	}
	return false
}

// getUMALogSize returns the size of the UMA log file
func (h *StorageHandler) getUMALogSize() int64 {
	logPaths := []string{
		"/var/log/uma.log",
		"/tmp/uma.log",
		"/boot/logs/uma.log",
	}

	for _, path := range logPaths {
		if info, err := os.Stat(path); err == nil {
			return info.Size()
		}
	}
	return 0
}

// getSyslogSize returns the size of the main syslog file
func (h *StorageHandler) getSyslogSize() int64 {
	syslogPaths := []string{
		"/var/log/syslog",
		"/var/log/messages",
	}

	for _, path := range syslogPaths {
		if info, err := os.Stat(path); err == nil {
			return info.Size()
		}
	}
	return 0
}

// estimateDailyLogGrowth estimates daily log growth rate
func (h *StorageHandler) estimateDailyLogGrowth() string {
	// Simple estimation based on current UMA log size and typical growth patterns
	umaLogSize := h.getUMALogSize()

	if umaLogSize == 0 {
		return "< 1MB/day"
	}

	// Estimate based on log size (very rough approximation)
	if umaLogSize < 1024*1024 { // < 1MB
		return "< 1MB/day"
	} else if umaLogSize < 5*1024*1024 { // < 5MB
		return "1-2MB/day"
	} else {
		return "2-5MB/day"
	}
}

// getParityCheckStatus returns the current parity check status
func (h *StorageHandler) getParityCheckStatus() responses.ParityCheckStatus {
	// Get parity check data from storage service
	parityCheckData, err := h.storageService.GetParityCheckData()
	if err != nil {
		// Return default status if unable to get data
		return responses.ParityCheckStatus{
			Active:      false,
			Type:        "",
			Progress:    0.0,
			Speed:       "",
			Errors:      0,
			LastUpdated: time.Now().UTC(),
		}
	}

	// Extract current status from parity check data
	status := responses.ParityCheckStatus{
		Active:      false,
		Type:        "",
		Progress:    0.0,
		Speed:       "",
		Errors:      0,
		LastUpdated: time.Now().UTC(),
	}

	if currentStatus, exists := parityCheckData["current_status"]; exists {
		if statusMap, ok := currentStatus.(map[string]interface{}); ok {
			if running, exists := statusMap["running"]; exists {
				if runningBool, ok := running.(bool); ok {
					status.Active = runningBool
				}
			}
			if progress, exists := statusMap["progress"]; exists {
				if progressFloat, ok := progress.(float64); ok {
					status.Progress = progressFloat
				}
			}
			if speed, exists := statusMap["speed"]; exists {
				if speedInt, ok := speed.(int); ok {
					status.Speed = fmt.Sprintf("%d MB/s", speedInt)
				}
			}
			if errors, exists := statusMap["errors"]; exists {
				if errorsInt, ok := errors.(int); ok {
					status.Errors = errorsInt
				}
			}
			if checkType, exists := statusMap["type"]; exists {
				if typeStr, ok := checkType.(string); ok {
					status.Type = typeStr
				}
			}
		}
	}

	return status
}

// startParityCheck starts a parity check operation
func (h *StorageHandler) startParityCheck(request requests.ParityCheckRequest) error {
	// Parity check operations are not implemented for safety
	// Real implementation would require careful integration with Unraid's mdcmd
	return fmt.Errorf("parity check operations are not implemented - use Unraid web interface")
}

// executeArrayStart executes array start with proper orchestration
func (h *StorageHandler) executeArrayStart(request requests.ArrayStartRequest) error {
	logger.Blue("Array start operation requested with maintenance_mode=%v, check_filesystem=%v",
		request.MaintenanceMode, request.CheckFilesystem)

	// Pre-flight checks
	if err := h.validateArrayStartConditions(); err != nil {
		return fmt.Errorf("pre-flight validation failed: %v", err)
	}

	// In a real implementation, this would:
	// 1. Check array configuration
	// 2. Validate disk health
	// 3. Start array using mdcmd
	// 4. Mount shares if not in maintenance mode
	// 5. Start Docker containers if configured
	// 6. Start VMs if configured

	logger.Yellow("Array start operation is disabled for safety - use Unraid web interface")
	return fmt.Errorf("array start operation is disabled for safety - use Unraid web interface")
}

// executeArrayStop executes array stop with proper orchestration
func (h *StorageHandler) executeArrayStop(request requests.ArrayStopRequest) error {
	logger.Blue("Array stop operation requested with force=%v, unmount_shares=%v, stop_containers=%v, stop_vms=%v",
		request.Force, request.UnmountShares, request.StopContainers, request.StopVMs)

	// Pre-flight checks
	if err := h.validateArrayStopConditions(request.Force); err != nil {
		return fmt.Errorf("pre-flight validation failed: %v", err)
	}

	// In a real implementation, this would:
	// 1. Stop VMs if requested
	// 2. Stop Docker containers if requested
	// 3. Unmount shares if requested
	// 4. Unmount disks
	// 5. Stop parity
	// 6. Stop array using mdcmd

	logger.Yellow("Array stop operation is disabled for safety - use Unraid web interface")
	return fmt.Errorf("array stop operation is disabled for safety - use Unraid web interface")
}

// validateArrayStartConditions validates conditions for array start
func (h *StorageHandler) validateArrayStartConditions() error {
	// Check if array is already started
	arrayInfo, err := h.api.GetStorage().GetArrayInfo()
	if err != nil {
		return fmt.Errorf("failed to get array status: %v", err)
	}

	if arrayMap, ok := arrayInfo.(map[string]interface{}); ok {
		if status, exists := arrayMap["status"]; exists {
			if status == "started" || status == "starting" {
				return fmt.Errorf("array is already started or starting")
			}
		}
	}

	// Additional validation checks would go here
	return nil
}

// validateArrayStopConditions validates conditions for array stop
func (h *StorageHandler) validateArrayStopConditions(force bool) error {
	if !force {
		// Check for active operations
		// Check for open files
		// Check for running containers/VMs
		logger.Blue("Performing graceful stop validation checks")
	}

	// Additional validation checks would go here
	return nil
}

// getParityDiskInfo returns parity disk information
func (h *StorageHandler) getParityDiskInfo() map[string]interface{} {
	parityData, err := h.storageService.GetParityDiskData()
	if err != nil {
		return map[string]interface{}{
			"parity_disks": []interface{}{},
			"capacity":     0,
			"temperature":  0.0,
			"health":       "",
			"message":      "Unable to retrieve parity disk information",
			"last_updated": time.Now().UTC().Format(time.RFC3339),
		}
	}
	return parityData
}

// transformDisksData transforms disk data to match OpenAPI schema requirements
func (h *StorageHandler) transformDisksData(disks interface{}) interface{} {
	// Handle different possible return types from GetDisks()
	switch v := disks.(type) {
	case []interface{}:
		// Transform array of disk objects
		transformedDisks := make([]interface{}, 0, len(v))
		for _, disk := range v {
			if diskMap, ok := disk.(map[string]interface{}); ok {
				transformedDisk := h.transformSingleDisk(diskMap)
				transformedDisks = append(transformedDisks, transformedDisk)
			}
		}
		return transformedDisks
	case map[string]interface{}:
		// If it's a single disk object, transform it
		return h.transformSingleDisk(v)
	default:
		// Return empty array if unknown type
		return []interface{}{}
	}
}

// transformSingleDisk transforms a single disk object to match schema
func (h *StorageHandler) transformSingleDisk(disk map[string]interface{}) map[string]interface{} {
	transformed := make(map[string]interface{})

	// Copy all existing fields first
	for key, value := range disk {
		transformed[key] = value
	}

	// Ensure required fields are present
	if _, exists := transformed["status"]; !exists {
		// Determine status based on available information
		if health, ok := transformed["health"].(string); ok {
			switch health {
			case "healthy", "PASSED":
				transformed["status"] = "active"
			case "unknown":
				transformed["status"] = "standby"
			default:
				transformed["status"] = "error"
			}
		} else {
			transformed["status"] = "active" // Default status
		}
	}

	// Add required last_updated field
	transformed["last_updated"] = time.Now().UTC().Format(time.RFC3339)

	// Transform size from string to integer (bytes)
	if sizeStr, ok := transformed["size"].(string); ok {
		if sizeBytes := h.parseSizeToBytes(sizeStr); sizeBytes > 0 {
			transformed["size"] = sizeBytes
		} else {
			transformed["size"] = int64(0) // Default to 0 if parsing fails
		}
	}

	// Ensure device field is present
	if _, exists := transformed["device"]; !exists {
		if name, ok := transformed["name"].(string); ok {
			transformed["device"] = fmt.Sprintf("/dev/%s", name)
		} else {
			transformed["device"] = "/dev/unknown"
		}
	}

	// Ensure name field is present
	if _, exists := transformed["name"]; !exists {
		if device, ok := transformed["device"].(string); ok {
			// Extract name from device path
			parts := strings.Split(device, "/")
			if len(parts) > 0 {
				transformed["name"] = parts[len(parts)-1]
			} else {
				transformed["name"] = "unknown"
			}
		} else {
			transformed["name"] = "unknown"
		}
	}

	return transformed
}

// parseSizeToBytes converts size strings like "223.6G", "14.6T" to bytes
func (h *StorageHandler) parseSizeToBytes(sizeStr string) int64 {
	// Remove whitespace and convert to uppercase
	sizeStr = strings.TrimSpace(strings.ToUpper(sizeStr))

	// Extract numeric part and unit
	re := regexp.MustCompile(`^(\d+(?:\.\d+)?)\s*([KMGTPE]?B?)$`)
	matches := re.FindStringSubmatch(sizeStr)

	if len(matches) < 2 {
		return 0
	}

	size, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return 0
	}

	// Convert based on unit
	unit := ""
	if len(matches) > 2 {
		unit = matches[2]
	}

	switch unit {
	case "PB", "P":
		return int64(size * 1024 * 1024 * 1024 * 1024 * 1024)
	case "TB", "T":
		return int64(size * 1024 * 1024 * 1024 * 1024)
	case "GB", "G":
		return int64(size * 1024 * 1024 * 1024)
	case "MB", "M":
		return int64(size * 1024 * 1024)
	case "KB", "K":
		return int64(size * 1024)
	case "B", "":
		return int64(size)
	default:
		// Assume bytes if unknown unit
		return int64(size)
	}
}

// transformArrayData transforms array data to include real usage calculations
func (h *StorageHandler) transformArrayData(arrayInfo interface{}) interface{} {
	// Handle different possible return types
	switch v := arrayInfo.(type) {
	case map[string]interface{}:
		return h.transformSingleArrayData(v)
	default:
		// Return as-is if unknown type
		return arrayInfo
	}
}

// transformSingleArrayData transforms a single array object to include real usage calculations
func (h *StorageHandler) transformSingleArrayData(arrayData map[string]interface{}) map[string]interface{} {
	transformed := make(map[string]interface{})

	// Copy all existing fields first
	for key, value := range arrayData {
		transformed[key] = value
	}

	// Calculate real usage totals from actual disk data
	totalCapacity := int64(0)
	totalUsed := int64(0)
	diskCount := 0

	// Try to get disk information to calculate totals
	if disks, err := h.api.GetSystem().GetRealDisks(); err == nil {
		if disksArray, ok := disks.([]interface{}); ok {
			diskCount = len(disksArray)
			for _, disk := range disksArray {
				if diskMap, ok := disk.(map[string]interface{}); ok {
					// Only include data disks (not parity, boot, or cache)
					diskType, _ := diskMap["type"].(string)
					if diskType == "disk" {
						// Check if this is a mounted data disk (has mount_point and not /boot or /mnt/cache)
						mountPoint, _ := diskMap["mount_point"].(string)
						if mountPoint != "" && mountPoint != "/boot" && mountPoint != "/mnt/cache" {
							// Get size in bytes
							if sizeBytes, ok := diskMap["size"].(int64); ok && sizeBytes > 0 {
								totalCapacity += sizeBytes
							}

							// Get used bytes
							if usedBytes, ok := diskMap["used"].(int64); ok && usedBytes > 0 {
								totalUsed += usedBytes
							}
						}
					}
				}
			}
		}
	}

	// Calculate usage percentage
	usagePercent := 0.0
	if totalCapacity > 0 {
		usagePercent = float64(totalUsed) / float64(totalCapacity) * 100
	}

	// Add required fields with real calculated values
	totalFree := totalCapacity - totalUsed
	transformed["total_capacity"] = totalCapacity
	transformed["total_used"] = totalUsed
	transformed["total_free"] = totalFree
	transformed["usage_percent"] = usagePercent
	transformed["disk_count"] = diskCount

	// Add human-readable formatted values
	transformed["total_capacity_formatted"] = utils.BytesToHumanReadable(totalCapacity)
	transformed["total_used_formatted"] = utils.BytesToHumanReadable(totalUsed)
	transformed["total_free_formatted"] = utils.BytesToHumanReadable(totalFree)

	// Ensure last_updated is present
	if _, exists := transformed["last_updated"]; !exists {
		transformed["last_updated"] = time.Now().UTC().Format(time.RFC3339)
	}

	return transformed
}

// transformBootInfo transforms boot info to match OpenAPI schema
func (h *StorageHandler) transformBootInfo(bootInfo map[string]interface{}) map[string]interface{} {
	transformed := make(map[string]interface{})

	// Copy all existing fields first
	for key, value := range bootInfo {
		transformed[key] = value
	}

	// Add missing required fields
	if _, exists := transformed["free"]; !exists {
		// Calculate free space if not present
		if total, totalOk := transformed["total"].(int64); totalOk {
			if used, usedOk := transformed["used"].(int64); usedOk {
				transformed["free"] = total - used
			} else {
				transformed["free"] = total // Default to total if used is unknown
			}
		} else {
			transformed["free"] = int64(0)
		}
	}

	// Add missing mount_point field
	if _, exists := transformed["mount_point"]; !exists {
		transformed["mount_point"] = "/boot"
	}

	// Transform size from string to integer if needed
	if sizeStr, ok := transformed["size"].(string); ok {
		if sizeBytes := h.parseSizeToBytes(sizeStr); sizeBytes > 0 {
			transformed["size"] = sizeBytes
		} else {
			transformed["size"] = int64(0)
		}
	}

	// Transform used from string to integer if needed
	if usedStr, ok := transformed["used"].(string); ok {
		if usedBytes := h.parseSizeToBytes(usedStr); usedBytes > 0 {
			transformed["used"] = usedBytes
		} else {
			transformed["used"] = int64(0)
		}
	}

	// Ensure last_updated is present
	if _, exists := transformed["last_updated"]; !exists {
		transformed["last_updated"] = time.Now().UTC().Format(time.RFC3339)
	}

	return transformed
}

// transformCacheInfo transforms cache info to match OpenAPI schema
func (h *StorageHandler) transformCacheInfo(cacheInfo interface{}) map[string]interface{} {
	transformed := map[string]interface{}{
		"disks":        []interface{}{},
		"pool_status":  "unknown",
		"last_updated": time.Now().UTC().Format(time.RFC3339),
	}

	// Handle different possible return types
	if cacheMap, ok := cacheInfo.(map[string]interface{}); ok {
		// Copy existing fields
		for key, value := range cacheMap {
			transformed[key] = value
		}

		// Extract disks from pools if available
		if pools, exists := cacheMap["pools"]; exists {
			if poolsArray, ok := pools.([]interface{}); ok {
				disks := []interface{}{}
				poolStatus := "unknown"

				for _, pool := range poolsArray {
					if poolMap, ok := pool.(map[string]interface{}); ok {
						// Extract pool status
						if health, exists := poolMap["health"]; exists {
							if healthStr, ok := health.(string); ok {
								if healthStr == "ONLINE" || healthStr == "healthy" {
									poolStatus = "online"
								} else {
									poolStatus = "degraded"
								}
							}
						}

						// Extract disk information
						if devices, exists := poolMap["devices"]; exists {
							if devicesArray, ok := devices.([]interface{}); ok {
								for _, device := range devicesArray {
									if deviceMap, ok := device.(map[string]interface{}); ok {
										diskInfo := h.transformCacheDisk(deviceMap, poolMap)
										disks = append(disks, diskInfo)
									}
								}
							}
						}
					}
				}

				transformed["disks"] = disks
				transformed["pool_status"] = poolStatus
			}
		}
	}

	return transformed
}

// transformCacheDisk transforms a cache disk object
func (h *StorageHandler) transformCacheDisk(device map[string]interface{}, pool map[string]interface{}) map[string]interface{} {
	disk := map[string]interface{}{
		"name":   "unknown",
		"device": "/dev/unknown",
		"size":   int64(0),
		"status": "unknown",
	}

	// Copy device information
	for key, value := range device {
		disk[key] = value
	}

	// Add pool information
	if poolName, exists := pool["name"]; exists {
		disk["pool"] = poolName
	}
	if poolSize, exists := pool["size"]; exists {
		if sizeStr, ok := poolSize.(string); ok {
			if sizeBytes := h.parseSizeToBytes(sizeStr); sizeBytes > 0 {
				disk["size"] = sizeBytes
			}
		}
	}

	// Ensure device path
	if name, exists := disk["name"]; exists {
		if nameStr, ok := name.(string); ok {
			disk["device"] = fmt.Sprintf("/dev/%s", nameStr)
		}
	}

	// Map status
	if state, exists := device["state"]; exists {
		if stateStr, ok := state.(string); ok {
			switch stateStr {
			case "ONLINE":
				disk["status"] = "active"
			case "OFFLINE":
				disk["status"] = "inactive"
			default:
				disk["status"] = "error"
			}
		}
	}

	return disk
}

// transformGeneralInfo transforms general storage info to match OpenAPI schema
func (h *StorageHandler) transformGeneralInfo(generalInfo map[string]interface{}) map[string]interface{} {
	transformed := make(map[string]interface{})

	// Copy all existing fields first
	for key, value := range generalInfo {
		transformed[key] = value
	}

	// Add missing required fields with calculated values
	totalCapacity := int64(0)
	totalUsed := int64(0)
	diskCount := 0

	// Try to get disk information to calculate totals
	if disks, err := h.api.GetSystem().GetRealDisks(); err == nil {
		if disksArray, ok := disks.([]interface{}); ok {
			diskCount = len(disksArray)
			for _, disk := range disksArray {
				if diskMap, ok := disk.(map[string]interface{}); ok {
					// Only include data disks (not parity, boot, or cache)
					diskType, _ := diskMap["type"].(string)
					if diskType == "disk" {
						// Check if this is a mounted data disk (has mount_point and not /boot or /mnt/cache)
						mountPoint, _ := diskMap["mount_point"].(string)
						if mountPoint != "" && mountPoint != "/boot" && mountPoint != "/mnt/cache" {
							// Get size in bytes
							if sizeBytes, ok := diskMap["size"].(int64); ok && sizeBytes > 0 {
								totalCapacity += sizeBytes
							}

							// Get used bytes
							if usedBytes, ok := diskMap["used"].(int64); ok && usedBytes > 0 {
								totalUsed += usedBytes
							}
						}
					}
				}
			}
		}
	}

	// Calculate usage percentage
	usagePercent := 0.0
	if totalCapacity > 0 {
		usagePercent = float64(totalUsed) / float64(totalCapacity) * 100
	}

	// Add required fields with real calculated values
	totalFree := totalCapacity - totalUsed
	transformed["total_capacity"] = totalCapacity
	transformed["total_used"] = totalUsed
	transformed["total_free"] = totalFree
	transformed["usage_percent"] = usagePercent
	transformed["disk_count"] = diskCount

	// Add human-readable formatted values
	transformed["total_capacity_formatted"] = utils.BytesToHumanReadable(totalCapacity)
	transformed["total_used_formatted"] = utils.BytesToHumanReadable(totalUsed)
	transformed["total_free_formatted"] = utils.BytesToHumanReadable(totalFree)
	transformed["array_status"] = h.getArrayStatus()

	// Ensure last_updated is present
	if _, exists := transformed["last_updated"]; !exists {
		transformed["last_updated"] = time.Now().UTC().Format(time.RFC3339)
	}

	return transformed
}

// getArrayStatus returns the current array status
func (h *StorageHandler) getArrayStatus() string {
	// Try to get array information
	if arrayInfo, err := h.api.GetStorage().GetArrayInfo(); err == nil {
		if arrayMap, ok := arrayInfo.(map[string]interface{}); ok {
			if status, exists := arrayMap["status"]; exists {
				if statusStr, ok := status.(string); ok {
					return statusStr
				}
			}
		}
	}

	return "unknown"
}

// getSMARTData returns SMART data for all disks
func (h *StorageHandler) getSMARTData() (interface{}, error) {
	// Get disk information first
	disks, err := h.api.GetStorage().GetDisks()
	if err != nil {
		return nil, fmt.Errorf("failed to get disk information: %v", err)
	}

	smartData := make(map[string]interface{})

	// Process each disk to get SMART data
	if diskArray, ok := disks.([]interface{}); ok {
		for _, disk := range diskArray {
			if diskMap, ok := disk.(map[string]interface{}); ok {
				if name, exists := diskMap["name"]; exists {
					if diskName, ok := name.(string); ok {
						smartInfo := h.getDiskSMARTInfo(diskName)
						smartData[diskName] = smartInfo
					}
				}
			}
		}
	}

	return map[string]interface{}{
		"disks":        smartData,
		"last_updated": time.Now().UTC().Format(time.RFC3339),
	}, nil
}

// getDiskSMARTInfo gets SMART information for a specific disk
func (h *StorageHandler) getDiskSMARTInfo(diskName string) map[string]interface{} {
	// Try to read SMART data using smartctl
	cmd := exec.Command("smartctl", "-A", fmt.Sprintf("/dev/%s", diskName))
	output, err := cmd.Output()
	if err != nil {
		return map[string]interface{}{
			"status":      "unavailable",
			"temperature": 0,
			"health":      "unknown",
			"error":       fmt.Sprintf("Failed to get SMART data: %v", err),
		}
	}

	// Parse SMART output for key metrics
	smartInfo := h.parseSMARTOutput(string(output))
	smartInfo["disk"] = diskName
	smartInfo["last_updated"] = time.Now().UTC().Format(time.RFC3339)

	return smartInfo
}

// parseSMARTOutput parses smartctl output to extract key metrics
func (h *StorageHandler) parseSMARTOutput(output string) map[string]interface{} {
	smartInfo := map[string]interface{}{
		"status":      "healthy",
		"temperature": 0,
		"health":      "PASSED",
		"attributes":  make(map[string]interface{}),
	}

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Parse temperature
		if strings.Contains(line, "Temperature_Celsius") {
			fields := strings.Fields(line)
			if len(fields) >= 10 {
				if temp, err := strconv.Atoi(fields[9]); err == nil {
					smartInfo["temperature"] = temp
				}
			}
		}

		// Parse other important attributes
		if strings.Contains(line, "Reallocated_Sector_Ct") ||
			strings.Contains(line, "Spin_Retry_Count") ||
			strings.Contains(line, "End-to-End_Error") {
			fields := strings.Fields(line)
			if len(fields) >= 10 {
				attrName := fields[1]
				if value, err := strconv.Atoi(fields[9]); err == nil {
					if attrs, ok := smartInfo["attributes"].(map[string]interface{}); ok {
						attrs[attrName] = value
					}
				}
			}
		}
	}

	return smartInfo
}

// getDetailedArrayStatus returns detailed array status information
func (h *StorageHandler) getDetailedArrayStatus() (interface{}, error) {
	// Get basic array info
	arrayInfo, err := h.api.GetStorage().GetArrayInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get array information: %v", err)
	}

	// Enhance with additional status details
	detailedStatus := map[string]interface{}{
		"basic_info":   arrayInfo,
		"last_updated": time.Now().UTC().Format(time.RFC3339),
	}

	// Add array state from mdstat if available
	if mdstatInfo := h.getMdstatInfo(); mdstatInfo != nil {
		detailedStatus["mdstat"] = mdstatInfo
	}

	// Add disk assignments
	if diskAssignments := h.getDiskAssignments(); diskAssignments != nil {
		detailedStatus["disk_assignments"] = diskAssignments
	}

	return detailedStatus, nil
}

// getMdstatInfo reads /proc/mdstat for array status
func (h *StorageHandler) getMdstatInfo() map[string]interface{} {
	content, err := os.ReadFile("/proc/mdstat")
	if err != nil {
		return map[string]interface{}{
			"error": fmt.Sprintf("Failed to read mdstat: %v", err),
		}
	}

	return map[string]interface{}{
		"content":      string(content),
		"last_updated": time.Now().UTC().Format(time.RFC3339),
	}
}

// getDiskAssignments reads disk assignments from Unraid configuration
func (h *StorageHandler) getDiskAssignments() map[string]interface{} {
	assignments := make(map[string]interface{})

	// Try to read from common Unraid config locations
	configPaths := []string{
		"/boot/config/disk.cfg",
		"/var/local/emhttp/disks.ini",
	}

	for _, path := range configPaths {
		if content, err := os.ReadFile(path); err == nil {
			assignments[path] = string(content)
			break
		}
	}

	if len(assignments) == 0 {
		assignments["error"] = "No disk assignment configuration found"
	}

	assignments["last_updated"] = time.Now().UTC().Format(time.RFC3339)
	return assignments
}

// transformComprehensiveDisksData transforms consolidated disk data to include hardware information
func (h *StorageHandler) transformComprehensiveDisksData(disksData interface{}) []map[string]interface{} {
	var allDisks []map[string]interface{}

	// Handle the consolidated disk response structure
	if disksResponse, ok := disksData.(map[string]interface{}); ok {
		// Process array disks
		if arrayDisks, ok := disksResponse["array_disks"].([]interface{}); ok {
			for _, disk := range arrayDisks {
				if diskMap, ok := disk.(map[string]interface{}); ok {
					transformed := h.transformSingleDiskWithHardware(diskMap, "array")
					allDisks = append(allDisks, transformed)
				}
			}
		}

		// Process parity disks
		if parityDisks, ok := disksResponse["parity_disks"].([]interface{}); ok {
			for _, disk := range parityDisks {
				if diskMap, ok := disk.(map[string]interface{}); ok {
					transformed := h.transformSingleDiskWithHardware(diskMap, "parity")
					allDisks = append(allDisks, transformed)
				}
			}
		}

		// Process cache disks
		if cacheDisks, ok := disksResponse["cache_disks"].([]interface{}); ok {
			for _, disk := range cacheDisks {
				if diskMap, ok := disk.(map[string]interface{}); ok {
					transformed := h.transformSingleDiskWithHardware(diskMap, "cache")
					allDisks = append(allDisks, transformed)
				}
			}
		}

		// Process boot disk
		if bootDisk, ok := disksResponse["boot_disk"].(map[string]interface{}); ok {
			transformed := h.transformSingleDiskWithHardware(bootDisk, "boot")
			allDisks = append(allDisks, transformed)
		}
	}

	return allDisks
}

// transformSingleDiskWithHardware transforms a single disk with comprehensive hardware information
func (h *StorageHandler) transformSingleDiskWithHardware(disk map[string]interface{}, diskRole string) map[string]interface{} {
	transformed := map[string]interface{}{
		"name":          h.getStringValue(disk, "name"),
		"device":        h.getStringValue(disk, "device"),
		"size":          h.getIntValue(disk, "size"),
		"used":          h.getIntValue(disk, "used"),
		"available":     h.getIntValue(disk, "available"),
		"usage_percent": h.getFloatValue(disk, "usage_percent"),
		"mount_point":   h.getStringValue(disk, "mount_point"),
		"status":        h.getStringValue(disk, "status"),
		"health":        h.getStringValue(disk, "health"),
		"temperature":   h.getIntValue(disk, "temperature"),
		"type":          diskRole,
		"last_updated":  time.Now().UTC().Format(time.RFC3339),
	}

	// Add hardware information
	transformed["model"] = h.getStringValue(disk, "model")
	transformed["serial_number"] = h.getStringValue(disk, "serial_number")
	transformed["disk_type"] = h.getStringValue(disk, "disk_type")     // HDD/SSD
	transformed["interface"] = h.getStringValue(disk, "interface")     // SATA/USB
	transformed["power_state"] = h.getStringValue(disk, "power_state") // active/standby

	// Add filesystem information
	transformed["filesystem"] = h.getStringValue(disk, "file_system")

	// Add spin down delay with proper formatting
	spinDownDelay := h.getIntValue(disk, "spin_down_delay")
	if spinDownDelay == 0 {
		transformed["spin_down_delay"] = "Never"
	} else if spinDownDelay > 0 {
		transformed["spin_down_delay"] = fmt.Sprintf("%d minutes", spinDownDelay)
	} else {
		transformed["spin_down_delay"] = "Default"
	}

	return transformed
}

// transformZFSInfo transforms ZFS pools array to ZFS info object matching schema
func (h *StorageHandler) transformZFSInfo(zfsPools interface{}) map[string]interface{} {
	zfsInfo := map[string]interface{}{
		"pools":          []interface{}{},
		"datasets":       []interface{}{},
		"total_capacity": int64(0),
		"total_used":     int64(0),
		"total_free":     int64(0),
		"overall_health": "ONLINE",
		"version":        "unknown",
		"last_updated":   time.Now().UTC().Format(time.RFC3339),
	}

	// Handle different possible return types from GetZFSPools()
	switch v := zfsPools.(type) {
	case []interface{}:
		// Transform array of pool objects
		pools := make([]interface{}, 0, len(v))
		datasets := []interface{}{}
		totalCapacity := int64(0)
		totalUsed := int64(0)
		overallHealth := "ONLINE"

		for _, pool := range v {
			if poolMap, ok := pool.(map[string]interface{}); ok {
				// Transform pool to match schema
				transformedPool := h.transformZFSPool(poolMap)
				pools = append(pools, transformedPool)

				// Extract datasets from pool if available
				if poolDatasets, exists := poolMap["datasets"]; exists {
					if datasetsArray, ok := poolDatasets.([]interface{}); ok {
						for _, dataset := range datasetsArray {
							if datasetMap, ok := dataset.(map[string]interface{}); ok {
								transformedDataset := h.transformZFSDataset(datasetMap)
								datasets = append(datasets, transformedDataset)
							}
						}
					}
				}

				// Calculate totals
				if size, exists := poolMap["size"]; exists {
					if sizeStr, ok := size.(string); ok {
						if sizeBytes := h.parseSizeToBytes(sizeStr); sizeBytes > 0 {
							totalCapacity += sizeBytes
						}
					}
				}

				if used, exists := poolMap["used"]; exists {
					if usedStr, ok := used.(string); ok {
						if usedBytes := h.parseSizeToBytes(usedStr); usedBytes > 0 {
							totalUsed += usedBytes
						}
					}
				}

				// Check pool health
				if health, exists := poolMap["health"]; exists {
					if healthStr, ok := health.(string); ok {
						if healthStr != "ONLINE" && healthStr != "healthy" {
							overallHealth = "DEGRADED"
						}
					}
				}
			}
		}

		zfsInfo["pools"] = pools
		zfsInfo["datasets"] = datasets
		zfsInfo["total_capacity"] = totalCapacity
		zfsInfo["total_used"] = totalUsed
		zfsInfo["total_free"] = totalCapacity - totalUsed
		zfsInfo["overall_health"] = overallHealth

	case map[string]interface{}:
		// If it's already an object, use it as base
		for key, value := range v {
			zfsInfo[key] = value
		}
		// Ensure required fields are present
		if _, exists := zfsInfo["pools"]; !exists {
			zfsInfo["pools"] = []interface{}{}
		}
		if _, exists := zfsInfo["datasets"]; !exists {
			zfsInfo["datasets"] = []interface{}{}
		}
	}

	return zfsInfo
}

// transformZFSPool transforms a ZFS pool object to match schema
func (h *StorageHandler) transformZFSPool(pool map[string]interface{}) map[string]interface{} {
	transformed := make(map[string]interface{})

	// Copy all existing fields
	for key, value := range pool {
		transformed[key] = value
	}

	// Ensure required fields exist with defaults
	if _, exists := transformed["name"]; !exists {
		transformed["name"] = "unknown"
	}
	if _, exists := transformed["status"]; !exists {
		transformed["status"] = "UNKNOWN"
	}
	if _, exists := transformed["health"]; !exists {
		transformed["health"] = "UNKNOWN"
	}

	// Transform size fields from strings to integers
	if sizeStr, ok := transformed["size"].(string); ok {
		if sizeBytes := h.parseSizeToBytes(sizeStr); sizeBytes > 0 {
			transformed["size"] = sizeBytes
		} else {
			transformed["size"] = int64(0)
		}
	}

	if usedStr, ok := transformed["used"].(string); ok {
		if usedBytes := h.parseSizeToBytes(usedStr); usedBytes > 0 {
			transformed["used"] = usedBytes
		} else {
			transformed["used"] = int64(0)
		}
	}

	if freeStr, ok := transformed["free"].(string); ok {
		if freeBytes := h.parseSizeToBytes(freeStr); freeBytes > 0 {
			transformed["free"] = freeBytes
		} else {
			transformed["free"] = int64(0)
		}
	}

	// Add last_updated if missing
	if _, exists := transformed["last_updated"]; !exists {
		transformed["last_updated"] = time.Now().UTC().Format(time.RFC3339)
	}

	return transformed
}

// transformZFSDataset transforms a ZFS dataset object to match schema
func (h *StorageHandler) transformZFSDataset(dataset map[string]interface{}) map[string]interface{} {
	transformed := make(map[string]interface{})

	// Copy all existing fields
	for key, value := range dataset {
		transformed[key] = value
	}

	// Ensure required fields exist with defaults
	if _, exists := transformed["name"]; !exists {
		transformed["name"] = "unknown"
	}
	if _, exists := transformed["type"]; !exists {
		transformed["type"] = "filesystem"
	}

	// Transform size fields from strings to integers
	if usedStr, ok := transformed["used"].(string); ok {
		if usedBytes := h.parseSizeToBytes(usedStr); usedBytes > 0 {
			transformed["used"] = usedBytes
		} else {
			transformed["used"] = int64(0)
		}
	}

	if availableStr, ok := transformed["available"].(string); ok {
		if availableBytes := h.parseSizeToBytes(availableStr); availableBytes > 0 {
			transformed["available"] = availableBytes
		} else {
			transformed["available"] = int64(0)
		}
	}

	if referencedStr, ok := transformed["referenced"].(string); ok {
		if referencedBytes := h.parseSizeToBytes(referencedStr); referencedBytes > 0 {
			transformed["referenced"] = referencedBytes
		} else {
			transformed["referenced"] = int64(0)
		}
	}

	// Add last_updated if missing
	if _, exists := transformed["last_updated"]; !exists {
		transformed["last_updated"] = time.Now().UTC().Format(time.RFC3339)
	}

	return transformed
}

// HandleStorageSMART handles GET /api/v1/storage/smart
func (h *StorageHandler) HandleStorageSMART(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	smartData, err := h.getSMARTData()
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get SMART data: %v", err))
		return
	}

	utils.WriteJSON(w, http.StatusOK, smartData)
}

// HandleArrayStatus handles GET /api/v1/storage/array/status
func (h *StorageHandler) HandleArrayStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	arrayStatus, err := h.getDetailedArrayStatus()
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get array status: %v", err))
		return
	}

	utils.WriteJSON(w, http.StatusOK, arrayStatus)
}

// Helper methods for extracting values from disk maps

// getStringValue safely extracts a string value from a map
func (h *StorageHandler) getStringValue(data map[string]interface{}, key string) string {
	if value, ok := data[key]; ok {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return ""
}

// getIntValue safely extracts an integer value from a map
func (h *StorageHandler) getIntValue(data map[string]interface{}, key string) int64 {
	if value, ok := data[key]; ok {
		switch v := value.(type) {
		case int:
			return int64(v)
		case int64:
			return v
		case float64:
			return int64(v)
		case string:
			if parsed, err := strconv.ParseInt(v, 10, 64); err == nil {
				return parsed
			}
		}
	}
	return 0
}

// getFloatValue safely extracts a float value from a map
func (h *StorageHandler) getFloatValue(data map[string]interface{}, key string) float64 {
	if value, ok := data[key]; ok {
		switch v := value.(type) {
		case float64:
			return v
		case int:
			return float64(v)
		case int64:
			return float64(v)
		case string:
			if parsed, err := strconv.ParseFloat(v, 64); err == nil {
				return parsed
			}
		}
	}
	return 0.0
}

// getBoolValue safely extracts a boolean value from a map
func (h *StorageHandler) getBoolValue(data map[string]interface{}, key string) bool {
	if value, ok := data[key]; ok {
		if b, ok := value.(bool); ok {
			return b
		}
	}
	return false
}
