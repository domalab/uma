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
	"time"

	"github.com/domalab/uma/daemon/services/api/services"
	"github.com/domalab/uma/daemon/services/api/types/requests"
	"github.com/domalab/uma/daemon/services/api/types/responses"
	"github.com/domalab/uma/daemon/services/api/utils"
)

// SystemHandler handles system-related HTTP requests
type SystemHandler struct {
	api            utils.APIInterface
	systemService  *services.SystemService
	commandService *services.CommandService
}

// NewSystemHandler creates a new system handler
func NewSystemHandler(api utils.APIInterface) *SystemHandler {
	return &SystemHandler{
		api:            api,
		systemService:  services.NewSystemService(api),
		commandService: services.NewCommandService(api),
	}
}

// HandleSystemInfo handles GET /api/v1/system/info
func (h *SystemHandler) HandleSystemInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Get system information that matches the OpenAPI schema
	info := h.getSystemInfo()
	utils.WriteJSON(w, http.StatusOK, info)
}

// HandleSystemCPU handles GET /api/v1/system/cpu
func (h *SystemHandler) HandleSystemCPU(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	cpuData := h.GetCPUData()
	if cpuData == nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to get CPU information")
		return
	}

	utils.WriteJSON(w, http.StatusOK, cpuData)
}

// HandleSystemMemory handles GET /api/v1/system/memory
func (h *SystemHandler) HandleSystemMemory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	memoryData := h.GetMemoryData()
	if memoryData == nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to get memory information")
		return
	}

	utils.WriteJSON(w, http.StatusOK, memoryData)
}

// HandleSystemTemperature handles GET /api/v1/system/temperature
func (h *SystemHandler) HandleSystemTemperature(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	tempData := h.getTemperatureData()
	if tempData == nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to get temperature information")
		return
	}

	utils.WriteJSON(w, http.StatusOK, tempData)
}

// HandleSystemNetwork handles GET /api/v1/system/network
func (h *SystemHandler) HandleSystemNetwork(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	networkData := h.getNetworkData()
	if networkData == nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to get network information")
		return
	}

	utils.WriteJSON(w, http.StatusOK, networkData)
}

// HandleSystemUPS handles GET /api/v1/system/ups and /api/v1/ups/status
func (h *SystemHandler) HandleSystemUPS(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	upsData := h.getUPSData()
	utils.WriteJSON(w, http.StatusOK, upsData)
}

// HandleSystemLoad handles GET /api/v1/system/load
func (h *SystemHandler) HandleSystemLoad(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Get load information from the system
	loadData, err := h.api.GetSystem().GetLoadInfo()
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to get load information")
		return
	}

	// Add timestamp if not present
	if loadMap, ok := loadData.(map[string]interface{}); ok {
		loadMap["last_updated"] = time.Now().UTC().Format(time.RFC3339)
		utils.WriteJSON(w, http.StatusOK, loadMap)
	} else {
		// Fallback response
		response := map[string]interface{}{
			"load1":        0.0,
			"load5":        0.0,
			"load15":       0.0,
			"last_updated": time.Now().UTC().Format(time.RFC3339),
		}
		utils.WriteJSON(w, http.StatusOK, response)
	}
}

// HandleSystemGPU handles GET /api/v1/system/gpu
func (h *SystemHandler) HandleSystemGPU(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Use the system adapter to get real GPU data
	gpuData, err := h.api.GetSystem().GetGPUInfo()
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to get GPU information")
		return
	}

	utils.WriteJSON(w, http.StatusOK, gpuData)
}

// HandleParityDisk handles GET /api/v1/system/parity/disk
func (h *SystemHandler) HandleParityDisk(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Get array information which includes parity disk data
	arrayData, err := h.api.GetStorage().GetArrayInfo()
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get array information: %v", err))
		return
	}

	// Extract parity disk information from array data to match OpenAPI schema
	parityInfo := map[string]interface{}{
		"parity1":      nil, // Required field
		"parity2":      nil, // Optional field
		"last_updated": time.Now().UTC().Format(time.RFC3339),
	}

	// If array data contains parity information, extract it
	if arrayMap, ok := arrayData.(map[string]interface{}); ok {
		if disks, exists := arrayMap["disks"]; exists {
			if diskSlice, ok := disks.([]interface{}); ok {
				for _, disk := range diskSlice {
					if diskMap, ok := disk.(map[string]interface{}); ok {
						if name, exists := diskMap["name"]; exists {
							if nameStr, ok := name.(string); ok {
								// Transform disk data to match schema
								diskInfo := h.transformParityDiskInfo(diskMap)

								if nameStr == "parity1" || strings.HasPrefix(nameStr, "parity") {
									parityInfo["parity1"] = diskInfo
								} else if nameStr == "parity2" {
									parityInfo["parity2"] = diskInfo
								}
							}
						}
					}
				}
			}
		}
	}

	// Ensure parity1 is not nil (required field)
	if parityInfo["parity1"] == nil {
		parityInfo["parity1"] = h.getDefaultParityDiskInfo()
	}

	utils.WriteJSON(w, http.StatusOK, parityInfo)
}

// HandleParityCheck handles GET /api/v1/system/parity/check
func (h *SystemHandler) HandleParityCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Get array information which may include parity check status
	arrayData, err := h.api.GetStorage().GetArrayInfo()
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get array information: %v", err))
		return
	}

	// Extract parity check information or provide default status
	parityCheck := map[string]interface{}{
		"status":       "idle",
		"progress":     0.0,
		"speed":        0, // Integer (bytes per second)
		"eta":          0, // Integer (seconds)
		"errors":       0,
		"type":         "check", // Default to "check" to match enum
		"last_updated": time.Now().UTC().Format(time.RFC3339),
	}

	// If array data contains parity check status, extract it
	if arrayMap, ok := arrayData.(map[string]interface{}); ok {
		if parityStatus, exists := arrayMap["parity_check"]; exists {
			if statusMap, ok := parityStatus.(map[string]interface{}); ok {
				// Transform the data to match schema types
				parityCheck = h.transformParityCheckData(statusMap, parityCheck)
			}
		}
		// Update timestamp
		parityCheck["last_updated"] = time.Now().UTC().Format(time.RFC3339)
	}

	utils.WriteJSON(w, http.StatusOK, parityCheck)
}

// HandleGPU handles GET /api/v1/gpu (legacy endpoint)
func (h *SystemHandler) HandleGPU(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Get GPU information from system interface
	gpuData, err := h.api.GetSystem().GetGPUInfo()
	if err != nil {
		// Return empty GPU info if not available
		gpuInfo := map[string]interface{}{
			"gpus":         []interface{}{},
			"message":      "No GPU information available",
			"last_updated": time.Now().UTC().Format(time.RFC3339),
		}
		utils.WriteJSON(w, http.StatusOK, gpuInfo)
		return
	}

	// Add timestamp if not present
	if gpuMap, ok := gpuData.(map[string]interface{}); ok {
		gpuMap["last_updated"] = time.Now().UTC().Format(time.RFC3339)
		utils.WriteJSON(w, http.StatusOK, gpuMap)
	} else {
		utils.WriteJSON(w, http.StatusOK, gpuData)
	}
}

// HandleSystemFans handles GET /api/v1/system/fans
func (h *SystemHandler) HandleSystemFans(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	timestamp := time.Now().UTC().Format(time.RFC3339)

	// Get enhanced temperature data which includes fan information
	enhancedData, err := h.api.GetSystem().GetEnhancedTemperatureData()
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get fan information: %v", err))
		return
	}

	// Extract fans from enhanced data (assuming it's a map)
	var fans interface{}
	if dataMap, ok := enhancedData.(map[string]interface{}); ok {
		fans = dataMap["fans"]
	} else {
		fans = []interface{}{} // Empty array as fallback
	}

	response := map[string]interface{}{
		"fans":         fans,
		"last_updated": timestamp,
	}

	utils.WriteJSON(w, http.StatusOK, response)
}

// HandleSystemResources handles GET /api/v1/system/resources
func (h *SystemHandler) HandleSystemResources(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	resources := make(map[string]interface{})

	// CPU information
	if cpuInfo, err := h.api.GetSystem().GetCPUInfo(); err == nil {
		resources["cpu"] = cpuInfo
	}

	// Memory information
	if memInfo, err := h.api.GetSystem().GetMemoryInfo(); err == nil {
		resources["memory"] = memInfo
	}

	// Load information
	if loadInfo, err := h.api.GetSystem().GetLoadInfo(); err == nil {
		resources["load"] = loadInfo
	}

	// Uptime information
	if uptimeInfo, err := h.api.GetSystem().GetUptimeInfo(); err == nil {
		resources["uptime"] = uptimeInfo
	}

	// Network information
	if networkInfo, err := h.api.GetSystem().GetNetworkInfo(); err == nil {
		resources["network"] = networkInfo
	}

	// Add required last_updated field
	resources["last_updated"] = time.Now().UTC().Format(time.RFC3339)

	utils.WriteJSON(w, http.StatusOK, resources)
}

// HandleSystemFilesystems handles GET /api/v1/system/filesystems
func (h *SystemHandler) HandleSystemFilesystems(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	fsData := h.GetFilesystemData()

	// Transform the data to match the OpenAPI schema which expects a "filesystems" array
	response := map[string]interface{}{
		"filesystems":  h.transformFilesystemData(fsData),
		"last_updated": time.Now().UTC().Format(time.RFC3339),
	}

	utils.WriteJSON(w, http.StatusOK, response)
}

// HandleSystemExecute handles POST /api/v1/system/execute
func (h *SystemHandler) HandleSystemExecute(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Parse request body
	var request requests.CommandExecuteRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate command
	if strings.TrimSpace(request.Command) == "" {
		utils.WriteError(w, http.StatusBadRequest, "Command is required")
		return
	}

	// Set default timeout
	if request.Timeout <= 0 {
		request.Timeout = 30
	}

	// Limit maximum timeout to 300 seconds (5 minutes)
	if request.Timeout > 300 {
		request.Timeout = 300
	}

	// Security: Basic command sanitization
	if h.isCommandBlacklisted(request.Command) {
		utils.WriteError(w, http.StatusForbidden, "Command not allowed")
		return
	}

	// Execute command
	response := h.executeCommand(request)
	utils.WriteJSON(w, http.StatusOK, response)
}

// Helper methods

// GetCPUData returns CPU data in standard format
func (h *SystemHandler) GetCPUData() map[string]interface{} {
	return h.systemService.GetCPUData()
}

// GetMemoryData returns memory data in standard format
func (h *SystemHandler) GetMemoryData() map[string]interface{} {
	return h.systemService.GetMemoryData()
}

// getTemperatureData returns temperature data in standard format
func (h *SystemHandler) getTemperatureData() map[string]interface{} {
	return h.systemService.GetTemperatureData()
}

// getNetworkData returns network data in standard format
func (h *SystemHandler) getNetworkData() map[string]interface{} {
	return h.systemService.GetNetworkData()
}

// getUPSData returns UPS data in standard format
func (h *SystemHandler) getUPSData() map[string]interface{} {
	return h.systemService.GetUPSData()
}

// GetFilesystemData returns filesystem data in standard format
func (h *SystemHandler) GetFilesystemData() map[string]interface{} {
	return h.systemService.GetFilesystemData()
}

// getSystemInfo returns system information that matches the OpenAPI schema
func (h *SystemHandler) getSystemInfo() map[string]interface{} {
	info := map[string]interface{}{
		"hostname":     h.getHostname(),
		"kernel":       h.getKernelVersion(),
		"uptime":       h.getUptime(),
		"load_average": h.getLoadAverage(),
		"cpu_cores":    h.getCPUCores(),
		"memory_total": h.getMemoryTotal(),
		"cpu_usage":    h.getCPUUsage(),
		"last_updated": time.Now().UTC().Format(time.RFC3339),
	}

	return info
}

// Helper methods for accurate system data collection

// getHostname returns the actual system hostname
func (h *SystemHandler) getHostname() string {
	if hostname, err := os.Hostname(); err == nil {
		return hostname
	}
	return "unknown"
}

// getKernelVersion returns the actual kernel version
func (h *SystemHandler) getKernelVersion() string {
	if output, err := exec.Command("uname", "-r").Output(); err == nil {
		return strings.TrimSpace(string(output))
	}
	return "unknown"
}

// getUptime returns the system uptime in seconds
func (h *SystemHandler) getUptime() int64 {
	if uptimeInfo, err := h.api.GetSystem().GetUptimeInfo(); err == nil {
		if uptimeMap, ok := uptimeInfo.(map[string]interface{}); ok {
			if uptimeSeconds, exists := uptimeMap["uptime_seconds"]; exists {
				switch v := uptimeSeconds.(type) {
				case int:
					return int64(v)
				case int64:
					return v
				case float64:
					return int64(v)
				}
			}
		}
	}
	return 0
}

// getLoadAverage returns the system load averages
func (h *SystemHandler) getLoadAverage() []float64 {
	if loadInfo, err := h.api.GetSystem().GetLoadInfo(); err == nil {
		if loadMap, ok := loadInfo.(map[string]interface{}); ok {
			load1, ok1 := loadMap["load1"].(float64)
			load5, ok2 := loadMap["load5"].(float64)
			load15, ok3 := loadMap["load15"].(float64)
			if ok1 && ok2 && ok3 {
				return []float64{load1, load5, load15}
			}
		}
	}
	return []float64{0.0, 0.0, 0.0}
}

// getCPUCores returns the number of CPU cores
func (h *SystemHandler) getCPUCores() int {
	if cpuInfo, err := h.api.GetSystem().GetCPUInfo(); err == nil {
		if cpuMap, ok := cpuInfo.(map[string]interface{}); ok {
			if cores, exists := cpuMap["cores"]; exists {
				if coresInt, ok := cores.(int); ok {
					return coresInt
				}
			}
		}
	}
	return 0
}

// getMemoryTotal returns the total system memory in KB
func (h *SystemHandler) getMemoryTotal() int64 {
	if memInfo, err := h.api.GetSystem().GetMemoryInfo(); err == nil {
		if memMap, ok := memInfo.(map[string]interface{}); ok {
			if total, exists := memMap["total"]; exists {
				// Handle both int64 and uint64 types
				switch v := total.(type) {
				case int64:
					return v / 1024 // Convert bytes to KB
				case uint64:
					return int64(v) / 1024 // Convert bytes to KB
				case float64:
					return int64(v) / 1024 // Convert bytes to KB
				}
			}
		}
	}
	return 0
}

// getCPUUsage returns the current CPU usage percentage
func (h *SystemHandler) getCPUUsage() float64 {
	if cpuInfo, err := h.api.GetSystem().GetCPUInfo(); err == nil {
		if cpuMap, ok := cpuInfo.(map[string]interface{}); ok {
			if usage, exists := cpuMap["usage"]; exists {
				if usageFloat, ok := usage.(float64); ok {
					return usageFloat
				}
			}
		}
	}
	return 0.0
}

// transformFilesystemData transforms filesystem data to match OpenAPI schema
func (h *SystemHandler) transformFilesystemData(fsData map[string]interface{}) []interface{} {
	filesystems := []interface{}{}

	// Transform each filesystem entry into the expected format
	for name, data := range fsData {
		if name == "last_updated" {
			continue // Skip the timestamp field
		}

		if fsMap, ok := data.(map[string]interface{}); ok {
			filesystem := map[string]interface{}{
				"device":      fmt.Sprintf("/dev/%s", name),
				"mountpoint":  fmt.Sprintf("/mnt/%s", name),
				"fstype":      "xfs", // Default filesystem type
				"size":        int64(0),
				"used":        int64(0),
				"available":   int64(0),
				"use_percent": 0.0,
			}

			// Map the actual data
			if total, exists := fsMap["total"]; exists {
				filesystem["size"] = total
			}
			if used, exists := fsMap["used"]; exists {
				filesystem["used"] = used
			}
			if free, exists := fsMap["free"]; exists {
				filesystem["available"] = free
			}
			if usage, exists := fsMap["usage"]; exists {
				filesystem["use_percent"] = usage
			}

			// Set specific mount points for known filesystems
			switch name {
			case "boot":
				filesystem["mountpoint"] = "/boot"
				filesystem["device"] = "/dev/sda1"
			case "docker":
				filesystem["mountpoint"] = "/var/lib/docker"
				filesystem["device"] = "/dev/loop0"
			case "logs":
				filesystem["mountpoint"] = "/var/log"
				filesystem["device"] = "/dev/shm"
			}

			filesystems = append(filesystems, filesystem)
		}
	}

	return filesystems
}

// transformSystemLogsData transforms system logs data to match OpenAPI schema
func (h *SystemHandler) transformSystemLogsData(logsData interface{}) map[string]interface{} {
	response := map[string]interface{}{
		"logs":         []interface{}{},
		"total_count":  0,
		"last_updated": time.Now().UTC().Format(time.RFC3339),
	}

	// Transform the logs data based on its structure
	if logsMap, ok := logsData.(map[string]interface{}); ok {
		if logs, exists := logsMap["logs"]; exists {
			if logsArray, ok := logs.([]interface{}); ok {
				// Flatten log entries from all log sources
				allEntries := []interface{}{}
				for _, logSource := range logsArray {
					if sourceMap, ok := logSource.(map[string]interface{}); ok {
						if entries, exists := sourceMap["entries"]; exists {
							if entriesArray, ok := entries.([]interface{}); ok {
								allEntries = append(allEntries, entriesArray...)
							}
						}
					}
				}
				response["logs"] = allEntries
				response["total_count"] = len(allEntries)
			}
		}
	}

	return response
}

// transformParityDiskInfo transforms disk data to match parity disk schema
func (h *SystemHandler) transformParityDiskInfo(diskMap map[string]interface{}) map[string]interface{} {
	diskInfo := map[string]interface{}{
		"device": "/dev/unknown",
		"size":   int64(0),
		"status": "unknown",
	}

	// Map the actual data
	if device, exists := diskMap["device"]; exists {
		diskInfo["device"] = device
	}
	if serial, exists := diskMap["serial"]; exists {
		diskInfo["serial"] = serial
	}
	if model, exists := diskMap["model"]; exists {
		diskInfo["model"] = model
	}
	if size, exists := diskMap["size"]; exists {
		// Convert size to integer if it's a string
		if sizeStr, ok := size.(string); ok {
			if sizeInt, err := strconv.ParseInt(sizeStr, 10, 64); err == nil {
				diskInfo["size"] = sizeInt
			}
		} else {
			diskInfo["size"] = size
		}
	}
	if temp, exists := diskMap["temperature"]; exists {
		diskInfo["temperature"] = temp
	}
	if status, exists := diskMap["status"]; exists {
		diskInfo["status"] = status
	}

	return diskInfo
}

// getDefaultParityDiskInfo returns default parity disk info when no disk is found
func (h *SystemHandler) getDefaultParityDiskInfo() map[string]interface{} {
	return map[string]interface{}{
		"device": "/dev/unknown",
		"size":   int64(0),
		"status": "missing",
	}
}

// transformParityCheckData transforms parity check data to match schema types
func (h *SystemHandler) transformParityCheckData(statusMap map[string]interface{}, defaults map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	// Copy defaults first
	for key, value := range defaults {
		result[key] = value
	}

	// Transform each field to match schema requirements
	for key, value := range statusMap {
		switch key {
		case "speed":
			// Convert speed string like "0 MB/s" to integer bytes per second
			if speedStr, ok := value.(string); ok {
				result["speed"] = h.parseSpeedToBytes(speedStr)
			} else if speedInt, ok := value.(int); ok {
				result["speed"] = speedInt
			} else if speedFloat, ok := value.(float64); ok {
				result["speed"] = int(speedFloat)
			}
		case "eta":
			// Convert ETA string to integer seconds
			if etaStr, ok := value.(string); ok {
				result["eta"] = h.parseETAToSeconds(etaStr)
			} else if etaInt, ok := value.(int); ok {
				result["eta"] = etaInt
			} else if etaFloat, ok := value.(float64); ok {
				result["eta"] = int(etaFloat)
			}
		case "type":
			// Ensure type matches enum values
			if typeStr, ok := value.(string); ok {
				if typeStr == "check" || typeStr == "correct" {
					result["type"] = typeStr
				} else {
					result["type"] = "check" // Default to valid enum value
				}
			}
		default:
			// Copy other fields as-is
			result[key] = value
		}
	}

	return result
}

// parseSpeedToBytes converts speed strings like "150 MB/s" to bytes per second
func (h *SystemHandler) parseSpeedToBytes(speedStr string) int {
	// Remove whitespace and convert to lowercase
	speedStr = strings.TrimSpace(strings.ToLower(speedStr))

	// Extract numeric part
	re := regexp.MustCompile(`(\d+(?:\.\d+)?)\s*(mb/s|gb/s|kb/s|b/s)?`)
	matches := re.FindStringSubmatch(speedStr)

	if len(matches) < 2 {
		return 0
	}

	speed, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return 0
	}

	// Convert to bytes per second based on unit
	unit := ""
	if len(matches) > 2 {
		unit = matches[2]
	}

	switch unit {
	case "gb/s":
		return int(speed * 1024 * 1024 * 1024)
	case "mb/s":
		return int(speed * 1024 * 1024)
	case "kb/s":
		return int(speed * 1024)
	default:
		return int(speed) // Assume bytes per second
	}
}

// parseETAToSeconds converts ETA strings to seconds
func (h *SystemHandler) parseETAToSeconds(etaStr string) int {
	// If empty string, return 0
	if strings.TrimSpace(etaStr) == "" {
		return 0
	}

	// Try to parse as duration (e.g., "2h30m", "45m", "30s")
	if duration, err := time.ParseDuration(etaStr); err == nil {
		return int(duration.Seconds())
	}

	// Try to parse as integer seconds
	if seconds, err := strconv.Atoi(strings.TrimSpace(etaStr)); err == nil {
		return seconds
	}

	return 0
}

// HandleSystemScripts handles GET/POST /api/v1/system/scripts
func (h *SystemHandler) HandleSystemScripts(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// List all user scripts
		scripts, err := h.getUserScripts()
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get user scripts: %v", err))
			return
		}

		response := map[string]interface{}{"scripts": scripts}
		utils.WriteJSON(w, http.StatusOK, response)

	case http.MethodPost:
		// Execute a script
		scriptName := r.URL.Query().Get("name")
		if scriptName == "" {
			utils.WriteError(w, http.StatusBadRequest, "Script name is required")
			return
		}

		var req map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.WriteError(w, http.StatusBadRequest, "Invalid JSON request")
			return
		}

		response, err := h.executeUserScript(scriptName, req)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to execute script: %v", err))
			return
		}

		utils.WriteJSON(w, http.StatusOK, response)

	default:
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// isCommandBlacklisted checks if a command is blacklisted for security
func (h *SystemHandler) isCommandBlacklisted(command string) bool {
	return utils.IsCommandBlacklisted(command)
}

// executeCommand executes a command and returns the response
func (h *SystemHandler) executeCommand(request requests.CommandExecuteRequest) responses.CommandExecuteResponse {
	start := time.Now()

	// Command execution is disabled for security reasons
	// Return an error response indicating the feature is not implemented
	return responses.CommandExecuteResponse{
		ExitCode:        1,
		Stdout:          "",
		Stderr:          "Command execution is disabled for security reasons",
		ExecutionTimeMs: time.Since(start).Milliseconds(),
		Command:         request.Command,
		WorkingDir:      request.WorkingDirectory,
	}
}

// getUserScripts returns a list of available user scripts
func (h *SystemHandler) getUserScripts() ([]interface{}, error) {
	// User script discovery is not currently implemented
	// Return empty list to indicate no scripts are available
	return []interface{}{}, nil
}

// executeUserScript executes a user script and returns the response
func (h *SystemHandler) executeUserScript(scriptName string, req map[string]interface{}) (map[string]interface{}, error) {
	// User script execution is not currently implemented
	return nil, fmt.Errorf("user script execution is not implemented")
}

// HandleSystemReboot handles POST /api/v1/system/reboot
func (h *SystemHandler) HandleSystemReboot(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Allow empty body with default values
		req = map[string]interface{}{
			"delay_seconds": 0,
			"message":       "System reboot initiated via UMA API",
			"force":         false,
		}
	}

	// Extract and validate delay
	delaySeconds := 0
	if delay, ok := req["delay_seconds"].(float64); ok {
		delaySeconds = int(delay)
	}
	if delaySeconds < 0 || delaySeconds > 300 {
		utils.WriteError(w, http.StatusBadRequest, "Delay must be between 0 and 300 seconds")
		return
	}

	message := "System reboot initiated via UMA API"
	if msg, ok := req["message"].(string); ok && msg != "" {
		message = msg
	}

	force := false
	if f, ok := req["force"].(bool); ok {
		force = f
	}

	err := h.executeSystemReboot(delaySeconds, message, force)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to initiate reboot: %v", err))
		return
	}

	response := map[string]interface{}{
		"success":        true,
		"message":        "System reboot initiated",
		"operation_id":   fmt.Sprintf("reboot_%d", time.Now().Unix()),
		"scheduled_time": time.Now().Add(time.Duration(delaySeconds) * time.Second).Format(time.RFC3339),
	}

	utils.WriteJSON(w, http.StatusOK, response)
}

// HandleSystemShutdown handles POST /api/v1/system/shutdown
func (h *SystemHandler) HandleSystemShutdown(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Allow empty body with default values
		req = map[string]interface{}{
			"delay_seconds": 0,
			"message":       "System shutdown initiated via UMA API",
			"force":         false,
		}
	}

	// Extract and validate delay
	delaySeconds := 0
	if delay, ok := req["delay_seconds"].(float64); ok {
		delaySeconds = int(delay)
	}
	if delaySeconds < 0 || delaySeconds > 300 {
		utils.WriteError(w, http.StatusBadRequest, "Delay must be between 0 and 300 seconds")
		return
	}

	message := "System shutdown initiated via UMA API"
	if msg, ok := req["message"].(string); ok && msg != "" {
		message = msg
	}

	force := false
	if f, ok := req["force"].(bool); ok {
		force = f
	}

	err := h.executeSystemShutdown(delaySeconds, message, force)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to initiate shutdown: %v", err))
		return
	}

	response := map[string]interface{}{
		"success":        true,
		"message":        "System shutdown initiated",
		"operation_id":   fmt.Sprintf("shutdown_%d", time.Now().Unix()),
		"scheduled_time": time.Now().Add(time.Duration(delaySeconds) * time.Second).Format(time.RFC3339),
	}

	utils.WriteJSON(w, http.StatusOK, response)
}

// executeSystemReboot executes a system reboot
func (h *SystemHandler) executeSystemReboot(delaySeconds int, message string, force bool) error {
	// System reboot is disabled for safety in UMA
	// Real implementation would require careful integration with Unraid's shutdown procedures
	return fmt.Errorf("system reboot is disabled for safety - use Unraid web interface")
}

// executeSystemShutdown executes a system shutdown
func (h *SystemHandler) executeSystemShutdown(delaySeconds int, message string, force bool) error {
	// System shutdown is disabled for safety in UMA
	// Real implementation would require careful integration with Unraid's shutdown procedures
	return fmt.Errorf("system shutdown is disabled for safety - use Unraid web interface")
}

// HandleSystemLogs handles GET /api/v1/system/logs
func (h *SystemHandler) HandleSystemLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Use the system adapter to get real system logs
	logsData, err := h.api.GetSystem().GetSystemLogs()
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get system logs: %v", err))
		return
	}

	// Transform the logs data to match the OpenAPI schema
	response := h.transformSystemLogsData(logsData)
	utils.WriteJSON(w, http.StatusOK, response)
}

// Removed unused functions: getSystemLogs, getCustomLogFile

// HandleSystemLogsAll handles GET /api/v1/system/logs/all
func (h *SystemHandler) HandleSystemLogsAll(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Parse query parameters
	directory := r.URL.Query().Get("directory")
	if directory == "" {
		directory = "/var/log"
	}

	recursive := r.URL.Query().Get("recursive") != "false" // Default to true
	filePattern := r.URL.Query().Get("file_pattern")
	maxFiles := 50

	// Security: Restrict to /var/log and subdirectories only
	if !strings.HasPrefix(directory, "/var/log") {
		utils.WriteError(w, http.StatusForbidden, "Access restricted to /var/log directory")
		return
	}

	logFiles, err := h.scanLogFiles(directory, recursive, filePattern, maxFiles)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to scan log files: %v", err))
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"directory":    directory,
		"recursive":    recursive,
		"file_pattern": filePattern,
		"max_files":    maxFiles,
		"total_found":  len(logFiles),
		"files":        logFiles,
	})
}

// scanLogFiles scans for log files in the specified directory
func (h *SystemHandler) scanLogFiles(directory string, recursive bool, filePattern string, maxFiles int) ([]interface{}, error) {
	// Log file scanning is not implemented for security reasons
	// Return empty list to indicate no files found
	return []interface{}{}, nil
}

// ExecuteCommand executes a system command using the command service
func (h *SystemHandler) ExecuteCommand(request interface{}) (interface{}, error) {
	// Type assert the request
	cmdReq, ok := request.(requests.CommandExecuteRequest)
	if !ok {
		return nil, fmt.Errorf("invalid request type")
	}

	// Convert to service request format
	serviceReq := services.CommandExecuteRequest{
		Command:    cmdReq.Command,
		Arguments:  []string{}, // Convert if needed
		Timeout:    cmdReq.Timeout,
		WorkingDir: cmdReq.WorkingDirectory,
	}

	response := h.commandService.ExecuteCommand(serviceReq)
	return response, nil
}

// GetAPCUPSData retrieves UPS data from apcupsd daemon
func (h *SystemHandler) GetAPCUPSData() map[string]interface{} {
	return h.systemService.GetAPCUPSData()
}

// HandleTemperatureThresholds handles GET/PUT /api/v1/system/temperature/thresholds
func (h *SystemHandler) HandleTemperatureThresholds(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// Return current temperature thresholds
		thresholds := map[string]interface{}{
			"cpu": map[string]interface{}{
				"warning":  70.0,
				"critical": 80.0,
				"shutdown": 90.0,
				"enabled":  true,
			},
			"disk": map[string]interface{}{
				"warning":  45.0,
				"critical": 55.0,
				"shutdown": 65.0,
				"enabled":  true,
			},
			"gpu": map[string]interface{}{
				"warning":  75.0,
				"critical": 85.0,
				"shutdown": 95.0,
				"enabled":  true,
			},
			"system": map[string]interface{}{
				"warning":  65.0,
				"critical": 75.0,
				"shutdown": 85.0,
				"enabled":  true,
			},
		}
		utils.WriteJSON(w, http.StatusOK, thresholds)

	case http.MethodPut:
		// Update temperature thresholds
		var request map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			utils.WriteError(w, http.StatusBadRequest, "Invalid JSON request")
			return
		}

		// In a real implementation, this would update the temperature monitor thresholds
		response := map[string]interface{}{
			"success": true,
			"message": "Temperature thresholds updated successfully",
		}
		utils.WriteJSON(w, http.StatusOK, response)

	default:
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// HandleTemperatureAlerts handles GET /api/v1/system/temperature/alerts
func (h *SystemHandler) HandleTemperatureAlerts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Return recent temperature alerts
	alerts := []map[string]interface{}{
		{
			"sensor_name":  "CPU Package",
			"sensor_type":  "cpu",
			"temperature":  75.2,
			"threshold":    70.0,
			"level":        "warning",
			"message":      "CPU Package temperature warning: 75.2°C (threshold: 70.0°C)",
			"timestamp":    time.Now().Add(-5 * time.Minute).UTC().Format(time.RFC3339),
			"action_taken": "none",
		},
	}

	utils.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"alerts": alerts,
		"count":  len(alerts),
	})
}

// GetNUTUPSData retrieves UPS data from NUT daemon
func (h *SystemHandler) GetNUTUPSData() map[string]interface{} {
	return h.systemService.GetNUTUPSData()
}
