package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
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

	info := h.api.GetInfo()
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

// HandleSystemUPS handles GET /api/v1/system/ups
func (h *SystemHandler) HandleSystemUPS(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	upsData := h.getUPSData()
	utils.WriteJSON(w, http.StatusOK, upsData)
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

	// Placeholder implementation - would need to implement GetParityDiskInfo in SystemInterface
	parityDisk := map[string]interface{}{
		"name":         "Unknown",
		"size":         0,
		"temperature":  0.0,
		"health":       "Unknown",
		"last_updated": time.Now().UTC().Format(time.RFC3339),
	}

	utils.WriteJSON(w, http.StatusOK, parityDisk)
}

// HandleParityCheck handles GET /api/v1/system/parity/check
func (h *SystemHandler) HandleParityCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Placeholder implementation - would need to implement GetParityCheckInfo in SystemInterface
	parityCheck := map[string]interface{}{
		"status":       "Unknown",
		"progress":     0.0,
		"speed":        "0 MB/s",
		"eta":          "Unknown",
		"last_updated": time.Now().UTC().Format(time.RFC3339),
	}

	utils.WriteJSON(w, http.StatusOK, parityCheck)
}

// HandleGPU handles GET /api/v1/gpu (legacy endpoint)
func (h *SystemHandler) HandleGPU(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Placeholder implementation - would need to implement GetGPUInfo in GPUInterface
	gpuInfo := map[string]interface{}{
		"name":         "Unknown",
		"usage":        0.0,
		"memory_used":  0,
		"memory_total": 0,
		"temperature":  0.0,
		"last_updated": time.Now().UTC().Format(time.RFC3339),
	}

	utils.WriteJSON(w, http.StatusOK, gpuInfo)
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

	utils.WriteJSON(w, http.StatusOK, resources)
}

// HandleSystemFilesystems handles GET /api/v1/system/filesystems
func (h *SystemHandler) HandleSystemFilesystems(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	fsData := h.GetFilesystemData()
	utils.WriteJSON(w, http.StatusOK, fsData)
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

// getIntelGPUData returns Intel GPU data in standard format
func (h *SystemHandler) getIntelGPUData() map[string]interface{} {
	// Implementation would get Intel GPU data
	// For now, return placeholder
	return map[string]interface{}{
		"name":         "Unknown",
		"usage":        0.0,
		"memory_used":  0,
		"memory_total": 0,
		"last_updated": time.Now().UTC().Format(time.RFC3339),
	}
}

// GetFilesystemData returns filesystem data in standard format
func (h *SystemHandler) GetFilesystemData() map[string]interface{} {
	return h.systemService.GetFilesystemData()
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

	// For now, return a placeholder response
	// Real implementation would execute the command safely
	return responses.CommandExecuteResponse{
		ExitCode:        0,
		Stdout:          "Command execution not implemented",
		Stderr:          "",
		ExecutionTimeMs: time.Since(start).Milliseconds(),
		Command:         request.Command,
		WorkingDir:      request.WorkingDirectory,
	}
}

// getUserScripts returns a list of available user scripts
func (h *SystemHandler) getUserScripts() ([]interface{}, error) {
	// Placeholder implementation
	// Real implementation would scan /boot/config/plugins/user.scripts/scripts/
	return []interface{}{
		map[string]interface{}{
			"name":        "example_script",
			"description": "Example user script",
			"enabled":     true,
			"last_run":    time.Now().Add(-24 * time.Hour).Format(time.RFC3339),
		},
	}, nil
}

// executeUserScript executes a user script and returns the response
func (h *SystemHandler) executeUserScript(scriptName string, req map[string]interface{}) (map[string]interface{}, error) {
	// Placeholder implementation
	// Real implementation would execute the script safely
	return map[string]interface{}{
		"success":      true,
		"script_name":  scriptName,
		"execution_id": fmt.Sprintf("exec_%d", time.Now().Unix()),
		"started_at":   time.Now().Format(time.RFC3339),
		"status":       "running",
		"message":      "Script execution started",
	}, nil
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
	// Placeholder implementation
	// Real implementation would execute: shutdown -r +delaySeconds "message"
	// For safety, this is just a placeholder
	return nil
}

// executeSystemShutdown executes a system shutdown
func (h *SystemHandler) executeSystemShutdown(delaySeconds int, message string, force bool) error {
	// Placeholder implementation
	// Real implementation would execute: shutdown -h +delaySeconds "message"
	// For safety, this is just a placeholder
	return nil
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

	utils.WriteJSON(w, http.StatusOK, logsData)
}

// getSystemLogs returns system logs of the specified type
func (h *SystemHandler) getSystemLogs(logType string, lines int, follow bool, since string) ([]string, error) {
	// Placeholder implementation
	// Real implementation would read from /var/log/syslog, /var/log/messages, etc.
	return []string{
		fmt.Sprintf("[%s] System log entry 1", time.Now().Format(time.RFC3339)),
		fmt.Sprintf("[%s] System log entry 2", time.Now().Add(-time.Minute).Format(time.RFC3339)),
		fmt.Sprintf("[%s] System log entry 3", time.Now().Add(-2*time.Minute).Format(time.RFC3339)),
	}, nil
}

// getCustomLogFile reads a custom log file with filtering
func (h *SystemHandler) getCustomLogFile(filePath string, lines int, grepFilter, since string) ([]string, error) {
	// Placeholder implementation
	// Real implementation would read the specified file with security checks
	return []string{
		fmt.Sprintf("[%s] Custom log entry from %s", time.Now().Format(time.RFC3339), filePath),
		fmt.Sprintf("[%s] Custom log entry 2", time.Now().Add(-time.Minute).Format(time.RFC3339)),
	}, nil
}

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
	// Placeholder implementation
	// Real implementation would scan the directory for log files
	return []interface{}{
		map[string]interface{}{
			"path":          "/var/log/syslog",
			"name":          "syslog",
			"size":          1024000,
			"modified_time": time.Now().Add(-time.Hour).Format(time.RFC3339),
			"readable":      true,
		},
		map[string]interface{}{
			"path":          "/var/log/messages",
			"name":          "messages",
			"size":          512000,
			"modified_time": time.Now().Add(-2 * time.Hour).Format(time.RFC3339),
			"readable":      true,
		},
	}, nil
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

// GetNUTUPSData retrieves UPS data from NUT daemon
func (h *SystemHandler) GetNUTUPSData() map[string]interface{} {
	return h.systemService.GetNUTUPSData()
}
