package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/domalab/omniraid/daemon/domain"
	"github.com/domalab/omniraid/daemon/dto"
	"github.com/domalab/omniraid/daemon/logger"
	"github.com/domalab/omniraid/daemon/plugins/storage"
)

// HTTPServer handles REST API requests
type HTTPServer struct {
	api    *Api
	server *http.Server
	port   int
}

// NewHTTPServer creates a new HTTP server instance
func NewHTTPServer(api *Api, port int) *HTTPServer {
	return &HTTPServer{
		api:  api,
		port: port,
	}
}

// Start starts the HTTP server
func (h *HTTPServer) Start() error {
	mux := http.NewServeMux()

	// System API routes
	mux.HandleFunc("/api/v1/system/info", h.handleSystemInfo)
	mux.HandleFunc("/api/v1/system/logs", h.handleSystemLogs)
	mux.HandleFunc("/api/v1/system/origin", h.handleSystemOrigin)
	mux.HandleFunc("/api/v1/system/resources", h.handleSystemResources)
	mux.HandleFunc("/api/v1/system/cpu", h.handleSystemCPU)
	mux.HandleFunc("/api/v1/system/memory", h.handleSystemMemory)
	mux.HandleFunc("/api/v1/system/temperature", h.handleSystemTemperature)
	mux.HandleFunc("/api/v1/system/network", h.handleSystemNetwork)
	mux.HandleFunc("/api/v1/system/ups", h.handleSystemUPS)
	mux.HandleFunc("/api/v1/system/gpu", h.handleSystemGPU)
	mux.HandleFunc("/api/v1/system/filesystems", h.handleSystemFilesystems)
	mux.HandleFunc("/api/v1/health", h.handleHealth)

	// Storage API routes
	mux.HandleFunc("/api/v1/storage/array", h.handleStorageArray)
	mux.HandleFunc("/api/v1/storage/cache", h.handleStorageCache)
	mux.HandleFunc("/api/v1/storage/boot", h.handleStorageBoot)
	mux.HandleFunc("/api/v1/storage/ha-format", h.handleStorageHAFormat)

	// GPU API routes
	mux.HandleFunc("/api/v1/gpu", h.handleGPU)

	// Docker API routes
	mux.HandleFunc("/api/v1/docker/containers", h.handleDockerContainers)
	mux.HandleFunc("/api/v1/docker/container/", h.handleDockerContainer)
	mux.HandleFunc("/api/v1/docker/info", h.handleDockerInfo)

	// VM API routes
	mux.HandleFunc("/api/v1/vm/list", h.handleVMList)
	mux.HandleFunc("/api/v1/vm/", h.handleVM)

	// Diagnostics API routes
	mux.HandleFunc("/api/v1/diagnostics/health", h.handleDiagnosticsHealth)
	mux.HandleFunc("/api/v1/diagnostics/info", h.handleDiagnosticsInfo)
	mux.HandleFunc("/api/v1/diagnostics/repair", h.handleDiagnosticsRepair)

	// Configuration routes
	mux.HandleFunc("/api/v1/config", h.handleConfig)

	// Build middleware chain
	handler := h.corsMiddleware(mux)
	handler = h.loggingMiddleware(handler)
	handler = h.api.rateLimiter.RateLimitMiddleware(handler)
	handler = h.api.authService.AuthMiddleware(handler)

	h.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", h.port),
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	logger.Blue("Starting HTTP API server on port %d", h.port)
	
	go func() {
		if err := h.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Yellow("HTTP server error: %v", err)
		}
	}()

	return nil
}

// Stop gracefully stops the HTTP server
func (h *HTTPServer) Stop() error {
	if h.server == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	logger.Blue("Shutting down HTTP API server...")
	return h.server.Shutdown(ctx)
}

// handleSystemInfo handles GET /api/v1/system/info
func (h *HTTPServer) handleSystemInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	info := h.api.getInfo()
	h.writeJSON(w, http.StatusOK, info)
}

// handleSystemLogs handles GET /api/v1/system/logs
func (h *HTTPServer) handleSystemLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	logType := r.URL.Query().Get("type")
	if logType == "" {
		logType = "system"
	}

	logs := h.api.getLogs(logType)
	h.writeJSON(w, http.StatusOK, logs)
}

// handleSystemOrigin handles GET /api/v1/system/origin
func (h *HTTPServer) handleSystemOrigin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	origin := h.api.getOrigin()
	h.writeJSON(w, http.StatusOK, origin)
}

// handleHealth handles GET /api/v1/health
func (h *HTTPServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"version":   h.api.ctx.Config.Version,
		"service":   "omniraid",
	}

	h.writeJSON(w, http.StatusOK, health)
}

// handleConfig handles GET/PUT /api/v1/config
func (h *HTTPServer) handleConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		config := h.api.configManager.GetConfig()
		// Remove sensitive information
		config.Auth.APIKey = ""
		config.Auth.JWTSecret = ""
		h.writeJSON(w, http.StatusOK, config)

	case http.MethodPut:
		var newConfig domain.Config
		if err := json.NewDecoder(r.Body).Decode(&newConfig); err != nil {
			h.writeError(w, http.StatusBadRequest, "Invalid JSON")
			return
		}

		// Preserve sensitive fields if not provided
		currentConfig := h.api.configManager.GetConfig()
		if newConfig.Auth.APIKey == "" {
			newConfig.Auth.APIKey = currentConfig.Auth.APIKey
		}
		if newConfig.Auth.JWTSecret == "" {
			newConfig.Auth.JWTSecret = currentConfig.Auth.JWTSecret
		}

		h.api.configManager.UpdateConfig(newConfig)
		if err := h.api.configManager.Save(); err != nil {
			h.writeError(w, http.StatusInternalServerError, "Failed to save configuration")
			return
		}

		h.writeJSON(w, http.StatusOK, map[string]string{"message": "Configuration updated"})

	default:
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// handleSystemResources handles GET /api/v1/system/resources
func (h *HTTPServer) handleSystemResources(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	resources := make(map[string]interface{})

	// CPU information
	if cpuInfo, err := h.api.system.GetCPUInfo(); err == nil {
		resources["cpu"] = cpuInfo
	}

	// Memory information
	if memInfo, err := h.api.system.GetMemoryInfo(); err == nil {
		resources["memory"] = memInfo
	}

	// Load information
	if loadInfo, err := h.api.system.GetLoadInfo(); err == nil {
		resources["load"] = loadInfo
	}

	// Uptime information
	if uptimeInfo, err := h.api.system.GetUptimeInfo(); err == nil {
		resources["uptime"] = uptimeInfo
	}

	// Network information
	if networkInfo, err := h.api.system.GetNetworkInfo(); err == nil {
		resources["network"] = networkInfo
	}

	h.writeJSON(w, http.StatusOK, resources)
}

// handleSystemCPU handles GET /api/v1/system/cpu
func (h *HTTPServer) handleSystemCPU(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	cpuData := h.getCPUDataForHA()
	if cpuData == nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to get CPU information")
		return
	}

	h.writeJSON(w, http.StatusOK, cpuData)
}

// handleSystemMemory handles GET /api/v1/system/memory
func (h *HTTPServer) handleSystemMemory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	memoryData := h.getMemoryDataForHA()
	if memoryData == nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to get memory information")
		return
	}

	h.writeJSON(w, http.StatusOK, memoryData)
}

// handleSystemTemperature handles GET /api/v1/system/temperature
func (h *HTTPServer) handleSystemTemperature(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	tempData := h.getTemperatureDataForHA()
	if tempData == nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to get temperature information")
		return
	}

	h.writeJSON(w, http.StatusOK, tempData)
}

// handleSystemNetwork handles GET /api/v1/system/network
func (h *HTTPServer) handleSystemNetwork(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	networkData := h.getNetworkDataForHA()
	if networkData == nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to get network information")
		return
	}

	h.writeJSON(w, http.StatusOK, networkData)
}

// handleSystemUPS handles GET /api/v1/system/ups
func (h *HTTPServer) handleSystemUPS(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	upsData := h.getUPSDataForHA()
	h.writeJSON(w, http.StatusOK, upsData)
}

// handleSystemGPU handles GET /api/v1/system/gpu
func (h *HTTPServer) handleSystemGPU(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	gpuData := h.getIntelGPUDataForHA()
	if gpuData == nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to get GPU information")
		return
	}

	h.writeJSON(w, http.StatusOK, gpuData)
}

// handleSystemFilesystems handles GET /api/v1/system/filesystems
func (h *HTTPServer) handleSystemFilesystems(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	fsData := h.getFilesystemDataForHA()
	h.writeJSON(w, http.StatusOK, fsData)
}

// handleStorageArray handles GET /api/v1/storage/array
func (h *HTTPServer) handleStorageArray(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	arrayInfo, err := h.api.storage.GetArrayInfo()
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get array info: %v", err))
		return
	}

	h.writeJSON(w, http.StatusOK, arrayInfo)
}

// handleStorageCache handles GET /api/v1/storage/cache
func (h *HTTPServer) handleStorageCache(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	cacheInfo, err := h.api.storage.GetCacheInfo()
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get cache info: %v", err))
		return
	}

	h.writeJSON(w, http.StatusOK, cacheInfo)
}

// handleStorageBoot handles GET /api/v1/storage/boot
func (h *HTTPServer) handleStorageBoot(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	bootInfo, err := h.api.storage.GetBootDiskInfo()
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get boot disk info: %v", err))
		return
	}

	h.writeJSON(w, http.StatusOK, bootInfo)
}

// handleStorageHAFormat handles GET /api/v1/storage/ha-format
// Returns storage data in Home Assistant integration compatible format
func (h *HTTPServer) handleStorageHAFormat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Get array information
	arrayInfo, err := h.api.storage.GetArrayInfo()
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get array info: %v", err))
		return
	}

	// Convert to HA-compatible format
	haFormat := h.convertToHAFormat(arrayInfo)
	h.writeJSON(w, http.StatusOK, haFormat)
}

// convertToHAFormat converts OmniRaid data to complete Home Assistant integration format
func (h *HTTPServer) convertToHAFormat(arrayInfo *storage.ArrayInfo) map[string]interface{} {
	systemStats := map[string]interface{}{
		"array_usage": map[string]interface{}{
			"total":      arrayInfo.TotalSize,
			"used":       arrayInfo.UsedSize,
			"free":       arrayInfo.FreeSize,
			"percentage": arrayInfo.UsedPercent,
		},
		"array_state": map[string]interface{}{
			"state":        arrayInfo.State,
			"num_disks":    arrayInfo.NumDisks,
			"num_devices":  arrayInfo.NumDevices,
			"num_parity":   arrayInfo.NumParity,
			"synced":       true, // Assume synced if started
			"sync_action":  nil,
			"sync_progress": 0,
			"sync_errors":  0,
			"num_disabled": 0,
			"num_invalid":  0,
			"num_missing":  0,
		},
		"individual_disks": h.convertDisksToHAFormat(arrayInfo.Disks),
	}

	// Add CPU usage data
	if cpuData := h.getCPUDataForHA(); cpuData != nil {
		systemStats["cpu_usage"] = cpuData
	}

	// Add memory usage data
	if memoryData := h.getMemoryDataForHA(); memoryData != nil {
		systemStats["memory_usage"] = memoryData
	}

	// Add temperature sensor data
	if tempData := h.getTemperatureDataForHA(); tempData != nil {
		systemStats["temperature_data"] = tempData
	}

	// Add network statistics
	if networkData := h.getNetworkDataForHA(); networkData != nil {
		systemStats["network_stats"] = networkData
	}

	// Add UPS information
	if upsData := h.getUPSDataForHA(); upsData != nil {
		systemStats["ups_info"] = upsData
	}

	// Add Intel GPU data
	if gpuData := h.getIntelGPUDataForHA(); gpuData != nil {
		systemStats["intel_gpu"] = gpuData
	}

	// Add filesystem usage data
	if fsData := h.getFilesystemDataForHA(); fsData != nil {
		systemStats["docker_vdisk"] = fsData["docker_vdisk"]
		systemStats["log_filesystem"] = fsData["log_filesystem"]
		systemStats["boot_usage"] = fsData["boot_usage"]
	}

	result := map[string]interface{}{
		"system_stats":   systemStats,
		"disk_mappings":  h.createDiskMappings(arrayInfo.Disks),
	}

	// Add Docker container data
	if dockerData := h.getDockerDataForHA(); dockerData != nil {
		result["docker_containers"] = dockerData
	}

	// Add VM data
	if vmData := h.getVMDataForHA(); vmData != nil {
		result["vms"] = vmData
	}

	return result
}

// convertDisksToHAFormat converts disk information to HA format
func (h *HTTPServer) convertDisksToHAFormat(disks []storage.DiskInfo) []map[string]interface{} {
	var result []map[string]interface{}

	for _, disk := range disks {
		if disk.Status == "not_present" || disk.Status == "disabled" {
			continue // Skip non-present disks
		}

		// Determine power state for HA format
		powerState := "active"
		if disk.PowerState == "standby" || disk.PowerState == "sleeping" {
			powerState = "standby"
		}

		diskData := map[string]interface{}{
			"name":        disk.Name,
			"device":      disk.Device,
			"total":       disk.Size,
			"used":        disk.Used,
			"free":        disk.Available,
			"percentage":  disk.UsedPercent,
			"mount_point": disk.MountPoint,
			"filesystem":  disk.FileSystem,
			"state":       powerState,
			"temperature": disk.Temperature,
			"health":      disk.Health,
			"spin_down_delay": disk.SpinDownDelay,
			"smart_data": map[string]interface{}{
				"serial_number": disk.SerialNumber,
				"model_name":    disk.Model,
				"rotation_rate": h.getRotationRate(disk.DiskType),
			},
		}

		result = append(result, diskData)
	}

	return result
}

// createDiskMappings creates disk mappings for HA format
func (h *HTTPServer) createDiskMappings(disks []storage.DiskInfo) map[string]interface{} {
	mappings := make(map[string]interface{})

	for _, disk := range disks {
		if disk.Status == "not_present" || disk.Status == "disabled" {
			continue
		}

		mappings[disk.Name] = map[string]interface{}{
			"device":     disk.Device,
			"serial":     disk.SerialNumber,
			"filesystem": disk.FileSystem,
			"fsSize":     disk.Size,
			"fsUsed":     disk.Used,
		}
	}

	return mappings
}

// getRotationRate returns rotation rate based on disk type (for HA compatibility)
func (h *HTTPServer) getRotationRate(diskType string) int {
	switch diskType {
	case "SSD", "NVMe":
		return 0 // SSDs have 0 rotation rate
	case "HDD":
		return 7200 // Assume 7200 RPM for HDDs
	default:
		return -1 // Unknown
	}
}

// getCPUDataForHA returns CPU data in HA integration format
func (h *HTTPServer) getCPUDataForHA() map[string]interface{} {
	cpuInfo, err := h.api.system.GetCPUInfo()
	if err != nil {
		return nil
	}

	return map[string]interface{}{
		"usage":       cpuInfo.Usage,
		"cores":       cpuInfo.Cores,
		"threads":     cpuInfo.Threads,
		"temperature": cpuInfo.Temperature,
		"frequency":   cpuInfo.Frequency,
		"model":       cpuInfo.Model,
	}
}

// getMemoryDataForHA returns memory data in HA integration format
func (h *HTTPServer) getMemoryDataForHA() map[string]interface{} {
	memInfo, err := h.api.system.GetMemoryInfo()
	if err != nil {
		return nil
	}

	return map[string]interface{}{
		"total":      memInfo.Total,
		"used":       memInfo.Used,
		"free":       memInfo.Free,
		"available":  memInfo.Available,
		"buffers":    memInfo.Buffers,
		"cached":     memInfo.Cached,
		"percentage": memInfo.UsedPercent,
	}
}

// getTemperatureDataForHA returns temperature sensor data in HA integration format
func (h *HTTPServer) getTemperatureDataForHA() map[string]interface{} {
	sensors := make([]map[string]interface{}, 0)

	// Get CPU temperature from system plugin
	if cpuInfo, err := h.api.system.GetCPUInfo(); err == nil && cpuInfo.Temperature > 0 {
		sensors = append(sensors, map[string]interface{}{
			"name":  "CPU",
			"value": cpuInfo.Temperature,
			"unit":  "°C",
		})
	}

	// Get additional sensors from sensor plugin
	if h.api.sensor != nil {
		samples := h.api.sensor.GetSamples()
		for _, sample := range samples {
			if sample.Unit == "°C" {
				sensors = append(sensors, map[string]interface{}{
					"name":  sample.Key,
					"value": sample.Value,
					"unit":  sample.Unit,
				})
			}
		}
	}

	if len(sensors) == 0 {
		return nil
	}

	return map[string]interface{}{
		"sensors": sensors,
	}
}

// getNetworkDataForHA returns network statistics in HA integration format
func (h *HTTPServer) getNetworkDataForHA() map[string]interface{} {
	networkInfo, err := h.api.system.GetNetworkInfo()
	if err != nil {
		return nil
	}

	interfaces := make([]map[string]interface{}, 0)
	for _, netInfo := range networkInfo {
		interfaces = append(interfaces, map[string]interface{}{
			"name":     netInfo.Interface,
			"rx_bytes": netInfo.BytesRecv,
			"tx_bytes": netInfo.BytesSent,
			"rx_packets": netInfo.PacketsRecv,
			"tx_packets": netInfo.PacketsSent,
			"rx_errors":  netInfo.ErrorsRecv,
			"tx_errors":  netInfo.ErrorsSent,
		})
	}

	return map[string]interface{}{
		"interfaces": interfaces,
	}
}

// getUPSDataForHA returns UPS information in HA integration format
func (h *HTTPServer) getUPSDataForHA() map[string]interface{} {
	if h.api.ups == nil {
		return map[string]interface{}{
			"status":         "not_available",
			"battery_charge": 0,
			"runtime":        0,
		}
	}

	samples := h.api.ups.GetStatus()
	upsData := map[string]interface{}{
		"status":         "unknown",
		"battery_charge": 0,
		"runtime":        0,
	}

	for _, sample := range samples {
		switch sample.Key {
		case "UPS STATUS":
			if sample.Condition == "green" {
				upsData["status"] = "online"
			} else if sample.Condition == "red" {
				upsData["status"] = "offline"
			} else {
				upsData["status"] = "unknown"
			}
		case "UPS CHARGE":
			if charge, err := strconv.ParseFloat(sample.Value, 64); err == nil {
				upsData["battery_charge"] = charge
			}
		case "UPS RUNTIME":
			if runtime, err := strconv.ParseFloat(sample.Value, 64); err == nil {
				upsData["runtime"] = runtime
			}
		}
	}

	return upsData
}

// getIntelGPUDataForHA returns Intel GPU data in HA integration format
func (h *HTTPServer) getIntelGPUDataForHA() map[string]interface{} {
	gpuInfo, err := h.api.gpu.GetGPUInfo()
	if err != nil {
		return nil
	}

	// Look for Intel GPU
	for _, gpu := range gpuInfo {
		if strings.Contains(strings.ToLower(gpu.Name), "intel") {
			return map[string]interface{}{
				"usage":       gpu.UtilizationGPU,
				"temperature": gpu.Temperature,
				"name":        gpu.Name,
				"memory_used": gpu.MemoryUsed,
				"memory_total": gpu.MemoryTotal,
			}
		}
	}

	return map[string]interface{}{
		"usage":       0,
		"temperature": 0,
		"name":        "No Intel GPU detected",
	}
}

// getFilesystemDataForHA returns filesystem usage data in HA integration format
func (h *HTTPServer) getFilesystemDataForHA() map[string]interface{} {
	result := make(map[string]interface{})

	// Docker vDisk usage
	if dockerVDisk := h.getDockerVDiskUsage(); dockerVDisk != nil {
		result["docker_vdisk"] = dockerVDisk
	}

	// Log filesystem usage
	if logFS := h.getLogFilesystemUsage(); logFS != nil {
		result["log_filesystem"] = logFS
	}

	// Boot usage
	if bootUsage := h.getBootUsage(); bootUsage != nil {
		result["boot_usage"] = bootUsage
	}

	return result
}

// getDockerVDiskUsage returns Docker vDisk usage information
func (h *HTTPServer) getDockerVDiskUsage() map[string]interface{} {
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
func (h *HTTPServer) getLogFilesystemUsage() map[string]interface{} {
	return h.getPathUsage("/var/log")
}

// getBootUsage returns boot filesystem usage information
func (h *HTTPServer) getBootUsage() map[string]interface{} {
	return h.getPathUsage("/boot")
}

// getPathUsage returns filesystem usage for a given path
func (h *HTTPServer) getPathUsage(path string) map[string]interface{} {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return map[string]interface{}{
			"total": 0,
			"used":  0,
			"free":  0,
		}
	}

	total := stat.Blocks * uint64(stat.Bsize)
	free := stat.Bavail * uint64(stat.Bsize)
	used := total - free

	return map[string]interface{}{
		"total": total,
		"used":  used,
		"free":  free,
	}
}

// getDockerDataForHA returns Docker container data in HA integration format
func (h *HTTPServer) getDockerDataForHA() []map[string]interface{} {
	containers, err := h.api.docker.ListContainers(false)
	if err != nil {
		return nil
	}

	result := make([]map[string]interface{}, 0)
	for _, container := range containers {
		// Get container stats for CPU and memory usage
		stats, err := h.api.docker.GetContainerStats(container.ID)
		var cpuUsage float64
		var memoryUsage uint64

		if err == nil && stats != nil {
			if cpu, ok := stats["cpu_usage"].(float64); ok {
				cpuUsage = cpu
			}
			if mem, ok := stats["memory_usage"].(uint64); ok {
				memoryUsage = mem
			}
		}

		containerData := map[string]interface{}{
			"name":         container.Name,
			"state":        container.State,
			"status":       container.Status,
			"cpu_usage":    cpuUsage,
			"memory_usage": memoryUsage,
			"image":        container.Image,
			"created":      container.Created,
		}

		result = append(result, containerData)
	}

	return result
}

// getVMDataForHA returns VM data in HA integration format
func (h *HTTPServer) getVMDataForHA() []map[string]interface{} {
	vms, err := h.api.vm.ListVMs(false)
	if err != nil {
		return nil
	}

	result := make([]map[string]interface{}, 0)
	for _, vm := range vms {
		// Get VM stats for CPU and memory usage
		stats, err := h.api.vm.GetVMStats(vm.Name)
		var cpuUsage float64
		var memoryUsage uint64

		if err == nil && stats != nil {
			if cpu, ok := stats["cpu_usage"].(float64); ok {
				cpuUsage = cpu
			}
			if mem, ok := stats["memory_usage"].(uint64); ok {
				memoryUsage = mem
			}
		}

		vmData := map[string]interface{}{
			"name":         vm.Name,
			"state":        vm.State,
			"cpu_usage":    cpuUsage,
			"memory_usage": memoryUsage,
			"uuid":         vm.UUID,
		}

		result = append(result, vmData)
	}

	return result
}

// handleGPU handles GET /api/v1/gpu
func (h *HTTPServer) handleGPU(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	gpuInfo, err := h.api.gpu.GetGPUInfo()
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get GPU info: %v", err))
		return
	}

	h.writeJSON(w, http.StatusOK, gpuInfo)
}

// handleDockerContainers handles GET /api/v1/docker/containers
func (h *HTTPServer) handleDockerContainers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	all := r.URL.Query().Get("all") == "true"
	containers, err := h.api.docker.ListContainers(all)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to list containers: %v", err))
		return
	}

	h.writeJSON(w, http.StatusOK, containers)
}

// handleDockerContainer handles Docker container operations
func (h *HTTPServer) handleDockerContainer(w http.ResponseWriter, r *http.Request) {
	// Extract container name/ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/docker/container/")
	parts := strings.Split(path, "/")
	if len(parts) == 0 || parts[0] == "" {
		h.writeError(w, http.StatusBadRequest, "Container name/ID required")
		return
	}

	containerID := parts[0]
	action := ""
	if len(parts) > 1 {
		action = parts[1]
	}

	switch r.Method {
	case http.MethodGet:
		if action == "logs" {
			lines := 100
			if linesParam := r.URL.Query().Get("lines"); linesParam != "" {
				if l, err := strconv.Atoi(linesParam); err == nil {
					lines = l
				}
			}

			logs, err := h.api.docker.GetContainerLogs(containerID, lines, false)
			if err != nil {
				h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get logs: %v", err))
				return
			}

			h.writeJSON(w, http.StatusOK, map[string]interface{}{"logs": logs})
		} else if action == "stats" {
			stats, err := h.api.docker.GetContainerStats(containerID)
			if err != nil {
				h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get stats: %v", err))
				return
			}

			h.writeJSON(w, http.StatusOK, stats)
		} else {
			container, err := h.api.docker.GetContainer(containerID)
			if err != nil {
				h.writeError(w, http.StatusNotFound, fmt.Sprintf("Container not found: %v", err))
				return
			}

			h.writeJSON(w, http.StatusOK, container)
		}

	case http.MethodPost:
		switch action {
		case "start":
			err := h.api.docker.StartContainer(containerID)
			if err != nil {
				h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to start container: %v", err))
				return
			}
			h.writeJSON(w, http.StatusOK, map[string]string{"message": "Container started"})

		case "stop":
			timeout := 10
			if timeoutParam := r.URL.Query().Get("timeout"); timeoutParam != "" {
				if t, err := strconv.Atoi(timeoutParam); err == nil {
					timeout = t
				}
			}

			err := h.api.docker.StopContainer(containerID, timeout)
			if err != nil {
				h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to stop container: %v", err))
				return
			}
			h.writeJSON(w, http.StatusOK, map[string]string{"message": "Container stopped"})

		case "restart":
			timeout := 10
			if timeoutParam := r.URL.Query().Get("timeout"); timeoutParam != "" {
				if t, err := strconv.Atoi(timeoutParam); err == nil {
					timeout = t
				}
			}

			err := h.api.docker.RestartContainer(containerID, timeout)
			if err != nil {
				h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to restart container: %v", err))
				return
			}
			h.writeJSON(w, http.StatusOK, map[string]string{"message": "Container restarted"})

		case "pause":
			err := h.api.docker.PauseContainer(containerID)
			if err != nil {
				h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to pause container: %v", err))
				return
			}
			h.writeJSON(w, http.StatusOK, map[string]string{"message": "Container paused"})

		case "unpause":
			err := h.api.docker.UnpauseContainer(containerID)
			if err != nil {
				h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to unpause container: %v", err))
				return
			}
			h.writeJSON(w, http.StatusOK, map[string]string{"message": "Container unpaused"})

		default:
			h.writeError(w, http.StatusBadRequest, "Invalid action")
		}

	case http.MethodDelete:
		force := r.URL.Query().Get("force") == "true"
		err := h.api.docker.RemoveContainer(containerID, force)
		if err != nil {
			h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to remove container: %v", err))
			return
		}
		h.writeJSON(w, http.StatusOK, map[string]string{"message": "Container removed"})

	default:
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// handleDockerInfo handles GET /api/v1/docker/info
func (h *HTTPServer) handleDockerInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	info, err := h.api.docker.GetDockerInfo()
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get Docker info: %v", err))
		return
	}

	h.writeJSON(w, http.StatusOK, info)
}

// handleVMList handles GET /api/v1/vm/list
func (h *HTTPServer) handleVMList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	all := r.URL.Query().Get("all") == "true"
	vms, err := h.api.vm.ListVMs(all)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to list VMs: %v", err))
		return
	}

	h.writeJSON(w, http.StatusOK, vms)
}

// handleVM handles VM operations
func (h *HTTPServer) handleVM(w http.ResponseWriter, r *http.Request) {
	// Extract VM name from URL path
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/vm/")
	parts := strings.Split(path, "/")
	if len(parts) == 0 || parts[0] == "" {
		h.writeError(w, http.StatusBadRequest, "VM name required")
		return
	}

	vmName := parts[0]
	action := ""
	if len(parts) > 1 {
		action = parts[1]
	}

	switch r.Method {
	case http.MethodGet:
		if action == "stats" {
			stats, err := h.api.vm.GetVMStats(vmName)
			if err != nil {
				h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get VM stats: %v", err))
				return
			}
			h.writeJSON(w, http.StatusOK, stats)
		} else if action == "console" {
			console, err := h.api.vm.GetVMConsole(vmName)
			if err != nil {
				h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get VM console: %v", err))
				return
			}
			h.writeJSON(w, http.StatusOK, map[string]string{"console": console})
		} else {
			vm, err := h.api.vm.GetVM(vmName)
			if err != nil {
				h.writeError(w, http.StatusNotFound, fmt.Sprintf("VM not found: %v", err))
				return
			}
			h.writeJSON(w, http.StatusOK, vm)
		}

	case http.MethodPost:
		switch action {
		case "start":
			err := h.api.vm.StartVM(vmName)
			if err != nil {
				h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to start VM: %v", err))
				return
			}
			h.writeJSON(w, http.StatusOK, map[string]string{"message": "VM started"})

		case "stop":
			force := r.URL.Query().Get("force") == "true"
			err := h.api.vm.StopVM(vmName, force)
			if err != nil {
				h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to stop VM: %v", err))
				return
			}
			h.writeJSON(w, http.StatusOK, map[string]string{"message": "VM stopped"})

		case "restart":
			err := h.api.vm.RestartVM(vmName)
			if err != nil {
				h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to restart VM: %v", err))
				return
			}
			h.writeJSON(w, http.StatusOK, map[string]string{"message": "VM restarted"})

		case "pause":
			err := h.api.vm.PauseVM(vmName)
			if err != nil {
				h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to pause VM: %v", err))
				return
			}
			h.writeJSON(w, http.StatusOK, map[string]string{"message": "VM paused"})

		case "resume":
			err := h.api.vm.ResumeVM(vmName)
			if err != nil {
				h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to resume VM: %v", err))
				return
			}
			h.writeJSON(w, http.StatusOK, map[string]string{"message": "VM resumed"})

		case "autostart":
			autostart := r.URL.Query().Get("enable") == "true"
			err := h.api.vm.SetVMAutostart(vmName, autostart)
			if err != nil {
				h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to set autostart: %v", err))
				return
			}
			h.writeJSON(w, http.StatusOK, map[string]string{"message": "VM autostart updated"})

		default:
			h.writeError(w, http.StatusBadRequest, "Invalid action")
		}

	default:
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// handleDiagnosticsHealth handles GET /api/v1/diagnostics/health
func (h *HTTPServer) handleDiagnosticsHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	health, err := h.api.diagnostics.RunHealthChecks()
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to run health checks: %v", err))
		return
	}

	h.writeJSON(w, http.StatusOK, health)
}

// handleDiagnosticsInfo handles GET /api/v1/diagnostics/info
func (h *HTTPServer) handleDiagnosticsInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	info, err := h.api.diagnostics.GetDiagnosticInfo()
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get diagnostic info: %v", err))
		return
	}

	h.writeJSON(w, http.StatusOK, info)
}

// handleDiagnosticsRepair handles GET/POST /api/v1/diagnostics/repair
func (h *HTTPServer) handleDiagnosticsRepair(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		repairs := h.api.diagnostics.GetAvailableRepairs()
		h.writeJSON(w, http.StatusOK, repairs)

	case http.MethodPost:
		repairName := r.URL.Query().Get("action")
		if repairName == "" {
			h.writeError(w, http.StatusBadRequest, "Repair action required")
			return
		}

		err := h.api.diagnostics.ExecuteRepair(repairName)
		if err != nil {
			h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to execute repair: %v", err))
			return
		}

		h.writeJSON(w, http.StatusOK, map[string]string{"message": "Repair executed successfully"})

	default:
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// corsMiddleware adds CORS headers
func (h *HTTPServer) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// loggingMiddleware logs HTTP requests
func (h *HTTPServer) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// Create a response writer wrapper to capture status code
		wrapper := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		
		next.ServeHTTP(wrapper, r)
		
		duration := time.Since(start)
		logger.LightGreen("HTTP %s %s %d %v", r.Method, r.URL.Path, wrapper.statusCode, duration)
	})
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// writeJSON writes a JSON response
func (h *HTTPServer) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	
	if err := json.NewEncoder(w).Encode(data); err != nil {
		logger.Yellow("Error encoding JSON response: %v", err)
	}
}

// writeError writes an error response
func (h *HTTPServer) writeError(w http.ResponseWriter, status int, message string) {
	errorResponse := dto.Response{
		Error:   message,
		Message: http.StatusText(status),
	}
	h.writeJSON(w, status, errorResponse)
}
