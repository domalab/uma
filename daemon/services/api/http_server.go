package api

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/domalab/uma/daemon/domain"
	"github.com/domalab/uma/daemon/dto"
	"github.com/domalab/uma/daemon/lib"
	"github.com/domalab/uma/daemon/logger"
	"github.com/domalab/uma/daemon/plugins/notifications"
	"github.com/domalab/uma/daemon/plugins/storage"
	"github.com/domalab/uma/daemon/services/auth"
	"github.com/domalab/uma/daemon/services/command"
	"github.com/domalab/uma/daemon/services/config"
	"github.com/go-playground/validator/v10"
)

// Context key types for type safety
type contextKey string

const (
	requestIDKey  contextKey = "request_id"
	apiVersionKey contextKey = "api_version"
)

// Array control request/response structures
type ArrayStartRequest struct {
	MaintenanceMode bool `json:"maintenance_mode"`
	CheckFilesystem bool `json:"check_filesystem"`
}

type ArrayStopRequest struct {
	Force          bool `json:"force"`
	UnmountShares  bool `json:"unmount_shares"`
	StopContainers bool `json:"stop_containers"`
	StopVMs        bool `json:"stop_vms"`
}

type ParityCheckRequest struct {
	Type     string `json:"type"`     // "check" or "correct"
	Priority string `json:"priority"` // "low", "normal", "high"
}

type DiskAddRequest struct {
	Device   string `json:"device"`   // e.g., "/dev/sdc"
	Position string `json:"position"` // e.g., "disk1", "parity2"
}

type DiskRemoveRequest struct {
	Position string `json:"position"` // e.g., "disk1", "parity2"
}

type ArrayOperationResponse struct {
	Success       bool   `json:"success"`
	Message       string `json:"message"`
	OperationID   string `json:"operation_id,omitempty"`
	EstimatedTime int    `json:"estimated_time,omitempty"` // seconds
}

type ParityCheckStatus struct {
	Active        bool    `json:"active"`
	Type          string  `json:"type,omitempty"`           // "check" or "correct"
	Progress      float64 `json:"progress,omitempty"`       // 0-100
	Speed         string  `json:"speed,omitempty"`          // e.g., "45.2 MB/s"
	TimeRemaining string  `json:"time_remaining,omitempty"` // e.g., "2h 15m"
	Errors        int     `json:"errors,omitempty"`
}

// Power management request/response structures
type SystemShutdownRequest struct {
	DelaySeconds int    `json:"delay_seconds"` // Delay before shutdown (0-300 seconds)
	Message      string `json:"message"`       // Message to display to users
	Force        bool   `json:"force"`         // Force shutdown even if users are logged in
}

type SystemRebootRequest struct {
	DelaySeconds int    `json:"delay_seconds"` // Delay before reboot (0-300 seconds)
	Message      string `json:"message"`       // Message to display to users
	Force        bool   `json:"force"`         // Force reboot even if users are logged in
}

type SystemSleepRequest struct {
	Type string `json:"type"` // "suspend", "hibernate", or "hybrid"
}

type SystemWakeRequest struct {
	TargetMAC   string `json:"target_mac"`   // MAC address to wake
	BroadcastIP string `json:"broadcast_ip"` // Broadcast IP (optional, defaults to 255.255.255.255)
	Port        int    `json:"port"`         // Port for WOL packet (optional, defaults to 9)
	RepeatCount int    `json:"repeat_count"` // Number of packets to send (optional, defaults to 3)
}

type PowerOperationResponse struct {
	Success       bool   `json:"success"`
	Message       string `json:"message"`
	OperationID   string `json:"operation_id,omitempty"`
	ScheduledTime string `json:"scheduled_time,omitempty"` // ISO 8601 format
}

// User Script Management data structures
type UserScript struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Path        string `json:"path"`
	Status      string `json:"status"`      // "idle", "running", "completed", "failed"
	LastRun     string `json:"last_run"`    // ISO 8601 format
	LastResult  string `json:"last_result"` // "success", "failed", "unknown"
	PID         int    `json:"pid,omitempty"`
}

type ScriptListResponse struct {
	Scripts []UserScript `json:"scripts"`
}

type ScriptExecuteRequest struct {
	Background bool     `json:"background"`
	Arguments  []string `json:"arguments,omitempty"`
}

type ScriptExecuteResponse struct {
	Success     bool   `json:"success"`
	Message     string `json:"message"`
	ExecutionID string `json:"execution_id,omitempty"`
	PID         int    `json:"pid,omitempty"`
}

type ScriptStatusResponse struct {
	Name      string `json:"name"`
	Status    string `json:"status"`
	PID       int    `json:"pid,omitempty"`
	StartTime string `json:"start_time,omitempty"`
	Duration  string `json:"duration,omitempty"`
	ExitCode  int    `json:"exit_code,omitempty"`
}

type ScriptLogsResponse struct {
	Name string   `json:"name"`
	Logs []string `json:"logs"`
}

// Share Management data structures
type Share struct {
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

type ShareUsage struct {
	Name           string  `json:"name"`
	TotalSize      int64   `json:"total_size"`   // bytes
	UsedSize       int64   `json:"used_size"`    // bytes
	FreeSize       int64   `json:"free_size"`    // bytes
	UsedPercent    float64 `json:"used_percent"` // 0-100
	FileCount      int64   `json:"file_count"`
	DirectoryCount int64   `json:"directory_count"`
	LastAccessed   string  `json:"last_accessed"` // ISO 8601 format
}

type ShareListResponse struct {
	Shares []Share `json:"shares"`
}

type ShareCreateRequest struct {
	Name             string   `json:"name"`
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

type ShareOperationResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	ShareName string `json:"share_name,omitempty"`
}

// CacheEntry represents a cached data entry with expiration
type CacheEntry struct {
	Data      interface{}
	ExpiresAt time.Time
}

// GeneralFormatCache caches expensive operations for the general format endpoint
type GeneralFormatCache struct {
	mu                sync.RWMutex
	systemData        *CacheEntry
	dockerData        *CacheEntry
	vmData            *CacheEntry
	cacheDuration     time.Duration
	lastArrayInfoHash string
}

// HTTPServer handles REST API requests
type HTTPServer struct {
	api             *Api
	server          *http.Server
	port            int
	commandExecutor *command.CommandExecutor
	generalCache    *GeneralFormatCache
	wsManager       *WebSocketManager
	authService     *auth.AuthService
	configService   *config.ViperConfigService
	validator       *validator.Validate
}

// NewHTTPServer creates a new HTTP server instance
func NewHTTPServer(api *Api, port int) *HTTPServer {
	httpServer := &HTTPServer{
		api:             api,
		port:            port,
		commandExecutor: command.NewCommandExecutor(),
		generalCache: &GeneralFormatCache{
			cacheDuration: 30 * time.Second, // Cache for 30 seconds
		},
		authService:   api.authService, // Use auth service from API
		configService: config.NewViperConfigService(),
		validator:     validator.New(),
	}

	// Initialize WebSocket manager
	httpServer.wsManager = NewWebSocketManager(httpServer)

	return httpServer
}

// authMiddleware wraps the auth service middleware for HTTP mux
func (h *HTTPServer) authMiddleware(next http.Handler) http.Handler {
	if h.authService != nil && h.authService.IsEnabled() {
		return h.authService.AuthMiddleware(next)
	}
	return next
}

// Start starts the HTTP server
func (h *HTTPServer) Start() error {
	// Setup HTTP mux with all API routes
	mux := http.NewServeMux()

	// Health and documentation endpoints
	mux.HandleFunc("/api/v1/health", h.handleHealth)
	mux.HandleFunc("/api/v1/docs", h.handleSwaggerUI)
	mux.HandleFunc("/api/v1/openapi.json", h.handleOpenAPISpec)
	mux.HandleFunc("/metrics", h.handleMetrics)

	// System monitoring endpoints
	mux.HandleFunc("/api/v1/system/info", h.handleSystemInfo)
	mux.HandleFunc("/api/v1/system/cpu", h.handleSystemCPU)
	mux.HandleFunc("/api/v1/system/memory", h.handleSystemMemory)
	mux.HandleFunc("/api/v1/system/fans", h.handleSystemFans)
	mux.HandleFunc("/api/v1/system/temperature", h.handleSystemTemperature)
	mux.HandleFunc("/api/v1/system/gpu", h.handleSystemGPU)
	mux.HandleFunc("/api/v1/system/ups", h.handleSystemUPS)
	mux.HandleFunc("/api/v1/system/network", h.handleSystemNetwork)
	mux.HandleFunc("/api/v1/system/resources", h.handleSystemResources)
	mux.HandleFunc("/api/v1/system/filesystems", h.handleSystemFilesystems)

	// System control endpoints
	mux.HandleFunc("/api/v1/system/scripts", h.handleSystemScripts)
	mux.HandleFunc("/api/v1/system/execute", h.handleSystemExecute)
	mux.HandleFunc("/api/v1/system/reboot", h.handleSystemReboot)
	mux.HandleFunc("/api/v1/system/shutdown", h.handleSystemShutdown)
	mux.HandleFunc("/api/v1/system/logs", h.handleSystemLogs)
	mux.HandleFunc("/api/v1/system/logs/all", h.handleSystemLogsAll)
	mux.HandleFunc("/api/v1/system/parity/disk", h.handleParityDisk)
	mux.HandleFunc("/api/v1/system/parity/check", h.handleParityCheck)

	// Storage management endpoints
	mux.HandleFunc("/api/v1/storage/disks", h.handleStorageDisks)
	mux.HandleFunc("/api/v1/storage/array", h.handleStorageArray)
	mux.HandleFunc("/api/v1/storage/cache", h.handleStorageCache)
	mux.HandleFunc("/api/v1/storage/boot", h.handleStorageBoot)
	mux.HandleFunc("/api/v1/storage/zfs", h.handleStorageZFS)
	mux.HandleFunc("/api/v1/storage/general", h.handleStorageGeneral)

	// Array control endpoints with enhanced orchestration
	mux.HandleFunc("/api/v1/storage/array/start", h.handleArrayStart)
	mux.HandleFunc("/api/v1/storage/array/stop", h.handleArrayStop)

	// Docker endpoints
	mux.HandleFunc("/api/v1/docker/containers", h.handleDockerContainers)
	mux.HandleFunc("/api/v1/docker/networks", h.handleDockerNetworks)
	mux.HandleFunc("/api/v1/docker/images", h.handleDockerImages)
	mux.HandleFunc("/api/v1/docker/info", h.handleDockerInfo)
	mux.HandleFunc("/api/v1/docker/containers/bulk/start", h.handleDockerBulkStart)
	mux.HandleFunc("/api/v1/docker/containers/bulk/stop", h.handleDockerBulkStop)
	mux.HandleFunc("/api/v1/docker/containers/bulk/restart", h.handleDockerBulkRestart)

	// Individual Docker container control endpoints
	mux.HandleFunc("/api/v1/docker/containers/", h.handleDockerContainer) // Handles individual container operations

	// VM management endpoints
	mux.HandleFunc("/api/v1/vms", h.handleVMList)
	mux.HandleFunc("/api/v1/vms/", h.handleVM) // Handles all VM operations with path parsing

	// Authentication endpoints
	mux.HandleFunc("/api/v1/auth/login", h.handleAuthLogin)
	mux.HandleFunc("/api/v1/auth/users", h.handleAuthUsers)
	mux.HandleFunc("/api/v1/auth/stats", h.handleAuthStats)

	// WebSocket endpoints
	mux.HandleFunc("/api/v1/ws/system/stats", h.handleSystemStatsWebSocket)
	mux.HandleFunc("/api/v1/ws/docker/events", h.handleDockerEventsWebSocket)
	mux.HandleFunc("/api/v1/ws/storage/status", h.handleStorageStatusWebSocket)

	// Build middleware chain
	handler := h.corsMiddleware(mux)
	handler = h.requestIDMiddleware(handler)
	handler = h.versioningMiddleware(handler)
	handler = h.compressionMiddleware(handler)
	handler = h.metricsMiddleware(handler)
	handler = h.loggingMiddleware(handler)
	// Authentication middleware ready for production (temporarily disabled for testing)
	// handler = h.authMiddleware(handler)

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

	// Start WebSocket broadcasters
	h.startWebSocketBroadcasters()

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

// Cache helper methods for general format optimization

// getCachedData retrieves cached data if valid, otherwise returns nil
func (cache *GeneralFormatCache) getCachedData(entry **CacheEntry) interface{} {
	cache.mu.RLock()
	defer cache.mu.RUnlock()

	if *entry != nil && time.Now().Before((*entry).ExpiresAt) {
		return (*entry).Data
	}
	return nil
}

// setCachedData stores data in cache with expiration
func (cache *GeneralFormatCache) setCachedData(entry **CacheEntry, data interface{}) {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	*entry = &CacheEntry{
		Data:      data,
		ExpiresAt: time.Now().Add(cache.cacheDuration),
	}
}

// Response helper methods for standardized API responses

// parsePaginationParams extracts pagination parameters from request
func (h *HTTPServer) parsePaginationParams(r *http.Request) *dto.PaginationParams {
	params := &dto.PaginationParams{}

	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil {
			params.Page = page
		}
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			params.Limit = limit
		}
	}

	if perPageStr := r.URL.Query().Get("per_page"); perPageStr != "" {
		if perPage, err := strconv.Atoi(perPageStr); err == nil {
			params.PerPage = perPage
		}
	}

	return params
}

// writeStandardResponse writes a standardized API response
func (h *HTTPServer) writeStandardResponse(w http.ResponseWriter, status int, data interface{}, pagination *dto.PaginationInfo) {
	response := dto.StandardResponse{
		Data:       data,
		Pagination: pagination,
		Meta: &dto.ResponseMeta{
			RequestID: h.getRequestID(w),
			Version:   h.api.ctx.Config.Version,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		},
	}

	h.writeJSON(w, status, response)
}

// writePaginatedResponse writes a paginated API response
func (h *HTTPServer) writePaginatedResponse(w http.ResponseWriter, status int, data interface{}, total int, params *dto.PaginationParams) {
	pagination := dto.CalculatePagination(total, params)
	h.writeStandardResponse(w, status, data, pagination)
}

// getRequestID gets the request ID from context or response header
func (h *HTTPServer) getRequestID(w http.ResponseWriter) string {
	// Check if request ID was set in response headers by middleware
	if requestID := w.Header().Get("X-Request-ID"); requestID != "" {
		return requestID
	}

	// Generate a simple request ID as fallback
	return fmt.Sprintf("req_%d", time.Now().UnixNano())
}

// getRequestIDFromContext gets the request ID from request context
func (h *HTTPServer) getRequestIDFromContext(r *http.Request) string {
	if requestID, ok := r.Context().Value(requestIDKey).(string); ok {
		return requestID
	}
	return ""
}

// paginateSlice applies pagination to a slice and returns the paginated subset
func (h *HTTPServer) paginateSlice(items interface{}, params *dto.PaginationParams) (interface{}, int) {
	// This is a generic pagination helper - in practice, you'd implement type-specific versions
	// For now, we'll handle common slice types

	switch v := items.(type) {
	case []interface{}:
		total := len(v)
		start := (params.GetPage() - 1) * params.GetLimit()
		end := start + params.GetLimit()

		if start >= total {
			return []interface{}{}, total
		}
		if end > total {
			end = total
		}

		return v[start:end], total
	default:
		// For other types, return as-is (no pagination)
		return items, 0
	}
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

// handleHealth handles GET /api/v1/health
func (h *HTTPServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Get request ID for tracking and start timing
	requestID := h.getRequestIDFromContext(r)
	startTime := time.Now()

	health := h.getEnhancedHealthStatus()

	// Record health check metrics
	duration := time.Since(startTime)
	if dependencies, ok := health["dependencies"].(map[string]string); ok {
		status, _ := health["status"].(string)
		RecordHealthCheck(status, dependencies, duration, requestID)
	}

	h.writeJSON(w, http.StatusOK, health)
}

// getEnhancedHealthStatus returns comprehensive health status information
func (h *HTTPServer) getEnhancedHealthStatus() map[string]interface{} {
	startTime := time.Now()

	// Check dependencies
	dependencies := h.checkDependencies()

	// Calculate overall status
	overallStatus := "healthy"
	for _, status := range dependencies {
		if status != "healthy" {
			overallStatus = "degraded"
			break
		}
	}

	// Get metrics
	metrics := h.getHealthMetrics()

	health := map[string]interface{}{
		"status":       overallStatus,
		"version":      h.api.ctx.Config.Version,
		"service":      "uma",
		"timestamp":    time.Now().UTC().Format(time.RFC3339),
		"dependencies": dependencies,
		"metrics":      metrics,
		"checks": map[string]interface{}{
			"response_time_ms": time.Since(startTime).Milliseconds(),
		},
	}

	return health
}

// checkDependencies checks the health of various system dependencies
func (h *HTTPServer) checkDependencies() map[string]string {
	dependencies := make(map[string]string)

	// Check Docker
	dependencies["docker"] = h.checkDockerHealth()

	// Check libvirt (VMs)
	dependencies["libvirt"] = h.checkLibvirtHealth()

	// Check storage
	dependencies["storage"] = h.checkStorageHealth()

	// Check notifications system
	dependencies["notifications"] = h.checkNotificationsHealth()

	return dependencies
}

// checkDockerHealth checks if Docker daemon is accessible
func (h *HTTPServer) checkDockerHealth() string {
	if h.api.docker == nil {
		return "unavailable"
	}

	// Try to get Docker info
	_, err := h.api.docker.GetDockerInfo()
	if err != nil {
		return "unhealthy"
	}

	return "healthy"
}

// checkLibvirtHealth checks if libvirt is accessible
func (h *HTTPServer) checkLibvirtHealth() string {
	if h.api.vm == nil {
		return "unavailable"
	}

	// Try to list VMs
	_, err := h.api.vm.ListVMs(false)
	if err != nil {
		return "unhealthy"
	}

	return "healthy"
}

// checkStorageHealth checks if storage monitoring is working
func (h *HTTPServer) checkStorageHealth() string {
	if h.api.storage == nil {
		return "unavailable"
	}

	// Try to get array info
	_, err := h.api.storage.GetArrayInfo()
	if err != nil {
		return "unhealthy"
	}

	return "healthy"
}

// checkNotificationsHealth checks if notifications system is working
func (h *HTTPServer) checkNotificationsHealth() string {
	if h.api.notifications == nil {
		return "unavailable"
	}

	// Try to get notification stats
	_, err := h.api.notifications.GetNotificationStats()
	if err != nil {
		return "unhealthy"
	}

	return "healthy"
}

// getHealthMetrics returns system metrics for health monitoring
func (h *HTTPServer) getHealthMetrics() map[string]interface{} {
	metrics := make(map[string]interface{})

	// Get uptime from system
	if uptimeInfo, err := h.api.system.GetUptimeInfo(); err == nil {
		metrics["uptime_seconds"] = int64(uptimeInfo.Uptime)
		metrics["uptime_human"] = fmt.Sprintf("%.0f seconds", uptimeInfo.Uptime)
	}

	// Get memory usage
	if memInfo, err := h.api.system.GetMemoryInfo(); err == nil {
		metrics["memory_usage_percent"] = memInfo.UsedPercent
		metrics["memory_total_bytes"] = memInfo.Total
		metrics["memory_used_bytes"] = memInfo.Used
		metrics["memory_total_formatted"] = memInfo.TotalFormatted
		metrics["memory_used_formatted"] = memInfo.UsedFormatted

		// Add detailed memory breakdown
		metrics["memory_available_bytes"] = memInfo.Available
		metrics["memory_available_formatted"] = memInfo.AvailableFormatted
		metrics["memory_free_bytes"] = memInfo.Free
		metrics["memory_free_formatted"] = memInfo.FreeFormatted
		metrics["memory_cached_bytes"] = memInfo.Cached
		metrics["memory_cached_formatted"] = memInfo.CachedFormatted
		metrics["memory_buffers_bytes"] = memInfo.Buffers
		metrics["memory_buffers_formatted"] = memInfo.BuffersFormatted

		// Add memory breakdown by category if available
		if memInfo.Breakdown != nil {
			metrics["memory_vm_bytes"] = memInfo.Breakdown.VM
			metrics["memory_vm_formatted"] = memInfo.Breakdown.VMFormatted
			metrics["memory_docker_bytes"] = memInfo.Breakdown.Docker
			metrics["memory_docker_formatted"] = memInfo.Breakdown.DockerFormatted
			metrics["memory_zfs_arc_bytes"] = memInfo.Breakdown.ZFSCache
			metrics["memory_zfs_arc_formatted"] = memInfo.Breakdown.ZFSCacheFormatted
		}
	}

	// Get basic system info
	if cpuInfo, err := h.api.system.GetCPUInfo(); err == nil {
		metrics["cpu_usage_percent"] = cpuInfo.Usage
		metrics["cpu_temperature"] = cpuInfo.Temperature
		metrics["cpu_cores"] = cpuInfo.Cores
		metrics["cpu_model"] = cpuInfo.Model
	}

	// Get CPU load averages
	if loadInfo, err := h.api.system.GetLoadInfo(); err == nil {
		metrics["cpu_load_1m"] = loadInfo.Load1
		metrics["cpu_load_5m"] = loadInfo.Load5
		metrics["cpu_load_15m"] = loadInfo.Load15
	}

	// Add API call metrics (placeholder - would need actual implementation)
	metrics["api_calls_total"] = 0 // This would be tracked by middleware

	// Add global timestamp
	metrics["last_updated"] = time.Now().UTC().Format(time.RFC3339)

	return metrics
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

	cpuData := h.getCPUData()
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

	memoryData := h.getMemoryData()
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

	tempData := h.getTemperatureData()
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

	networkData := h.getNetworkData()
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

	upsData := h.getUPSData()
	h.writeJSON(w, http.StatusOK, upsData)
}

// handleSystemGPU handles GET /api/v1/system/gpu
func (h *HTTPServer) handleSystemGPU(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	gpuData := h.getIntelGPUData()
	if gpuData == nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to get GPU information")
		return
	}

	h.writeJSON(w, http.StatusOK, gpuData)
}

// handleSystemFans handles GET /api/v1/system/fans
func (h *HTTPServer) handleSystemFans(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	timestamp := time.Now().UTC().Format(time.RFC3339)

	// Get enhanced temperature data which includes fan information
	enhancedData, err := h.api.system.GetEnhancedTemperatureData()
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get fan information: %v", err))
		return
	}

	response := map[string]interface{}{
		"fans":         enhancedData.Fans,
		"last_updated": timestamp,
	}

	h.writeJSON(w, http.StatusOK, response)
}

// handleSystemFilesystems handles GET /api/v1/system/filesystems
func (h *HTTPServer) handleSystemFilesystems(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	fsData := h.getFilesystemData()
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

// handleStorageZFS handles GET /api/v1/storage/zfs
func (h *HTTPServer) handleStorageZFS(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	zfsInfo, err := h.api.storage.GetZFSInfo()
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get ZFS info: %v", err))
		return
	}

	// Return appropriate response based on ZFS availability
	if !zfsInfo.Available {
		h.writeJSON(w, http.StatusOK, map[string]interface{}{
			"available":    false,
			"message":      "ZFS is not available on this system",
			"last_updated": zfsInfo.LastUpdated,
		})
		return
	}

	h.writeJSON(w, http.StatusOK, zfsInfo)
}

// handleStorageGeneral handles GET /api/v1/storage/general
// Returns storage data in general format with optimized caching and parallel processing
func (h *HTTPServer) handleStorageGeneral(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	startTime := time.Now()
	logger.Blue("Starting optimized general format data collection...")

	// Get array information (this is the core data that changes less frequently)
	arrayInfo, err := h.api.storage.GetArrayInfo()
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get array info: %v", err))
		return
	}

	// Convert to general format with optimized parallel data collection
	generalFormat := h.convertToGeneralFormatOptimized(arrayInfo)

	duration := time.Since(startTime)
	logger.Blue("General format data collection completed in %v", duration)

	h.writeJSON(w, http.StatusOK, generalFormat)
}

// handleStorageDisks handles GET /api/v1/storage/disks
func (h *HTTPServer) handleStorageDisks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Check if client wants paginated response
	params := h.parsePaginationParams(r)
	usePagination := r.URL.Query().Get("page") != "" || r.URL.Query().Get("limit") != ""

	disksInfo, err := h.api.storage.GetConsolidatedDisksInfo()
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get consolidated disks info: %v", err))
		return
	}

	if usePagination {
		// For pagination, we'll flatten all disks into a single array
		allDisks := make([]interface{}, 0)
		for _, disk := range disksInfo.ArrayDisks {
			allDisks = append(allDisks, disk)
		}
		for _, disk := range disksInfo.ParityDisks {
			allDisks = append(allDisks, disk)
		}
		for _, disk := range disksInfo.CacheDisks {
			allDisks = append(allDisks, disk)
		}
		if disksInfo.BootDisk != nil {
			allDisks = append(allDisks, *disksInfo.BootDisk)
		}

		paginatedDisks, total := h.paginateSlice(allDisks, params)
		h.writePaginatedResponse(w, http.StatusOK, paginatedDisks, total, params)
	} else {
		// Return the original structured format for backward compatibility
		h.writeJSON(w, http.StatusOK, disksInfo)
	}
}

// convertToGeneralFormatOptimized converts array info to general format with caching and parallel processing
func (h *HTTPServer) convertToGeneralFormatOptimized(arrayInfo *storage.ArrayInfo) map[string]interface{} {
	startTime := time.Now()

	// Use channels for parallel data collection
	type dataResult struct {
		name string
		data interface{}
		err  error
	}

	resultChan := make(chan dataResult, 8) // Buffer for 8 concurrent operations
	var wg sync.WaitGroup

	// Parallel data collection functions
	collectSystemData := func() {
		defer wg.Done()

		// Check cache first
		if cached := h.generalCache.getCachedData(&h.generalCache.systemData); cached != nil {
			resultChan <- dataResult{name: "system", data: cached}
			return
		}

		// Collect system data in parallel
		var systemWg sync.WaitGroup
		systemResults := make(map[string]interface{})
		systemMutex := sync.Mutex{}

		// CPU data
		systemWg.Add(1)
		go func() {
			defer systemWg.Done()
			if data := h.getCPUData(); data != nil {
				systemMutex.Lock()
				systemResults["cpu"] = data
				systemMutex.Unlock()
			}
		}()

		// Memory data
		systemWg.Add(1)
		go func() {
			defer systemWg.Done()
			if data := h.getMemoryData(); data != nil {
				systemMutex.Lock()
				systemResults["memory"] = data
				systemMutex.Unlock()
			}
		}()

		// Temperature data
		systemWg.Add(1)
		go func() {
			defer systemWg.Done()
			if data := h.getTemperatureData(); data != nil {
				systemMutex.Lock()
				systemResults["temperature"] = data
				systemMutex.Unlock()
			}
		}()

		// Network data
		systemWg.Add(1)
		go func() {
			defer systemWg.Done()
			if data := h.getNetworkData(); data != nil {
				systemMutex.Lock()
				systemResults["network"] = data
				systemMutex.Unlock()
			}
		}()

		// UPS data
		systemWg.Add(1)
		go func() {
			defer systemWg.Done()
			if data := h.getUPSData(); data != nil {
				systemMutex.Lock()
				systemResults["ups"] = data
				systemMutex.Unlock()
			}
		}()

		// Intel GPU data
		systemWg.Add(1)
		go func() {
			defer systemWg.Done()
			if data := h.getIntelGPUData(); data != nil {
				systemMutex.Lock()
				systemResults["intel_gpu"] = data
				systemMutex.Unlock()
			}
		}()

		// Filesystem data
		systemWg.Add(1)
		go func() {
			defer systemWg.Done()
			if data := h.getFilesystemData(); data != nil {
				systemMutex.Lock()
				systemResults["filesystem"] = data
				systemMutex.Unlock()
			}
		}()

		systemWg.Wait()

		// Cache the results
		h.generalCache.setCachedData(&h.generalCache.systemData, systemResults)
		resultChan <- dataResult{name: "system", data: systemResults}
	}

	collectDockerData := func() {
		defer wg.Done()

		// Check cache first
		if cached := h.generalCache.getCachedData(&h.generalCache.dockerData); cached != nil {
			resultChan <- dataResult{name: "docker", data: cached}
			return
		}

		if data := h.getDockerDataOptimized(); data != nil {
			h.generalCache.setCachedData(&h.generalCache.dockerData, data)
			resultChan <- dataResult{name: "docker", data: data}
		} else {
			resultChan <- dataResult{name: "docker", data: nil}
		}
	}

	collectVMData := func() {
		defer wg.Done()

		// Check cache first
		if cached := h.generalCache.getCachedData(&h.generalCache.vmData); cached != nil {
			resultChan <- dataResult{name: "vm", data: cached}
			return
		}

		if data := h.getVMDataOptimized(); data != nil {
			h.generalCache.setCachedData(&h.generalCache.vmData, data)
			resultChan <- dataResult{name: "vm", data: data}
		} else {
			resultChan <- dataResult{name: "vm", data: nil}
		}
	}

	// Start parallel data collection
	wg.Add(3)
	go collectSystemData()
	go collectDockerData()
	go collectVMData()

	// Close channel when all goroutines complete
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	results := make(map[string]interface{})
	for result := range resultChan {
		if result.err != nil {
			logger.Red("Error collecting %s data: %v", result.name, result.err)
			continue
		}
		if result.data != nil {
			results[result.name] = result.data
		}
	}

	// Build the final system format
	systemFormat := map[string]interface{}{
		"array_usage": map[string]interface{}{
			"total":        arrayInfo.TotalSize,
			"used":         arrayInfo.UsedSize,
			"free":         arrayInfo.FreeSize,
			"used_percent": arrayInfo.UsedPercent,
		},
		"array_state": map[string]interface{}{
			"state":       arrayInfo.State,
			"num_devices": arrayInfo.NumDevices,
			"num_disks":   arrayInfo.NumDisks,
			"num_parity":  arrayInfo.NumParity,
		},
		"disks": h.convertDisksOptimized(arrayInfo.Disks),
	}

	// Add collected data to system format
	if systemData, ok := results["system"].(map[string]interface{}); ok {
		if cpu, exists := systemData["cpu"]; exists {
			systemFormat["cpu"] = cpu
		}
		if memory, exists := systemData["memory"]; exists {
			systemFormat["memory"] = memory
		}
		if temp, exists := systemData["temperature"]; exists {
			systemFormat["temperature"] = temp
		}
		if network, exists := systemData["network"]; exists {
			systemFormat["network"] = network
		}
		if ups, exists := systemData["ups"]; exists {
			systemFormat["ups"] = ups
		}
		if gpu, exists := systemData["intel_gpu"]; exists {
			systemFormat["intel_gpu"] = gpu
		}
		if fs, exists := systemData["filesystem"]; exists {
			systemFormat["filesystem"] = fs
		}
	}

	if dockerData, ok := results["docker"]; ok {
		systemFormat["docker"] = dockerData
	}

	if vmData, ok := results["vm"]; ok {
		systemFormat["vms"] = vmData
	}

	duration := time.Since(startTime)
	logger.Blue("System format conversion completed in %v", duration)

	return systemFormat
}

// getDockerDataOptimized collects Docker data with parallel container processing
func (h *HTTPServer) getDockerDataOptimized() interface{} {
	dockerManager := h.api.docker
	if dockerManager == nil {
		return nil
	}

	containers, err := dockerManager.ListContainers(false) // false = don't include all containers
	if err != nil {
		logger.Red("Failed to get Docker containers: %v", err)
		return nil
	}

	if len(containers) == 0 {
		return map[string]interface{}{
			"containers": []interface{}{},
			"total":      0,
		}
	}

	// Process containers in parallel
	containerChan := make(chan interface{}, len(containers))
	var wg sync.WaitGroup

	for _, container := range containers {
		wg.Add(1)
		go func(c interface{}) {
			defer wg.Done()

			// Get container stats in parallel
			if containerMap, ok := c.(map[string]interface{}); ok {
				if id, exists := containerMap["id"].(string); exists {
					if stats, err := dockerManager.GetContainerStats(id); err == nil {
						containerMap["stats"] = stats
					}
				}
			}
			containerChan <- c
		}(container)
	}

	// Close channel when all goroutines complete
	go func() {
		wg.Wait()
		close(containerChan)
	}()

	// Collect results
	var processedContainers []interface{}
	for container := range containerChan {
		processedContainers = append(processedContainers, container)
	}

	return map[string]interface{}{
		"containers": processedContainers,
		"total":      len(processedContainers),
	}
}

// getVMDataOptimized collects VM data with parallel VM processing
func (h *HTTPServer) getVMDataOptimized() interface{} {
	vmManager := h.api.vm
	if vmManager == nil {
		return nil
	}

	vms, err := vmManager.ListVMs(false) // false = don't include inactive VMs
	if err != nil {
		logger.Red("Failed to get VMs: %v", err)
		return nil
	}

	if len(vms) == 0 {
		return map[string]interface{}{
			"vms":   []interface{}{},
			"total": 0,
		}
	}

	// Process VMs in parallel
	vmChan := make(chan interface{}, len(vms))
	var wg sync.WaitGroup

	for _, vm := range vms {
		wg.Add(1)
		go func(v interface{}) {
			defer wg.Done()

			// Get VM stats in parallel
			if vmMap, ok := v.(map[string]interface{}); ok {
				if name, exists := vmMap["name"].(string); exists {
					if stats, err := vmManager.GetVMStats(name); err == nil {
						vmMap["stats"] = stats
					}
				}
			}
			vmChan <- v
		}(vm)
	}

	// Close channel when all goroutines complete
	go func() {
		wg.Wait()
		close(vmChan)
	}()

	// Collect results
	var processedVMs []interface{}
	for vm := range vmChan {
		processedVMs = append(processedVMs, vm)
	}

	return map[string]interface{}{
		"vms":   processedVMs,
		"total": len(processedVMs),
	}
}

// convertDisksOptimized processes disk data with parallel disk information collection
func (h *HTTPServer) convertDisksOptimized(disks []storage.DiskInfo) []map[string]interface{} {
	if len(disks) == 0 {
		return []map[string]interface{}{}
	}

	// Process disks in parallel
	diskChan := make(chan map[string]interface{}, len(disks))
	var wg sync.WaitGroup

	for _, disk := range disks {
		wg.Add(1)
		go func(d storage.DiskInfo) {
			defer wg.Done()

			diskInfo := map[string]interface{}{
				"name":         d.Name,
				"device":       d.Device,
				"size":         d.Size,
				"used":         d.Used,
				"available":    d.Available,
				"used_percent": d.UsedPercent,
				"mount_point":  d.MountPoint,
				"file_system":  d.FileSystem,
				"status":       d.Status,
				"health":       d.Health,
				"power_state":  d.PowerState,
				"disk_type":    d.DiskType,
				"interface":    d.Interface,
				"model":        d.Model,
			}

			// Add temperature if available
			if d.Temperature > 0 {
				diskInfo["temperature"] = d.Temperature
			}

			// Add serial number if available
			if d.SerialNumber != "" {
				diskInfo["serial_number"] = d.SerialNumber
			}

			// Add spin down delay if available
			if d.SpinDownDelay > 0 {
				diskInfo["spin_down_delay"] = d.SpinDownDelay
			}
			diskChan <- diskInfo
		}(disk)
	}

	// Close channel when all goroutines complete
	go func() {
		wg.Wait()
		close(diskChan)
	}()

	// Collect results
	var processedDisks []map[string]interface{}
	for diskInfo := range diskChan {
		processedDisks = append(processedDisks, diskInfo)
	}

	return processedDisks
}

// handleDockerContainer handles individual Docker container operations
func (h *HTTPServer) handleDockerContainer(w http.ResponseWriter, r *http.Request) {
	// Extract container ID and operation from URL path
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/docker/containers/")
	parts := strings.Split(path, "/")

	if len(parts) < 2 {
		h.writeError(w, http.StatusBadRequest, "Container ID and operation are required")
		return
	}

	containerID := parts[0]
	operation := parts[1]

	if containerID == "" {
		h.writeError(w, http.StatusBadRequest, "Container ID is required")
		return
	}

	// Only allow POST methods for container operations
	if r.Method != http.MethodPost {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Handle different container operations
	switch operation {
	case "start":
		h.handleDockerContainerStart(w, r, containerID)
	case "stop":
		h.handleDockerContainerStop(w, r, containerID)
	case "restart":
		h.handleDockerContainerRestart(w, r, containerID)
	case "pause":
		h.handleDockerContainerPause(w, r, containerID)
	case "resume", "unpause":
		h.handleDockerContainerResume(w, r, containerID)
	default:
		h.writeError(w, http.StatusBadRequest, fmt.Sprintf("Unknown operation: %s", operation))
	}
}

// handleDockerContainerStart handles starting an individual Docker container
func (h *HTTPServer) handleDockerContainerStart(w http.ResponseWriter, r *http.Request, containerID string) {
	err := h.api.docker.StartContainer(containerID)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to start container %s: %v", containerID, err))
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"message":      "Container started successfully",
		"container_id": containerID,
		"operation":    "start",
		"timestamp":    time.Now().UTC().Format(time.RFC3339),
	})
}

// handleDockerContainerStop handles stopping an individual Docker container
func (h *HTTPServer) handleDockerContainerStop(w http.ResponseWriter, r *http.Request, containerID string) {
	// Default timeout of 10 seconds, can be overridden with query parameter
	timeout := 10
	if timeoutParam := r.URL.Query().Get("timeout"); timeoutParam != "" {
		if t, err := strconv.Atoi(timeoutParam); err == nil {
			timeout = t
		}
	}

	err := h.api.docker.StopContainer(containerID, timeout)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to stop container %s: %v", containerID, err))
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"message":      "Container stopped successfully",
		"container_id": containerID,
		"operation":    "stop",
		"timeout":      timeout,
		"timestamp":    time.Now().UTC().Format(time.RFC3339),
	})
}

// handleDockerContainerRestart handles restarting an individual Docker container
func (h *HTTPServer) handleDockerContainerRestart(w http.ResponseWriter, r *http.Request, containerID string) {
	// Default timeout of 10 seconds, can be overridden with query parameter
	timeout := 10
	if timeoutParam := r.URL.Query().Get("timeout"); timeoutParam != "" {
		if t, err := strconv.Atoi(timeoutParam); err == nil {
			timeout = t
		}
	}

	err := h.api.docker.RestartContainer(containerID, timeout)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to restart container %s: %v", containerID, err))
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"message":      "Container restarted successfully",
		"container_id": containerID,
		"operation":    "restart",
		"timeout":      timeout,
		"timestamp":    time.Now().UTC().Format(time.RFC3339),
	})
}

// handleSystemScripts handles GET/POST /api/v1/system/scripts
func (h *HTTPServer) handleSystemScripts(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// List all user scripts
		scripts, err := h.getUserScripts()
		if err != nil {
			h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get user scripts: %v", err))
			return
		}

		response := ScriptListResponse{Scripts: scripts}
		h.writeJSON(w, http.StatusOK, response)

	case http.MethodPost:
		// Execute a script
		scriptName := r.URL.Query().Get("name")
		if scriptName == "" {
			h.writeError(w, http.StatusBadRequest, "Script name is required")
			return
		}

		var req ScriptExecuteRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.writeError(w, http.StatusBadRequest, "Invalid JSON request")
			return
		}

		response, err := h.executeUserScript(scriptName, req)
		if err != nil {
			h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to execute script: %v", err))
			return
		}

		h.writeJSON(w, http.StatusOK, response)

	default:
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// handleSystemReboot handles POST /api/v1/system/reboot
func (h *HTTPServer) handleSystemReboot(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req SystemRebootRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Allow empty body with default values
		req = SystemRebootRequest{
			DelaySeconds: 0,
			Message:      "System reboot initiated via UMA API",
			Force:        false,
		}
	}

	// Validate delay
	if req.DelaySeconds < 0 || req.DelaySeconds > 300 {
		h.writeError(w, http.StatusBadRequest, "Delay must be between 0 and 300 seconds")
		return
	}

	err := h.executeSystemReboot(req.DelaySeconds, req.Message, req.Force)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to initiate reboot: %v", err))
		return
	}

	response := PowerOperationResponse{
		Success:       true,
		Message:       "System reboot initiated",
		OperationID:   fmt.Sprintf("reboot_%d", time.Now().Unix()),
		ScheduledTime: time.Now().Add(time.Duration(req.DelaySeconds) * time.Second).Format(time.RFC3339),
	}

	h.writeJSON(w, http.StatusOK, response)
}

// CommandExecuteRequest represents a command execution request
type CommandExecuteRequest struct {
	Command          string `json:"command"`
	Timeout          int    `json:"timeout,omitempty"`           // Timeout in seconds, default 30
	WorkingDirectory string `json:"working_directory,omitempty"` // Optional working directory
}

// CommandExecuteResponse represents a command execution response
type CommandExecuteResponse struct {
	ExitCode        int    `json:"exit_code"`
	Stdout          string `json:"stdout"`
	Stderr          string `json:"stderr"`
	ExecutionTimeMs int64  `json:"execution_time_ms"`
	Command         string `json:"command"`
	WorkingDir      string `json:"working_directory,omitempty"`
}

// handleSystemExecute handles POST /api/v1/system/execute
func (h *HTTPServer) handleSystemExecute(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Parse request body
	var request CommandExecuteRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate command
	if strings.TrimSpace(request.Command) == "" {
		h.writeError(w, http.StatusBadRequest, "Command is required")
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
		h.writeError(w, http.StatusForbidden, "Command not allowed")
		return
	}

	// Execute command
	response := h.executeCommand(request)
	h.writeJSON(w, http.StatusOK, response)
}

// handleSystemShutdown handles POST /api/v1/system/shutdown
func (h *HTTPServer) handleSystemShutdown(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req SystemShutdownRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Allow empty body with default values
		req = SystemShutdownRequest{
			DelaySeconds: 0,
			Message:      "System shutdown initiated via UMA API",
			Force:        false,
		}
	}

	// Validate delay
	if req.DelaySeconds < 0 || req.DelaySeconds > 300 {
		h.writeError(w, http.StatusBadRequest, "Delay must be between 0 and 300 seconds")
		return
	}

	err := h.executeSystemShutdown(req.DelaySeconds, req.Message, req.Force)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to initiate shutdown: %v", err))
		return
	}

	response := PowerOperationResponse{
		Success:       true,
		Message:       "System shutdown initiated",
		OperationID:   fmt.Sprintf("shutdown_%d", time.Now().Unix()),
		ScheduledTime: time.Now().Add(time.Duration(req.DelaySeconds) * time.Second).Format(time.RFC3339),
	}

	h.writeJSON(w, http.StatusOK, response)
}

// handleSystemLogs handles GET /api/v1/system/logs
func (h *HTTPServer) handleSystemLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Parse query parameters
	logType := r.URL.Query().Get("type")
	if logType == "" {
		logType = "system"
	}

	// Support custom log file path
	customPath := r.URL.Query().Get("path")

	// Support grep filtering
	grepFilter := r.URL.Query().Get("grep")

	// Support tail parameter (alias for lines)
	lines := 100
	if linesParam := r.URL.Query().Get("lines"); linesParam != "" {
		if l, err := strconv.Atoi(linesParam); err == nil && l > 0 && l <= 10000 {
			lines = l
		}
	}
	if tailParam := r.URL.Query().Get("tail"); tailParam != "" {
		if l, err := strconv.Atoi(tailParam); err == nil && l > 0 && l <= 10000 {
			lines = l
		}
	}

	follow := r.URL.Query().Get("follow") == "true"
	since := r.URL.Query().Get("since") // ISO 8601 format

	var logs []string
	var err error

	if customPath != "" {
		// Use custom log file path
		logs, err = h.getCustomLogFile(customPath, lines, grepFilter, since)
	} else {
		// Use standard log types
		logs, err = h.getSystemLogs(logType, lines, follow, since)
	}

	if err != nil {
		h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get system logs: %v", err))
		return
	}

	// Apply grep filter if specified and not already applied
	if grepFilter != "" && customPath == "" {
		logs = h.filterLogLines(logs, grepFilter)
	}

	response := map[string]interface{}{
		"type":  logType,
		"lines": lines,
		"logs":  logs,
	}

	if customPath != "" {
		response["path"] = customPath
	}
	if grepFilter != "" {
		response["grep_filter"] = grepFilter
	}

	h.writeJSON(w, http.StatusOK, response)
}

// LogFileInfo represents metadata about a log file
type LogFileInfo struct {
	Path         string `json:"path"`
	Name         string `json:"name"`
	Size         int64  `json:"size"`
	ModifiedTime string `json:"modified_time"`
	Readable     bool   `json:"readable"`
}

// handleSystemLogsAll handles GET /api/v1/system/logs/all
func (h *HTTPServer) handleSystemLogsAll(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
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
	if maxFilesParam := r.URL.Query().Get("max_files"); maxFilesParam != "" {
		if m, err := strconv.Atoi(maxFilesParam); err == nil && m > 0 && m <= 1000 {
			maxFiles = m
		}
	}

	// Security: Restrict to /var/log and subdirectories only
	if !strings.HasPrefix(directory, "/var/log") {
		h.writeError(w, http.StatusForbidden, "Access restricted to /var/log directory")
		return
	}

	logFiles, err := h.scanLogFiles(directory, recursive, filePattern, maxFiles)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to scan log files: %v", err))
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"directory":    directory,
		"recursive":    recursive,
		"file_pattern": filePattern,
		"max_files":    maxFiles,
		"total_found":  len(logFiles),
		"files":        logFiles,
	})
}

// getCustomLogFile reads a custom log file with filtering
func (h *HTTPServer) getCustomLogFile(filePath string, lines int, grepFilter, since string) ([]string, error) {
	// Security: Restrict to /var/log directory
	if !strings.HasPrefix(filePath, "/var/log") {
		return nil, fmt.Errorf("access restricted to /var/log directory")
	}

	// Check if file exists and is readable
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return []string{}, nil
	}

	// Build tail command
	args := []string{"-n", strconv.Itoa(lines), filePath}
	cmd := exec.Command("tail", args...)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to read log file: %v", err)
	}

	logLines := strings.Split(strings.TrimSpace(string(output)), "\n")

	// Apply grep filter if specified
	if grepFilter != "" {
		logLines = h.filterLogLines(logLines, grepFilter)
	}

	// Apply since filter if specified
	if since != "" {
		logLines = h.filterLogLinesBySince(logLines, since)
	}

	return logLines, nil
}

// filterLogLines filters log lines using grep-like pattern matching
func (h *HTTPServer) filterLogLines(lines []string, pattern string) []string {
	filtered := make([]string, 0)
	for _, line := range lines {
		if strings.Contains(strings.ToLower(line), strings.ToLower(pattern)) {
			filtered = append(filtered, line)
		}
	}
	return filtered
}

// filterLogLinesBySince filters log lines by timestamp
func (h *HTTPServer) filterLogLinesBySince(lines []string, since string) []string {
	sinceTime, err := time.Parse(time.RFC3339, since)
	if err != nil {
		return lines // Return original if parsing fails
	}

	filtered := make([]string, 0)
	for _, line := range lines {
		// Simple time parsing - this could be improved based on log format
		if strings.Contains(line, sinceTime.Format("2006-01-02")) {
			filtered = append(filtered, line)
		}
	}
	return filtered
}

// scanLogFiles scans for log files in the specified directory
func (h *HTTPServer) scanLogFiles(directory string, recursive bool, filePattern string, maxFiles int) ([]LogFileInfo, error) {
	logFiles := make([]LogFileInfo, 0)
	count := 0

	walkFunc := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files with errors
		}

		if count >= maxFiles {
			return filepath.SkipDir
		}

		// Skip directories unless we're doing recursive scan
		if info.IsDir() {
			if !recursive && path != directory {
				return filepath.SkipDir
			}
			return nil
		}

		// Apply file pattern filter if specified
		if filePattern != "" {
			matched, err := filepath.Match(filePattern, info.Name())
			if err != nil || !matched {
				return nil
			}
		}

		// Check if file is likely a log file
		if h.isLogFile(path, info.Name()) {
			logFile := LogFileInfo{
				Path:         path,
				Name:         info.Name(),
				Size:         info.Size(),
				ModifiedTime: info.ModTime().Format(time.RFC3339),
				Readable:     h.isFileReadable(path),
			}
			logFiles = append(logFiles, logFile)
			count++
		}

		return nil
	}

	err := filepath.Walk(directory, walkFunc)
	if err != nil {
		return nil, err
	}

	return logFiles, nil
}

// isLogFile determines if a file is likely a log file
func (h *HTTPServer) isLogFile(path, name string) bool {
	// Common log file extensions and patterns
	logExtensions := []string{".log", ".txt", ".out", ".err"}
	logPatterns := []string{"log", "syslog", "messages", "kern", "auth", "mail", "cron", "daemon"}

	lowerName := strings.ToLower(name)

	// Check extensions
	for _, ext := range logExtensions {
		if strings.HasSuffix(lowerName, ext) {
			return true
		}
	}

	// Check patterns
	for _, pattern := range logPatterns {
		if strings.Contains(lowerName, pattern) {
			return true
		}
	}

	// Check if it's a numbered log file (e.g., syslog.1, messages.2)
	if strings.Contains(lowerName, ".") {
		parts := strings.Split(lowerName, ".")
		if len(parts) >= 2 {
			// Check if last part is a number
			if _, err := strconv.Atoi(parts[len(parts)-1]); err == nil {
				return true
			}
		}
	}

	return false
}

// isFileReadable checks if a file is readable
func (h *HTTPServer) isFileReadable(path string) bool {
	file, err := os.Open(path)
	if err != nil {
		return false
	}
	file.Close()
	return true
}

// handleParityDisk handles GET /api/v1/system/parity/disk
func (h *HTTPServer) handleParityDisk(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	parityDisk, err := h.api.system.GetParityDiskInfo()
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get parity disk information: %v", err))
		return
	}

	h.writeJSON(w, http.StatusOK, parityDisk)
}

// handleParityCheck handles GET /api/v1/system/parity/check
func (h *HTTPServer) handleParityCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	parityCheck, err := h.api.system.GetParityCheckInfo()
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get parity check information: %v", err))
		return
	}

	h.writeJSON(w, http.StatusOK, parityCheck)
}

// getUserScripts returns a list of available user scripts
func (h *HTTPServer) getUserScripts() ([]UserScript, error) {
	scriptsDir := "/boot/config/plugins/user.scripts/scripts"
	scripts := []UserScript{}

	// Check if scripts directory exists
	if _, err := os.Stat(scriptsDir); os.IsNotExist(err) {
		return scripts, nil // Return empty list if directory doesn't exist
	}

	entries, err := os.ReadDir(scriptsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read scripts directory: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			scriptName := entry.Name()
			scriptPath := filepath.Join(scriptsDir, scriptName, "script")

			// Check if script file exists
			if _, err := os.Stat(scriptPath); err == nil {
				script := UserScript{
					Name:        scriptName,
					Description: fmt.Sprintf("User script: %s", scriptName),
					Path:        scriptPath,
					Status:      "idle",
					LastRun:     "",
					LastResult:  "unknown",
				}

				// Try to get script status from running processes
				if pid := h.getScriptPID(scriptName); pid > 0 {
					script.Status = "running"
					script.PID = pid
				}

				scripts = append(scripts, script)
			}
		}
	}

	return scripts, nil
}

// executeUserScript executes a user script
func (h *HTTPServer) executeUserScript(scriptName string, req ScriptExecuteRequest) (*ScriptExecuteResponse, error) {
	scriptsDir := "/boot/config/plugins/user.scripts/scripts"
	scriptPath := filepath.Join(scriptsDir, scriptName, "script")

	// Verify script exists
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("script not found: %s", scriptName)
	}

	// Prepare command
	args := []string{scriptPath}
	if len(req.Arguments) > 0 {
		args = append(args, req.Arguments...)
	}

	cmd := exec.Command("/bin/bash", args...)

	if req.Background {
		// Start in background
		err := cmd.Start()
		if err != nil {
			return nil, fmt.Errorf("failed to start script: %v", err)
		}

		return &ScriptExecuteResponse{
			Success:     true,
			Message:     fmt.Sprintf("Script %s started in background", scriptName),
			ExecutionID: fmt.Sprintf("script_%s_%d", scriptName, time.Now().Unix()),
			PID:         cmd.Process.Pid,
		}, nil
	} else {
		// Run synchronously
		output, err := cmd.CombinedOutput()
		if err != nil {
			return &ScriptExecuteResponse{
				Success: false,
				Message: fmt.Sprintf("Script failed: %v\nOutput: %s", err, string(output)),
			}, nil
		}

		return &ScriptExecuteResponse{
			Success: true,
			Message: fmt.Sprintf("Script completed successfully\nOutput: %s", string(output)),
		}, nil
	}
}

// getScriptPID checks if a script is currently running and returns its PID
func (h *HTTPServer) getScriptPID(scriptName string) int {
	cmd := exec.Command("pgrep", "-f", fmt.Sprintf("scripts/%s/script", scriptName))
	output, err := cmd.Output()
	if err != nil {
		return 0
	}

	pidStr := strings.TrimSpace(string(output))
	if pid, err := strconv.Atoi(pidStr); err == nil {
		return pid
	}

	return 0
}

// getSystemLogs retrieves system logs based on type and parameters
func (h *HTTPServer) getSystemLogs(logType string, lines int, follow bool, since string) ([]string, error) {
	var logFile string

	// Determine log file based on type
	switch logType {
	case "system", "syslog":
		logFile = "/var/log/syslog"
	case "kernel", "dmesg":
		// Use dmesg for kernel logs
		cmd := exec.Command("dmesg")
		if since != "" {
			cmd.Args = append(cmd.Args, "--since", since)
		}
		output, err := cmd.Output()
		if err != nil {
			return nil, fmt.Errorf("failed to get kernel logs: %v", err)
		}

		logLines := strings.Split(strings.TrimSpace(string(output)), "\n")
		if len(logLines) > lines {
			logLines = logLines[len(logLines)-lines:]
		}
		return logLines, nil

	case "docker":
		logFile = "/var/log/docker.log"
	case "nginx":
		logFile = "/var/log/nginx/error.log"
	case "unraid":
		logFile = "/var/log/unraid.log"
	default:
		return nil, fmt.Errorf("unsupported log type: %s", logType)
	}

	// Check if log file exists
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		return []string{}, nil // Return empty if log file doesn't exist
	}

	// Use tail command to get logs
	args := []string{"-n", strconv.Itoa(lines)}
	if follow {
		args = append(args, "-f")
	}
	args = append(args, logFile)

	cmd := exec.Command("tail", args...)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to read log file: %v", err)
	}

	logLines := strings.Split(strings.TrimSpace(string(output)), "\n")

	// Filter by since time if provided
	if since != "" {
		filteredLines := []string{}
		sinceTime, err := time.Parse(time.RFC3339, since)
		if err == nil {
			for _, line := range logLines {
				// Simple time parsing - this could be improved based on log format
				if strings.Contains(line, sinceTime.Format("2006-01-02")) {
					filteredLines = append(filteredLines, line)
				}
			}
			logLines = filteredLines
		}
	}

	return logLines, nil
}

// handleDockerContainerPause handles pausing an individual Docker container
func (h *HTTPServer) handleDockerContainerPause(w http.ResponseWriter, r *http.Request, containerID string) {
	err := h.api.docker.PauseContainer(containerID)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to pause container %s: %v", containerID, err))
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"message":      "Container paused successfully",
		"container_id": containerID,
		"operation":    "pause",
		"timestamp":    time.Now().UTC().Format(time.RFC3339),
	})
}

// handleDockerContainerResume handles resuming an individual Docker container
func (h *HTTPServer) handleDockerContainerResume(w http.ResponseWriter, r *http.Request, containerID string) {
	err := h.api.docker.UnpauseContainer(containerID)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to resume container %s: %v", containerID, err))
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"message":      "Container resumed successfully",
		"container_id": containerID,
		"operation":    "resume",
		"timestamp":    time.Now().UTC().Format(time.RFC3339),
	})
}

// getCPUData returns CPU data in standard format
func (h *HTTPServer) getCPUData() map[string]interface{} {
	timestamp := time.Now().UTC().Format(time.RFC3339)

	cpuInfo, err := h.api.system.GetCPUInfo()
	if err != nil {
		return nil
	}

	response := map[string]interface{}{
		"usage_percent":     cpuInfo.Usage,
		"cores":             cpuInfo.Cores,
		"threads":           cpuInfo.Threads,
		"threads_per_core":  cpuInfo.ThreadsPerCore,
		"sockets":           cpuInfo.Sockets,
		"model":             cpuInfo.Model,
		"frequency_mhz":     cpuInfo.Frequency,
		"max_frequency_mhz": cpuInfo.MaxFrequency,
		"min_frequency_mhz": cpuInfo.MinFrequency,
		"temperature":       cpuInfo.Temperature,
		"last_updated":      timestamp,
	}

	// Get CPU load averages
	if loadInfo, err := h.api.system.GetLoadInfo(); err == nil {
		response["load_1min"] = loadInfo.Load1
		response["load_5min"] = loadInfo.Load5
		response["load_15min"] = loadInfo.Load15
	}

	return response
}

// getMemoryData returns memory data in standard format
func (h *HTTPServer) getMemoryData() map[string]interface{} {
	timestamp := time.Now().UTC().Format(time.RFC3339)

	memInfo, err := h.api.system.GetMemoryInfo()
	if err != nil {
		return nil
	}

	response := map[string]interface{}{
		"usage_percent":       memInfo.UsedPercent,
		"total_bytes":         memInfo.Total,
		"used_bytes":          memInfo.Used,
		"free_bytes":          memInfo.Free,
		"available_bytes":     memInfo.Available,
		"buffers_bytes":       memInfo.Buffers,
		"cached_bytes":        memInfo.Cached,
		"total_formatted":     memInfo.TotalFormatted,
		"used_formatted":      memInfo.UsedFormatted,
		"free_formatted":      memInfo.FreeFormatted,
		"available_formatted": memInfo.AvailableFormatted,
		"buffers_formatted":   memInfo.BuffersFormatted,
		"cached_formatted":    memInfo.CachedFormatted,
		"last_updated":        timestamp,
	}

	// Add memory breakdown if available
	if memInfo.Breakdown != nil {
		response["breakdown"] = map[string]interface{}{
			"system_bytes":        memInfo.Breakdown.System,
			"system_formatted":    memInfo.Breakdown.SystemFormatted,
			"system_percent":      memInfo.Breakdown.SystemPercent,
			"vm_bytes":            memInfo.Breakdown.VM,
			"vm_formatted":        memInfo.Breakdown.VMFormatted,
			"vm_percent":          memInfo.Breakdown.VMPercent,
			"docker_bytes":        memInfo.Breakdown.Docker,
			"docker_formatted":    memInfo.Breakdown.DockerFormatted,
			"docker_percent":      memInfo.Breakdown.DockerPercent,
			"zfs_cache_bytes":     memInfo.Breakdown.ZFSCache,
			"zfs_cache_formatted": memInfo.Breakdown.ZFSCacheFormatted,
			"zfs_cache_percent":   memInfo.Breakdown.ZFSCachePercent,
			"other_bytes":         memInfo.Breakdown.Other,
			"other_formatted":     memInfo.Breakdown.OtherFormatted,
			"other_percent":       memInfo.Breakdown.OtherPercent,
		}
	}

	return response
}

// getTemperatureData returns temperature sensor data in standard format
func (h *HTTPServer) getTemperatureData() map[string]interface{} {
	// Get enhanced temperature data
	enhancedData, err := h.api.system.GetEnhancedTemperatureData()
	if err != nil {
		// Fallback to basic temperature data
		return h.getBasicTemperatureData()
	}

	// Convert enhanced data to API format
	result := make(map[string]interface{})

	// Add sensor chips
	if len(enhancedData.Sensors) > 0 {
		result["sensor_chips"] = enhancedData.Sensors
	}

	// Add standalone fan data
	if len(enhancedData.Fans) > 0 {
		result["fans"] = enhancedData.Fans
	}

	// Add backward compatibility sensors array
	sensors := make([]map[string]interface{}, 0)
	for chipName, chip := range enhancedData.Sensors {
		for _, temp := range chip.Temperatures {
			sensors = append(sensors, map[string]interface{}{
				"name":  fmt.Sprintf("%s - %s", chipName, temp.Label),
				"value": float64(temp.Value) / 1000.0, // Convert back to Celsius for compatibility
				"unit":  "°C",
				"chip":  chipName,
			})
		}
	}

	// Add CPU temperature from system plugin for compatibility
	if cpuInfo, err := h.api.system.GetCPUInfo(); err == nil && cpuInfo.Temperature > 0 {
		sensors = append(sensors, map[string]interface{}{
			"name":  "CPU",
			"value": cpuInfo.Temperature,
			"unit":  "°C",
			"chip":  "system",
		})
	}

	if len(sensors) > 0 {
		result["sensors"] = sensors
	}

	return result
}

// getBasicTemperatureData returns basic temperature data as fallback
func (h *HTTPServer) getBasicTemperatureData() map[string]interface{} {
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
		prefs := dto.Prefs{Unit: "C"} // Default to Celsius
		samples := h.api.sensor.GetReadings(prefs)
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

// getNetworkData returns network statistics in standard format
func (h *HTTPServer) getNetworkData() map[string]interface{} {
	networkInfo, err := h.api.system.GetNetworkInfo()
	if err != nil {
		return nil
	}

	interfaces := make([]map[string]interface{}, 0)
	for _, netInfo := range networkInfo {
		interfaceData := map[string]interface{}{
			"name":       netInfo.Interface,
			"rx_bytes":   netInfo.BytesRecv,
			"tx_bytes":   netInfo.BytesSent,
			"rx_packets": netInfo.PacketsRecv,
			"tx_packets": netInfo.PacketsSent,
			"rx_errors":  netInfo.ErrorsRecv,
			"tx_errors":  netInfo.ErrorsSent,
			"connected":  netInfo.Connected,
		}

		// Add speed and duplex if available
		if netInfo.SpeedMbps > 0 {
			interfaceData["speed_mbps"] = netInfo.SpeedMbps
		}
		if netInfo.Duplex != "" {
			interfaceData["duplex"] = netInfo.Duplex
		}

		interfaces = append(interfaces, interfaceData)
	}

	return map[string]interface{}{
		"interfaces":   interfaces,
		"last_updated": time.Now().UTC().Format(time.RFC3339),
	}
}

// getUPSData returns UPS information in standard format
func (h *HTTPServer) getUPSData() map[string]interface{} {
	// Default UPS data structure
	upsData := map[string]interface{}{
		"status":         "unknown",
		"battery_charge": 0,
		"runtime":        0,
		"load_percent":   0,
		"input_voltage":  0,
		"output_voltage": 0,
		"model":          "unknown",
		"connected":      false,
		"last_updated":   time.Now().UTC().Format(time.RFC3339),
	}

	// Always check for UPS hardware regardless of config setting
	// Check for APC UPS daemon first (most common)
	if _, err := os.Stat("/var/run/apcupsd.pid"); err == nil {
		// apcupsd is running, try to get UPS data
		if apcData := h.getAPCUPSData(); apcData != nil {
			return apcData
		}
	}

	// Check for NUT (Network UPS Tools) as fallback
	if _, err := os.Stat("/var/run/nut/upsmon.pid"); err == nil {
		// NUT is running, try to get UPS data
		if nutData := h.getNUTUPSData(); nutData != nil {
			return nutData
		}
	}

	// Fallback to legacy UPS API if available
	if h.api.ups != nil {
		samples := h.api.ups.GetStatus()
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
		upsData["connected"] = true
	}

	return upsData
}

// isCommandBlacklisted checks if a command is blacklisted for security
func (h *HTTPServer) isCommandBlacklisted(command string) bool {
	// List of dangerous commands that should not be allowed
	blacklistedCommands := []string{
		"rm -rf /",
		"dd if=/dev/zero",
		":(){ :|:& };:", // Fork bomb
		"mkfs",
		"fdisk",
		"parted",
		"format",
		"shutdown",
		"reboot",
		"halt",
		"poweroff",
		"init 0",
		"init 6",
		"telinit",
		"systemctl poweroff",
		"systemctl reboot",
		"systemctl halt",
	}

	// Convert command to lowercase for comparison
	lowerCommand := strings.ToLower(strings.TrimSpace(command))

	// Check against blacklisted commands
	for _, blacklisted := range blacklistedCommands {
		if strings.Contains(lowerCommand, strings.ToLower(blacklisted)) {
			return true
		}
	}

	// Additional security checks
	if strings.Contains(lowerCommand, ">/dev/") ||
		strings.Contains(lowerCommand, "rm -rf") ||
		strings.Contains(lowerCommand, "chmod 777") ||
		strings.Contains(lowerCommand, "chown root") ||
		strings.Contains(lowerCommand, "sudo su") ||
		strings.Contains(lowerCommand, "su -") {
		return true
	}

	return false
}

// executeCommand executes a system command with timeout and returns the result
func (h *HTTPServer) executeCommand(request CommandExecuteRequest) CommandExecuteResponse {
	startTime := time.Now()

	// Prepare command
	cmd := exec.Command("/bin/bash", "-c", request.Command)

	// Set working directory if specified
	if request.WorkingDirectory != "" {
		if _, err := os.Stat(request.WorkingDirectory); err == nil {
			cmd.Dir = request.WorkingDirectory
		}
	}

	// Set timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(request.Timeout)*time.Second)
	defer cancel()
	cmd = exec.CommandContext(ctx, "/bin/bash", "-c", request.Command)

	// Capture stdout and stderr
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Execute command
	err := cmd.Run()
	executionTime := time.Since(startTime)

	// Prepare response
	response := CommandExecuteResponse{
		Command:         request.Command,
		Stdout:          stdout.String(),
		Stderr:          stderr.String(),
		ExecutionTimeMs: executionTime.Milliseconds(),
		WorkingDir:      request.WorkingDirectory,
	}

	// Set exit code
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			response.ExitCode = exitError.ExitCode()
		} else {
			response.ExitCode = -1
		}
	} else {
		response.ExitCode = 0
	}

	return response
}

// getAPCUPSData retrieves UPS data from apcupsd daemon
func (h *HTTPServer) getAPCUPSData() map[string]interface{} {
	// Execute apcaccess command to get UPS status
	cmd := exec.Command("apcaccess")
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	// Parse apcaccess output
	upsData := map[string]interface{}{
		"status":         "unknown",
		"battery_charge": 0.0,
		"runtime":        0.0,
		"load_percent":   0.0,
		"input_voltage":  0.0,
		"output_voltage": 0.0,
		"model":          "unknown",
		"connected":      true,
		"ups_type":       "apc",
		"last_updated":   time.Now().UTC().Format(time.RFC3339),
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Split on colon and spaces
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "STATUS":
			upsData["status"] = strings.ToLower(value)
		case "BCHARGE":
			if charge, err := strconv.ParseFloat(strings.TrimSuffix(value, " Percent"), 64); err == nil {
				upsData["battery_charge"] = charge
			}
		case "TIMELEFT":
			if runtime, err := strconv.ParseFloat(strings.TrimSuffix(value, " Minutes"), 64); err == nil {
				upsData["runtime"] = runtime
			}
		case "LOADPCT":
			if load, err := strconv.ParseFloat(strings.TrimSuffix(value, " Percent"), 64); err == nil {
				upsData["load_percent"] = load
			}
		case "LINEV":
			if voltage, err := strconv.ParseFloat(strings.TrimSuffix(value, " Volts"), 64); err == nil {
				upsData["input_voltage"] = voltage
			}
		case "OUTPUTV":
			if voltage, err := strconv.ParseFloat(strings.TrimSuffix(value, " Volts"), 64); err == nil {
				upsData["output_voltage"] = voltage
			}
		case "MODEL":
			upsData["model"] = value
		case "UPSNAME":
			upsData["name"] = value
		case "SERIALNO":
			upsData["serial_number"] = value
		case "NOMPOWER":
			if power, err := strconv.ParseFloat(strings.TrimSuffix(value, " Watts"), 64); err == nil {
				upsData["nominal_power"] = power
			}
		}
	}

	return upsData
}

// getNUTUPSData retrieves UPS data from NUT (Network UPS Tools)
func (h *HTTPServer) getNUTUPSData() map[string]interface{} {
	// Execute upsc command to get UPS status
	cmd := exec.Command("upsc", "ups")
	output, err := cmd.Output()
	if err != nil {
		// Try with different UPS name
		cmd = exec.Command("upsc", "ups@localhost")
		output, err = cmd.Output()
		if err != nil {
			return nil
		}
	}

	// Parse upsc output
	upsData := map[string]interface{}{
		"status":         "unknown",
		"battery_charge": 0.0,
		"runtime":        0.0,
		"load_percent":   0.0,
		"input_voltage":  0.0,
		"output_voltage": 0.0,
		"model":          "unknown",
		"connected":      true,
		"ups_type":       "nut",
		"last_updated":   time.Now().UTC().Format(time.RFC3339),
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Split on colon and spaces
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "ups.status":
			upsData["status"] = strings.ToLower(value)
		case "battery.charge":
			if charge, err := strconv.ParseFloat(value, 64); err == nil {
				upsData["battery_charge"] = charge
			}
		case "battery.runtime":
			if runtime, err := strconv.ParseFloat(value, 64); err == nil {
				upsData["runtime"] = runtime / 60 // Convert seconds to minutes
			}
		case "ups.load":
			if load, err := strconv.ParseFloat(value, 64); err == nil {
				upsData["load_percent"] = load
			}
		case "input.voltage":
			if voltage, err := strconv.ParseFloat(value, 64); err == nil {
				upsData["input_voltage"] = voltage
			}
		case "output.voltage":
			if voltage, err := strconv.ParseFloat(value, 64); err == nil {
				upsData["output_voltage"] = voltage
			}
		case "ups.model":
			upsData["model"] = value
		case "ups.serial":
			upsData["serial_number"] = value
		}
	}

	return upsData
}

// getIntelGPUData returns Intel GPU data in standard format
func (h *HTTPServer) getIntelGPUData() map[string]interface{} {
	gpuInfo, err := h.api.gpu.GetGPUInfo()
	if err != nil {
		return nil
	}

	// Look for Intel GPU
	for _, gpu := range gpuInfo {
		if strings.Contains(strings.ToLower(gpu.Name), "intel") {
			return map[string]interface{}{
				"usage":        gpu.UtilizationGPU,
				"temperature":  gpu.Temperature,
				"name":         gpu.Name,
				"memory_used":  gpu.MemoryUsed,
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

// getFilesystemData returns filesystem usage data in standard format
func (h *HTTPServer) getFilesystemData() map[string]interface{} {
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

// getDockerData returns Docker container data in standard format
func (h *HTTPServer) getDockerData() []map[string]interface{} {
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
			cpuUsage = stats.CPUPercent
			memoryUsage = stats.MemUsage
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

// getVMData returns VM data in standard format
func (h *HTTPServer) getVMData() []map[string]interface{} {
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
			cpuUsage = stats.CPUUsage
			memoryUsage = stats.MemoryUsage
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

	// Parse query parameters
	all := r.URL.Query().Get("all") == "true"
	params := h.parsePaginationParams(r)
	usePagination := r.URL.Query().Get("page") != "" || r.URL.Query().Get("limit") != ""

	containers, err := h.api.docker.ListContainers(all)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to list containers: %v", err))
		return
	}

	if usePagination {
		// Convert containers to interface{} slice for pagination
		containerList := make([]interface{}, len(containers))
		for i, container := range containers {
			containerList[i] = container
		}

		paginatedContainers, total := h.paginateSlice(containerList, params)
		h.writePaginatedResponse(w, http.StatusOK, paginatedContainers, total, params)
	} else {
		// Return original format for backward compatibility
		h.writeJSON(w, http.StatusOK, containers)
	}
}

// handleDockerNetworks handles GET /api/v1/docker/networks
func (h *HTTPServer) handleDockerNetworks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	networks, err := h.api.docker.ListNetworks()
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to list networks: %v", err))
		return
	}

	h.writeJSON(w, http.StatusOK, networks)
}

// handleDockerImages handles GET /api/v1/docker/images
func (h *HTTPServer) handleDockerImages(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	images, err := h.api.docker.ListImages()
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to list images: %v", err))
		return
	}

	h.writeJSON(w, http.StatusOK, images)
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
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/vms/")
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

		case "hibernate":
			err := h.api.vm.HibernateVM(vmName)
			if err != nil {
				h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to hibernate VM: %v", err))
				return
			}
			h.writeJSON(w, http.StatusOK, map[string]string{"message": "VM hibernated"})

		case "restore":
			err := h.api.vm.RestoreVM(vmName)
			if err != nil {
				h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to restore VM: %v", err))
				return
			}
			h.writeJSON(w, http.StatusOK, map[string]string{"message": "VM restored"})

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

// Array Control Handlers

// handleArrayStart handles POST /api/v1/array/start
func (h *HTTPServer) handleArrayStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req ArrayStartRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid JSON request")
		return
	}

	// Check current array state
	arrayInfo, err := h.api.storage.GetArrayInfo()
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get array state: %v", err))
		return
	}

	if arrayInfo.State == "started" {
		response := ArrayOperationResponse{
			Success: false,
			Message: "Array is already started",
		}
		h.writeJSON(w, http.StatusConflict, response)
		return
	}

	// Start the array
	err = h.api.storage.StartArray(req.MaintenanceMode, req.CheckFilesystem)
	if err != nil {
		response := ArrayOperationResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to start array: %v", err),
		}
		h.writeJSON(w, http.StatusInternalServerError, response)
		return
	}

	response := ArrayOperationResponse{
		Success:       true,
		Message:       "Array start initiated",
		OperationID:   fmt.Sprintf("array_start_%d", time.Now().Unix()),
		EstimatedTime: 30, // seconds
	}
	h.writeJSON(w, http.StatusOK, response)
}

// handleArrayStop handles POST /api/v1/array/stop
func (h *HTTPServer) handleArrayStop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req ArrayStopRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid JSON request")
		return
	}

	// Check current array state
	arrayInfo, err := h.api.storage.GetArrayInfo()
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get array state: %v", err))
		return
	}

	if arrayInfo.State == "stopped" {
		response := ArrayOperationResponse{
			Success: false,
			Message: "Array is already stopped",
		}
		h.writeJSON(w, http.StatusConflict, response)
		return
	}

	// Stop the array with enhanced orchestration
	err = h.api.storage.StopArray(req.Force, req.UnmountShares, req.StopContainers, req.StopVMs)
	if err != nil {
		response := ArrayOperationResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to stop array: %v", err),
		}
		h.writeJSON(w, http.StatusInternalServerError, response)
		return
	}

	response := ArrayOperationResponse{
		Success:       true,
		Message:       "Array stop initiated",
		OperationID:   fmt.Sprintf("array_stop_%d", time.Now().Unix()),
		EstimatedTime: 15, // seconds
	}
	h.writeJSON(w, http.StatusOK, response)
}

// handleArrayParityCheck handles GET/POST/DELETE /api/v1/array/parity-check
func (h *HTTPServer) handleArrayParityCheck(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// Get parity check status
		status, err := h.api.storage.GetParityCheckStatus()
		if err != nil {
			h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get parity check status: %v", err))
			return
		}
		h.writeJSON(w, http.StatusOK, status)

	case http.MethodPost:
		// Start parity check
		var req ParityCheckRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.writeError(w, http.StatusBadRequest, "Invalid JSON request")
			return
		}

		// Validate request
		if req.Type != "check" && req.Type != "correct" {
			h.writeError(w, http.StatusBadRequest, "Invalid parity check type. Must be 'check' or 'correct'")
			return
		}

		if req.Priority != "low" && req.Priority != "normal" && req.Priority != "high" {
			h.writeError(w, http.StatusBadRequest, "Invalid priority. Must be 'low', 'normal', or 'high'")
			return
		}

		err := h.api.storage.StartParityCheck(req.Type, req.Priority)
		if err != nil {
			response := ArrayOperationResponse{
				Success: false,
				Message: fmt.Sprintf("Failed to start parity check: %v", err),
			}
			h.writeJSON(w, http.StatusInternalServerError, response)
			return
		}

		response := ArrayOperationResponse{
			Success:       true,
			Message:       fmt.Sprintf("Parity %s started", req.Type),
			OperationID:   fmt.Sprintf("parity_%s_%d", req.Type, time.Now().Unix()),
			EstimatedTime: 3600, // 1 hour estimate
		}
		h.writeJSON(w, http.StatusOK, response)

	case http.MethodDelete:
		// Cancel parity check
		err := h.api.storage.CancelParityCheck()
		if err != nil {
			response := ArrayOperationResponse{
				Success: false,
				Message: fmt.Sprintf("Failed to cancel parity check: %v", err),
			}
			h.writeJSON(w, http.StatusInternalServerError, response)
			return
		}

		response := ArrayOperationResponse{
			Success: true,
			Message: "Parity check cancelled",
		}
		h.writeJSON(w, http.StatusOK, response)

	default:
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// handleArrayDiskAdd handles POST /api/v1/array/disk/add
func (h *HTTPServer) handleArrayDiskAdd(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req DiskAddRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid JSON request")
		return
	}

	// Validate request
	if req.Device == "" || req.Position == "" {
		h.writeError(w, http.StatusBadRequest, "Device and position are required")
		return
	}

	err := h.api.storage.AddDisk(req.Device, req.Position)
	if err != nil {
		response := ArrayOperationResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to add disk: %v", err),
		}
		h.writeJSON(w, http.StatusInternalServerError, response)
		return
	}

	response := ArrayOperationResponse{
		Success:       true,
		Message:       fmt.Sprintf("Disk %s added to position %s", req.Device, req.Position),
		OperationID:   fmt.Sprintf("disk_add_%d", time.Now().Unix()),
		EstimatedTime: 10, // seconds
	}
	h.writeJSON(w, http.StatusOK, response)
}

// handleArrayDiskRemove handles POST /api/v1/array/disk/remove
func (h *HTTPServer) handleArrayDiskRemove(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req DiskRemoveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid JSON request")
		return
	}

	// Validate request
	if req.Position == "" {
		h.writeError(w, http.StatusBadRequest, "Position is required")
		return
	}

	err := h.api.storage.RemoveDisk(req.Position)
	if err != nil {
		response := ArrayOperationResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to remove disk: %v", err),
		}
		h.writeJSON(w, http.StatusInternalServerError, response)
		return
	}

	response := ArrayOperationResponse{
		Success:       true,
		Message:       fmt.Sprintf("Disk removed from position %s", req.Position),
		OperationID:   fmt.Sprintf("disk_remove_%d", time.Now().Unix()),
		EstimatedTime: 10, // seconds
	}
	h.writeJSON(w, http.StatusOK, response)
}

// handleSystemSleep handles POST /api/v1/system/sleep
func (h *HTTPServer) handleSystemSleep(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req SystemSleepRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid JSON request")
		return
	}

	// Validate sleep type
	if req.Type != "suspend" && req.Type != "hibernate" && req.Type != "hybrid" {
		h.writeError(w, http.StatusBadRequest, "Invalid sleep type. Must be 'suspend', 'hibernate', or 'hybrid'")
		return
	}

	// Execute sleep
	err := h.executeSystemSleep(req.Type)
	if err != nil {
		response := PowerOperationResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to %s system: %v", req.Type, err),
		}
		h.writeJSON(w, http.StatusInternalServerError, response)
		return
	}

	response := PowerOperationResponse{
		Success:     true,
		Message:     fmt.Sprintf("System %s initiated", req.Type),
		OperationID: fmt.Sprintf("%s_%d", req.Type, time.Now().Unix()),
	}
	h.writeJSON(w, http.StatusOK, response)
}

// handleSystemWake handles POST /api/v1/system/wake
func (h *HTTPServer) handleSystemWake(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req SystemWakeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid JSON request")
		return
	}

	// Validate MAC address
	if req.TargetMAC == "" {
		h.writeError(w, http.StatusBadRequest, "Target MAC address is required")
		return
	}

	// Set defaults
	if req.BroadcastIP == "" {
		req.BroadcastIP = "255.255.255.255"
	}
	if req.Port == 0 {
		req.Port = 9
	}
	if req.RepeatCount == 0 {
		req.RepeatCount = 3
	}

	// Execute Wake-on-LAN
	err := h.executeWakeOnLAN(req.TargetMAC, req.BroadcastIP, req.Port, req.RepeatCount)
	if err != nil {
		response := PowerOperationResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to send Wake-on-LAN packet: %v", err),
		}
		h.writeJSON(w, http.StatusInternalServerError, response)
		return
	}

	response := PowerOperationResponse{
		Success:     true,
		Message:     fmt.Sprintf("Wake-on-LAN packet sent to %s", req.TargetMAC),
		OperationID: fmt.Sprintf("wake_%d", time.Now().Unix()),
	}
	h.writeJSON(w, http.StatusOK, response)
}

// Power Management Execution Functions

// executeSystemShutdown executes a system shutdown with the specified parameters
func (h *HTTPServer) executeSystemShutdown(delaySeconds int, message string, force bool) error {
	// Build shutdown command
	cmd := "shutdown"

	// Add delay
	if delaySeconds > 0 {
		cmd += fmt.Sprintf(" +%d", delaySeconds/60) // Convert to minutes for shutdown command
	} else {
		cmd += " now"
	}

	// Add message if provided
	if message != "" {
		cmd += fmt.Sprintf(" \"%s\"", message)
	}

	// Add force flag if needed
	if force {
		cmd = "shutdown -f" + cmd[8:] // Replace "shutdown" with "shutdown -f"
	}

	// Execute the command in background
	go func() {
		if delaySeconds > 0 && delaySeconds < 60 {
			// For delays less than 1 minute, use sleep + immediate shutdown
			time.Sleep(time.Duration(delaySeconds) * time.Second)
			cmd = "shutdown now"
			if message != "" {
				cmd += fmt.Sprintf(" \"%s\"", message)
			}
		}

		// Execute shutdown command
		output := lib.GetCmdOutput("sh", "-c", cmd)
		for _, line := range output {
			if strings.Contains(strings.ToLower(line), "error") {
				logger.Red("Shutdown error: %s", line)
			}
		}
	}()

	return nil
}

// executeSystemReboot executes a system reboot with the specified parameters
func (h *HTTPServer) executeSystemReboot(delaySeconds int, message string, force bool) error {
	// Build reboot command
	cmd := "reboot"

	// Add force flag if needed
	if force {
		cmd = "reboot -f"
	}

	// Execute the command in background with delay
	go func() {
		if delaySeconds > 0 {
			time.Sleep(time.Duration(delaySeconds) * time.Second)
		}

		// Send wall message if provided
		if message != "" {
			wallCmd := fmt.Sprintf("wall \"%s - System rebooting now\"", message)
			lib.GetCmdOutput("sh", "-c", wallCmd)
			time.Sleep(2 * time.Second) // Give time for message to be displayed
		}

		// Execute reboot command
		output := lib.GetCmdOutput("sh", "-c", cmd)
		for _, line := range output {
			if strings.Contains(strings.ToLower(line), "error") {
				logger.Red("Reboot error: %s", line)
			}
		}
	}()

	return nil
}

// executeSystemSleep executes a system sleep/suspend operation
func (h *HTTPServer) executeSystemSleep(sleepType string) error {
	var cmd string

	switch sleepType {
	case "suspend":
		cmd = "systemctl suspend"
	case "hibernate":
		cmd = "systemctl hibernate"
	case "hybrid":
		cmd = "systemctl hybrid-sleep"
	default:
		return fmt.Errorf("invalid sleep type: %s", sleepType)
	}

	// Execute the command in background
	go func() {
		output := lib.GetCmdOutput("sh", "-c", cmd)
		for _, line := range output {
			if strings.Contains(strings.ToLower(line), "error") {
				logger.Red("Sleep error: %s", line)
			}
		}
	}()

	return nil
}

// executeWakeOnLAN sends a Wake-on-LAN packet to the specified MAC address
func (h *HTTPServer) executeWakeOnLAN(targetMAC, broadcastIP string, port, repeatCount int) error {
	// Validate and parse MAC address
	macBytes, err := parseMACAddress(targetMAC)
	if err != nil {
		return fmt.Errorf("invalid MAC address: %v", err)
	}

	// Create Wake-on-LAN packet
	packet := createWOLPacket(macBytes)

	// Send packets
	for i := 0; i < repeatCount; i++ {
		err := sendWOLPacket(packet, broadcastIP, port)
		if err != nil {
			return fmt.Errorf("failed to send WOL packet %d: %v", i+1, err)
		}

		// Small delay between packets
		if i < repeatCount-1 {
			time.Sleep(100 * time.Millisecond)
		}
	}

	return nil
}

// parseMACAddress parses a MAC address string into bytes
func parseMACAddress(mac string) ([]byte, error) {
	// Remove common separators
	mac = strings.ReplaceAll(mac, ":", "")
	mac = strings.ReplaceAll(mac, "-", "")
	mac = strings.ReplaceAll(mac, ".", "")

	// Validate length
	if len(mac) != 12 {
		return nil, fmt.Errorf("MAC address must be 12 hex characters")
	}

	// Parse hex string
	macBytes := make([]byte, 6)
	for i := 0; i < 6; i++ {
		b, err := strconv.ParseUint(mac[i*2:i*2+2], 16, 8)
		if err != nil {
			return nil, fmt.Errorf("invalid hex in MAC address: %v", err)
		}
		macBytes[i] = byte(b)
	}

	return macBytes, nil
}

// createWOLPacket creates a Wake-on-LAN magic packet
func createWOLPacket(macBytes []byte) []byte {
	// WOL packet: 6 bytes of 0xFF followed by 16 repetitions of the MAC address
	packet := make([]byte, 102) // 6 + (6 * 16) = 102 bytes

	// Fill first 6 bytes with 0xFF
	for i := 0; i < 6; i++ {
		packet[i] = 0xFF
	}

	// Repeat MAC address 16 times
	for i := 0; i < 16; i++ {
		copy(packet[6+i*6:6+(i+1)*6], macBytes)
	}

	return packet
}

// sendWOLPacket sends a Wake-on-LAN packet via UDP
func sendWOLPacket(packet []byte, broadcastIP string, port int) error {
	// Create UDP address
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", broadcastIP, port))
	if err != nil {
		return fmt.Errorf("failed to resolve UDP address: %v", err)
	}

	// Create UDP connection
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return fmt.Errorf("failed to create UDP connection: %v", err)
	}
	defer conn.Close()

	// Send packet
	_, err = conn.Write(packet)
	if err != nil {
		return fmt.Errorf("failed to send packet: %v", err)
	}

	return nil
}

// User Script Management Handlers

// handleScripts handles GET /api/v1/scripts
func (h *HTTPServer) handleScripts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	scripts, err := h.getUserScripts()
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get user scripts: %v", err))
		return
	}

	response := ScriptListResponse{Scripts: scripts}
	h.writeJSON(w, http.StatusOK, response)
}

// handleScript handles script operations
func (h *HTTPServer) handleScript(w http.ResponseWriter, r *http.Request) {
	// Extract script name from URL path
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/scripts/")
	parts := strings.Split(path, "/")
	if len(parts) == 0 || parts[0] == "" {
		h.writeError(w, http.StatusBadRequest, "Script name required")
		return
	}

	scriptName := parts[0]
	action := ""
	if len(parts) > 1 {
		action = parts[1]
	}

	switch r.Method {
	case http.MethodGet:
		switch action {
		case "status":
			status, err := h.getScriptStatus(scriptName)
			if err != nil {
				h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get script status: %v", err))
				return
			}
			h.writeJSON(w, http.StatusOK, status)

		case "logs":
			logs, err := h.getScriptLogs(scriptName)
			if err != nil {
				h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get script logs: %v", err))
				return
			}
			response := ScriptLogsResponse{Name: scriptName, Logs: logs}
			h.writeJSON(w, http.StatusOK, response)

		default:
			h.writeError(w, http.StatusBadRequest, "Invalid action. Use 'status' or 'logs'")
		}

	case http.MethodPost:
		switch action {
		case "execute":
			var req ScriptExecuteRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				h.writeError(w, http.StatusBadRequest, "Invalid JSON request")
				return
			}

			response, err := h.executeScript(scriptName, req)
			if err != nil {
				h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to execute script: %v", err))
				return
			}
			h.writeJSON(w, http.StatusOK, response)

		case "stop":
			err := h.stopScript(scriptName)
			if err != nil {
				h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to stop script: %v", err))
				return
			}
			h.writeJSON(w, http.StatusOK, map[string]string{"message": "Script stopped successfully"})

		default:
			h.writeError(w, http.StatusBadRequest, "Invalid action. Use 'execute' or 'stop'")
		}

	default:
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// Share Management Handlers

// handleShares handles GET /api/v1/shares
func (h *HTTPServer) handleShares(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		shares, err := h.getShares()
		if err != nil {
			h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get shares: %v", err))
			return
		}

		response := ShareListResponse{Shares: shares}
		h.writeJSON(w, http.StatusOK, response)
		return
	}

	if r.Method == http.MethodPost {
		var req ShareCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.writeError(w, http.StatusBadRequest, "Invalid JSON request")
			return
		}

		if err := h.validateShareCreateRequest(&req); err != nil {
			h.writeError(w, http.StatusBadRequest, fmt.Sprintf("Invalid request: %v", err))
			return
		}

		if err := h.createShare(&req); err != nil {
			h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to create share: %v", err))
			return
		}

		response := ShareOperationResponse{
			Success:   true,
			Message:   fmt.Sprintf("Share '%s' created successfully", req.Name),
			ShareName: req.Name,
		}
		h.writeJSON(w, http.StatusCreated, response)
		return
	}

	h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
}

// handleShare handles share operations
func (h *HTTPServer) handleShare(w http.ResponseWriter, r *http.Request) {
	// Extract share name from URL path
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/shares/")
	parts := strings.Split(path, "/")
	if len(parts) == 0 || parts[0] == "" {
		h.writeError(w, http.StatusBadRequest, "Share name required")
		return
	}

	shareName := parts[0]
	action := ""
	if len(parts) > 1 {
		action = parts[1]
	}

	switch r.Method {
	case http.MethodGet:
		if action == "usage" {
			usage, err := h.getShareUsage(shareName)
			if err != nil {
				h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get share usage: %v", err))
				return
			}
			h.writeJSON(w, http.StatusOK, usage)
		} else if action == "" {
			share, err := h.getShare(shareName)
			if err != nil {
				h.writeError(w, http.StatusNotFound, fmt.Sprintf("Share not found: %v", err))
				return
			}
			h.writeJSON(w, http.StatusOK, share)
		} else {
			h.writeError(w, http.StatusBadRequest, "Invalid action. Use 'usage' or omit for share details")
		}

	case http.MethodPut:
		if action != "" {
			h.writeError(w, http.StatusBadRequest, "Action not allowed for PUT requests")
			return
		}

		var req ShareUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.writeError(w, http.StatusBadRequest, "Invalid JSON request")
			return
		}

		if err := h.validateShareUpdateRequest(&req); err != nil {
			h.writeError(w, http.StatusBadRequest, fmt.Sprintf("Invalid request: %v", err))
			return
		}

		if err := h.updateShare(shareName, &req); err != nil {
			h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to update share: %v", err))
			return
		}

		response := ShareOperationResponse{
			Success:   true,
			Message:   fmt.Sprintf("Share '%s' updated successfully", shareName),
			ShareName: shareName,
		}
		h.writeJSON(w, http.StatusOK, response)

	case http.MethodDelete:
		if action != "" {
			h.writeError(w, http.StatusBadRequest, "Action not allowed for DELETE requests")
			return
		}

		if err := h.deleteShare(shareName); err != nil {
			h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to delete share: %v", err))
			return
		}

		response := ShareOperationResponse{
			Success:   true,
			Message:   fmt.Sprintf("Share '%s' deleted successfully", shareName),
			ShareName: shareName,
		}
		h.writeJSON(w, http.StatusOK, response)

	default:
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// getScriptDescription reads the script description from the description file
func (h *HTTPServer) getScriptDescription(scriptName string) string {
	descPath := fmt.Sprintf("/boot/config/plugins/user.scripts/scripts/%s/description", scriptName)
	content, err := os.ReadFile(descPath)
	if err != nil {
		return "No description available"
	}
	return strings.TrimSpace(string(content))
}

// getScriptCurrentStatus returns the current status of a script
func (h *HTTPServer) getScriptCurrentStatus(scriptName string) string {
	// Check if script is currently running by looking for PID file
	pidPath := fmt.Sprintf("/tmp/user.scripts.%s.pid", scriptName)
	if pidData, err := os.ReadFile(pidPath); err == nil {
		pidStr := strings.TrimSpace(string(pidData))
		if pid, err := strconv.Atoi(pidStr); err == nil {
			// Check if process is still running
			if h.isProcessRunning(pid) {
				return "running"
			}
		}
	}

	return "idle"
}

// getScriptLastRun returns the last run time and result of a script
func (h *HTTPServer) getScriptLastRun(scriptName string) (string, string) {
	// Check for log file to determine last run
	logPath := fmt.Sprintf("/tmp/user.scripts/tmpScripts/%s.log", scriptName)
	if stat, err := os.Stat(logPath); err == nil {
		lastRun := stat.ModTime().Format(time.RFC3339)

		// Try to determine result from log content
		if content, err := os.ReadFile(logPath); err == nil {
			logContent := string(content)
			if strings.Contains(logContent, "error") || strings.Contains(logContent, "failed") {
				return lastRun, "failed"
			}
			return lastRun, "success"
		}

		return lastRun, "unknown"
	}

	return "", "unknown"
}

// isProcessRunning checks if a process with the given PID is still running
func (h *HTTPServer) isProcessRunning(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	// On Unix systems, sending signal 0 checks if process exists
	err = process.Signal(syscall.Signal(0))
	return err == nil
}

// getScriptStatus returns detailed status information for a script
func (h *HTTPServer) getScriptStatus(scriptName string) (*ScriptStatusResponse, error) {
	status := &ScriptStatusResponse{
		Name:   scriptName,
		Status: "idle",
	}

	// Check if script is currently running
	pidPath := fmt.Sprintf("/tmp/user.scripts.%s.pid", scriptName)
	if pidData, err := os.ReadFile(pidPath); err == nil {
		pidStr := strings.TrimSpace(string(pidData))
		if pid, err := strconv.Atoi(pidStr); err == nil {
			status.PID = pid
			if h.isProcessRunning(pid) {
				status.Status = "running"

				// Try to get start time from process
				if startTime, err := h.getProcessStartTime(pid); err == nil {
					status.StartTime = startTime.Format(time.RFC3339)
					status.Duration = time.Since(startTime).String()
				}
			} else {
				// Process not running, check for exit code
				status.Status = "completed"
				if exitCode, err := h.getScriptExitCode(scriptName); err == nil {
					status.ExitCode = exitCode
					if exitCode != 0 {
						status.Status = "failed"
					}
				}
			}
		}
	}

	return status, nil
}

// getScriptLogs returns the log output for a script
func (h *HTTPServer) getScriptLogs(scriptName string) ([]string, error) {
	logPath := fmt.Sprintf("/tmp/user.scripts/tmpScripts/%s.log", scriptName)

	content, err := os.ReadFile(logPath)
	if err != nil {
		// Return empty logs if file doesn't exist
		return []string{}, nil
	}

	// Split content into lines
	lines := strings.Split(string(content), "\n")

	// Remove empty lines at the end
	for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}

	return lines, nil
}

// executeScript executes a user script with the given parameters
func (h *HTTPServer) executeScript(scriptName string, req ScriptExecuteRequest) (*ScriptExecuteResponse, error) {
	// Validate script exists
	scriptPath := fmt.Sprintf("/boot/config/plugins/user.scripts/scripts/%s/script", scriptName)
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("script '%s' not found", scriptName)
	}

	// Check if script is already running
	if h.getScriptCurrentStatus(scriptName) == "running" {
		return nil, fmt.Errorf("script '%s' is already running", scriptName)
	}

	// Build command arguments
	args := []string{scriptPath}
	args = append(args, req.Arguments...)

	// Create execution ID
	executionID := fmt.Sprintf("%s_%d", scriptName, time.Now().Unix())

	// Execute script
	if req.Background {
		// Execute in background
		cmd := exec.Command("/bin/bash", args...)

		// Set up logging
		logPath := fmt.Sprintf("/tmp/user.scripts/tmpScripts/%s.log", scriptName)
		logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to create log file: %v", err)
		}

		cmd.Stdout = logFile
		cmd.Stderr = logFile

		// Start the process
		if err := cmd.Start(); err != nil {
			logFile.Close()
			return nil, fmt.Errorf("failed to start script: %v", err)
		}

		// Save PID
		pidPath := fmt.Sprintf("/tmp/user.scripts.%s.pid", scriptName)
		pidFile, err := os.Create(pidPath)
		if err != nil {
			logFile.Close()
			return nil, fmt.Errorf("failed to create PID file: %v", err)
		}
		fmt.Fprintf(pidFile, "%d", cmd.Process.Pid)
		pidFile.Close()

		// Monitor process completion in background
		go func() {
			defer logFile.Close()
			cmd.Wait()
			// Remove PID file when process completes
			os.Remove(pidPath)
		}()

		return &ScriptExecuteResponse{
			Success:     true,
			Message:     fmt.Sprintf("Script '%s' started successfully", scriptName),
			ExecutionID: executionID,
			PID:         cmd.Process.Pid,
		}, nil
	} else {
		// Execute synchronously
		cmd := exec.Command("/bin/bash", args...)
		output, err := cmd.CombinedOutput()

		// Save output to log file
		logPath := fmt.Sprintf("/tmp/user.scripts/tmpScripts/%s.log", scriptName)
		os.WriteFile(logPath, output, 0644)

		if err != nil {
			return &ScriptExecuteResponse{
				Success:     false,
				Message:     fmt.Sprintf("Script '%s' failed: %v", scriptName, err),
				ExecutionID: executionID,
			}, nil
		}

		return &ScriptExecuteResponse{
			Success:     true,
			Message:     fmt.Sprintf("Script '%s' completed successfully", scriptName),
			ExecutionID: executionID,
		}, nil
	}
}

// stopScript stops a running user script
func (h *HTTPServer) stopScript(scriptName string) error {
	// Check if script is running
	pidPath := fmt.Sprintf("/tmp/user.scripts.%s.pid", scriptName)
	pidData, err := os.ReadFile(pidPath)
	if err != nil {
		return fmt.Errorf("script '%s' is not running", scriptName)
	}

	pidStr := strings.TrimSpace(string(pidData))
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		return fmt.Errorf("invalid PID file for script '%s'", scriptName)
	}

	// Find and terminate the process
	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("failed to find process %d: %v", pid, err)
	}

	// Send SIGTERM first
	if err := process.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("failed to terminate process %d: %v", pid, err)
	}

	// Wait a bit for graceful shutdown
	time.Sleep(2 * time.Second)

	// Check if process is still running
	if h.isProcessRunning(pid) {
		// Force kill with SIGKILL
		if err := process.Signal(syscall.SIGKILL); err != nil {
			return fmt.Errorf("failed to kill process %d: %v", pid, err)
		}
	}

	// Remove PID file
	os.Remove(pidPath)

	return nil
}

// getProcessStartTime gets the start time of a process (Unix-specific)
func (h *HTTPServer) getProcessStartTime(pid int) (time.Time, error) {
	// Read process stat file
	statPath := fmt.Sprintf("/proc/%d/stat", pid)
	content, err := os.ReadFile(statPath)
	if err != nil {
		return time.Time{}, err
	}

	// Parse stat file (field 22 is start time in clock ticks)
	fields := strings.Fields(string(content))
	if len(fields) < 22 {
		return time.Time{}, fmt.Errorf("invalid stat file format")
	}

	startTicks, err := strconv.ParseInt(fields[21], 10, 64)
	if err != nil {
		return time.Time{}, err
	}

	// Get system boot time
	bootTime, err := h.getSystemBootTime()
	if err != nil {
		return time.Time{}, err
	}

	// Calculate start time (assuming 100 ticks per second)
	startTime := bootTime.Add(time.Duration(startTicks*10) * time.Millisecond)

	return startTime, nil
}

// getSystemBootTime gets the system boot time
func (h *HTTPServer) getSystemBootTime() (time.Time, error) {
	content, err := os.ReadFile("/proc/stat")
	if err != nil {
		return time.Time{}, err
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "btime ") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				bootTimestamp, err := strconv.ParseInt(fields[1], 10, 64)
				if err != nil {
					return time.Time{}, err
				}
				return time.Unix(bootTimestamp, 0), nil
			}
		}
	}

	return time.Time{}, fmt.Errorf("boot time not found in /proc/stat")
}

// getScriptExitCode gets the exit code of a completed script
func (h *HTTPServer) getScriptExitCode(scriptName string) (int, error) {
	// Try to read exit code from a status file (if User Scripts plugin creates one)
	statusPath := fmt.Sprintf("/tmp/user.scripts.%s.status", scriptName)
	if content, err := os.ReadFile(statusPath); err == nil {
		return strconv.Atoi(strings.TrimSpace(string(content)))
	}

	// If no status file, assume success (0) if log exists, error (1) otherwise
	logPath := fmt.Sprintf("/tmp/user.scripts/tmpScripts/%s.log", scriptName)
	if _, err := os.Stat(logPath); err == nil {
		return 0, nil
	}

	return 1, nil
}

// Share Management Implementation Functions

// getShares returns a list of all configured shares
func (h *HTTPServer) getShares() ([]Share, error) {
	var shares []Share

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
			share, err := h.parseShareConfig(shareName)
			if err != nil {
				// Log error but continue with other shares
				continue
			}
			shares = append(shares, *share)
		}
	}

	return shares, nil
}

// getShare returns detailed information for a specific share
func (h *HTTPServer) getShare(shareName string) (*Share, error) {
	return h.parseShareConfig(shareName)
}

// parseShareConfig parses a share configuration file
func (h *HTTPServer) parseShareConfig(shareName string) (*Share, error) {
	configPath := fmt.Sprintf("/boot/config/shares/%s.cfg", shareName)

	content, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("share '%s' not found", shareName)
	}

	share := &Share{
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

// getShareUsage calculates usage statistics for a share
func (h *HTTPServer) getShareUsage(shareName string) (*ShareUsage, error) {
	sharePath := fmt.Sprintf("/mnt/user/%s", shareName)

	// Check if share path exists
	if _, err := os.Stat(sharePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("share '%s' not found", shareName)
	}

	usage := &ShareUsage{
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
		fileCount, dirCount := h.countFilesAndDirs(sharePath)
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
func (h *HTTPServer) countFilesAndDirs(path string) (int64, int64) {
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

// validateShareCreateRequest validates a share creation request
func (h *HTTPServer) validateShareCreateRequest(req *ShareCreateRequest) error {
	if req.Name == "" {
		return fmt.Errorf("share name is required")
	}

	// Validate share name (alphanumeric, underscore, hyphen only)
	if !isValidShareName(req.Name) {
		return fmt.Errorf("invalid share name: must contain only letters, numbers, underscore, and hyphen")
	}

	// Check if share already exists
	if _, err := h.getShare(req.Name); err == nil {
		return fmt.Errorf("share '%s' already exists", req.Name)
	}

	// Validate allocator method
	if req.AllocatorMethod != "" {
		validMethods := []string{"high-water", "most-free", "fill-up"}
		if !contains(validMethods, req.AllocatorMethod) {
			return fmt.Errorf("invalid allocator method: must be one of %v", validMethods)
		}
	}

	// Validate cache usage
	if req.UseCache != "" {
		validCache := []string{"yes", "no", "only", "prefer"}
		if !contains(validCache, req.UseCache) {
			return fmt.Errorf("invalid cache usage: must be one of %v", validCache)
		}
	}

	// Validate security settings
	if req.SMBSecurity != "" {
		validSecurity := []string{"public", "secure", "private"}
		if !contains(validSecurity, req.SMBSecurity) {
			return fmt.Errorf("invalid SMB security: must be one of %v", validSecurity)
		}
	}

	if req.NFSSecurity != "" {
		validSecurity := []string{"public", "secure", "private"}
		if !contains(validSecurity, req.NFSSecurity) {
			return fmt.Errorf("invalid NFS security: must be one of %v", validSecurity)
		}
	}

	return nil
}

// validateShareUpdateRequest validates a share update request
func (h *HTTPServer) validateShareUpdateRequest(req *ShareUpdateRequest) error {
	// Similar validation as create, but name is not required
	if req.AllocatorMethod != "" {
		validMethods := []string{"high-water", "most-free", "fill-up"}
		if !contains(validMethods, req.AllocatorMethod) {
			return fmt.Errorf("invalid allocator method: must be one of %v", validMethods)
		}
	}

	if req.UseCache != "" {
		validCache := []string{"yes", "no", "only", "prefer"}
		if !contains(validCache, req.UseCache) {
			return fmt.Errorf("invalid cache usage: must be one of %v", validCache)
		}
	}

	if req.SMBSecurity != "" {
		validSecurity := []string{"public", "secure", "private"}
		if !contains(validSecurity, req.SMBSecurity) {
			return fmt.Errorf("invalid SMB security: must be one of %v", validSecurity)
		}
	}

	if req.NFSSecurity != "" {
		validSecurity := []string{"public", "secure", "private"}
		if !contains(validSecurity, req.NFSSecurity) {
			return fmt.Errorf("invalid NFS security: must be one of %v", validSecurity)
		}
	}

	return nil
}

// createShare creates a new share
func (h *HTTPServer) createShare(req *ShareCreateRequest) error {
	configPath := fmt.Sprintf("/boot/config/shares/%s.cfg", req.Name)

	// Ensure shares directory exists
	sharesDir := "/boot/config/shares"
	if err := os.MkdirAll(sharesDir, 0755); err != nil {
		return fmt.Errorf("failed to create shares directory: %v", err)
	}

	// Create share configuration content
	config := h.buildShareConfig(req)

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
		h.reloadSMBConfig()
	}

	return nil
}

// updateShare updates an existing share
func (h *HTTPServer) updateShare(shareName string, req *ShareUpdateRequest) error {
	configPath := fmt.Sprintf("/boot/config/shares/%s.cfg", shareName)

	// Check if share exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("share '%s' not found", shareName)
	}

	// Read existing configuration
	existingShare, err := h.parseShareConfig(shareName)
	if err != nil {
		return fmt.Errorf("failed to read existing share config: %v", err)
	}

	// Update fields that are provided
	h.updateShareFields(existingShare, req)

	// Build new configuration
	config := h.buildShareConfigFromShare(existingShare)

	// Write updated configuration
	if err := os.WriteFile(configPath, []byte(config), 0644); err != nil {
		return fmt.Errorf("failed to update share config: %v", err)
	}

	// Reload SMB configuration if SMB settings changed
	if req.SMBEnabled || req.SMBSecurity != "" {
		h.reloadSMBConfig()
	}

	return nil
}

// deleteShare deletes a share
func (h *HTTPServer) deleteShare(shareName string) error {
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
	h.reloadSMBConfig()

	return nil
}

// Helper functions

// isValidShareName validates a share name
func isValidShareName(name string) bool {
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
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// buildShareConfig builds configuration content for a new share
func (h *HTTPServer) buildShareConfig(req *ShareCreateRequest) string {
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

// buildShareConfigFromShare builds configuration content from a Share struct
func (h *HTTPServer) buildShareConfigFromShare(share *Share) string {
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
func (h *HTTPServer) updateShareFields(share *Share, req *ShareUpdateRequest) {
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
func (h *HTTPServer) reloadSMBConfig() {
	// Execute command to reload SMB configuration
	// This is typically done by restarting the SMB service or reloading config
	exec.Command("/etc/rc.d/rc.samba", "reload").Run()
}

// corsMiddleware adds CORS headers
func (h *HTTPServer) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// requestIDMiddleware adds request ID tracking for debugging and tracing
func (h *HTTPServer) requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if request already has an ID from client
		requestID := r.Header.Get("X-Request-ID")

		// Generate a new request ID if not provided
		if requestID == "" {
			requestID = h.generateRequestID()
		}

		// Set the request ID in response headers
		w.Header().Set("X-Request-ID", requestID)

		// Add request ID to request context for use in handlers
		ctx := context.WithValue(r.Context(), requestIDKey, requestID)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

// generateRequestID creates a unique request ID for tracing
func (h *HTTPServer) generateRequestID() string {
	// Use timestamp + random component for uniqueness
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("req_%d_%d", timestamp, timestamp%10000)
}

// compressionMiddleware adds gzip compression for large responses
func (h *HTTPServer) compressionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if client accepts gzip encoding
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		// Only compress responses for certain endpoints or large responses
		shouldCompress := h.shouldCompressResponse(r)
		if !shouldCompress {
			next.ServeHTTP(w, r)
			return
		}

		// Set compression headers
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Set("Vary", "Accept-Encoding")

		// Create gzip writer
		gzipWriter := gzip.NewWriter(w)
		defer gzipWriter.Close()

		// Create wrapper that writes to gzip writer
		gzipResponseWriter := &gzipResponseWriter{
			ResponseWriter: w,
			Writer:         gzipWriter,
		}

		next.ServeHTTP(gzipResponseWriter, r)
	})
}

// shouldCompressResponse determines if a response should be compressed
func (h *HTTPServer) shouldCompressResponse(r *http.Request) bool {
	// Compress responses for endpoints that typically return large data
	compressiblePaths := []string{
		"/api/v1/storage/disks",
		"/api/v1/storage/zfs",
		"/api/v1/docker/containers",
		"/api/v1/notifications",
		"/api/v1/openapi.json",
		"/api/v1/system/resources",
	}

	for _, path := range compressiblePaths {
		if strings.HasPrefix(r.URL.Path, path) {
			return true
		}
	}

	return false
}

// gzipResponseWriter wraps http.ResponseWriter to compress responses
type gzipResponseWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

// versioningMiddleware handles API version negotiation
func (h *HTTPServer) versioningMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Parse Accept header for version negotiation
		acceptHeader := r.Header.Get("Accept")
		apiVersion := h.parseAPIVersion(acceptHeader)

		// Set the negotiated version in request context
		ctx := context.WithValue(r.Context(), apiVersionKey, apiVersion)
		r = r.WithContext(ctx)

		// Set version information in response headers
		w.Header().Set("X-API-Version", apiVersion)
		w.Header().Set("X-API-Supported-Versions", "v1")

		next.ServeHTTP(w, r)
	})
}

// parseAPIVersion extracts API version from Accept header
func (h *HTTPServer) parseAPIVersion(acceptHeader string) string {
	// Default to v1 if no specific version requested
	defaultVersion := "v1"

	if acceptHeader == "" {
		return defaultVersion
	}

	// Look for version-specific media types
	// Format: application/vnd.uma.v1+json
	if strings.Contains(acceptHeader, "application/vnd.uma.v1+json") {
		return "v1"
	}

	// Look for version parameter in Accept header
	// Format: application/json; version=v1
	if strings.Contains(acceptHeader, "version=v1") {
		return "v1"
	}

	// Future versions can be added here
	// if strings.Contains(acceptHeader, "application/vnd.uma.v2+json") {
	//     return "v2"
	// }

	return defaultVersion
}

// getAPIVersionFromContext retrieves the API version from request context
func (h *HTTPServer) getAPIVersionFromContext(r *http.Request) string {
	if version, ok := r.Context().Value(apiVersionKey).(string); ok {
		return version
	}
	return "v1" // Default fallback
}

// writeVersionedResponse writes a response with version-specific formatting
func (h *HTTPServer) writeVersionedResponse(w http.ResponseWriter, r *http.Request, status int, data interface{}, pagination *dto.PaginationInfo) {
	version := h.getAPIVersionFromContext(r)

	switch version {
	case "v1":
		// Current v1 format with standardized response structure
		h.writeStandardResponse(w, status, data, pagination)
	default:
		// Future versions can have different response formats
		h.writeStandardResponse(w, status, data, pagination)
	}
}

// loggingMiddleware logs HTTP requests
func (h *HTTPServer) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a response writer wrapper to capture status code
		wrapper := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapper, r)

		duration := time.Since(start)
		requestID := h.getRequestIDFromContext(r)
		if requestID != "" {
			logger.LightGreen("HTTP %s %s %d %v [%s]", r.Method, r.URL.Path, wrapper.statusCode, duration, requestID)
		} else {
			logger.LightGreen("HTTP %s %s %d %v", r.Method, r.URL.Path, wrapper.statusCode, duration)
		}
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

// Notification API handlers

// handleNotifications handles GET /api/v1/notifications and POST /api/v1/notifications
func (h *HTTPServer) handleNotifications(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.handleGetNotifications(w, r)
	case http.MethodPost:
		h.handleCreateNotification(w, r)
	default:
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// handleNotification handles GET/PUT/DELETE /api/v1/notifications/{id}
func (h *HTTPServer) handleNotification(w http.ResponseWriter, r *http.Request) {
	// Extract notification ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/notifications/")
	if path == "" {
		h.writeError(w, http.StatusBadRequest, "Notification ID required")
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.handleGetNotification(w, r, path)
	case http.MethodPut:
		h.handleUpdateNotification(w, r, path)
	case http.MethodDelete:
		h.handleDeleteNotification(w, r, path)
	default:
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// handleNotificationsClear handles POST /api/v1/notifications/clear
func (h *HTTPServer) handleNotificationsClear(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	if err := h.api.notifications.ClearAllNotifications(); err != nil {
		h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to clear notifications: %v", err))
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{"message": "All notifications cleared"})
}

// handleNotificationsStats handles GET /api/v1/notifications/stats
func (h *HTTPServer) handleNotificationsStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	stats, err := h.api.notifications.GetNotificationStats()
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get notification stats: %v", err))
		return
	}

	h.writeJSON(w, http.StatusOK, stats)
}

// handleNotificationsMarkAllRead handles POST /api/v1/notifications/mark-all-read
func (h *HTTPServer) handleNotificationsMarkAllRead(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	if err := h.api.notifications.MarkAllAsRead(); err != nil {
		h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to mark all notifications as read: %v", err))
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{"message": "All notifications marked as read"})
}

// handleGetNotifications handles GET /api/v1/notifications with filtering
func (h *HTTPServer) handleGetNotifications(w http.ResponseWriter, r *http.Request) {
	// Parse pagination parameters
	params := h.parsePaginationParams(r)
	usePagination := r.URL.Query().Get("page") != "" || r.URL.Query().Get("limit") != ""

	// Parse query parameters for filtering
	filter := &notifications.NotificationFilter{}

	if level := r.URL.Query().Get("level"); level != "" {
		filter.Level = notifications.NotificationLevel(level)
	}

	if category := r.URL.Query().Get("category"); category != "" {
		filter.Category = notifications.NotificationCategory(category)
	}

	if readStr := r.URL.Query().Get("read"); readStr != "" {
		if read, err := strconv.ParseBool(readStr); err == nil {
			filter.Read = &read
		}
	}

	if persistentStr := r.URL.Query().Get("persistent"); persistentStr != "" {
		if persistent, err := strconv.ParseBool(persistentStr); err == nil {
			filter.Persistent = &persistent
		}
	}

	// For backward compatibility, still support the old limit parameter
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" && !usePagination {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			filter.Limit = limit
		}
	}

	// Parse time filters
	if sinceStr := r.URL.Query().Get("since"); sinceStr != "" {
		if since, err := time.Parse(time.RFC3339, sinceStr); err == nil {
			filter.Since = &since
		}
	}

	if untilStr := r.URL.Query().Get("until"); untilStr != "" {
		if until, err := time.Parse(time.RFC3339, untilStr); err == nil {
			filter.Until = &until
		}
	}

	notificationsList, err := h.api.notifications.GetNotifications(filter)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get notifications: %v", err))
		return
	}

	if usePagination {
		// Convert notifications to interface{} slice for pagination
		notificationItems := make([]interface{}, len(notificationsList))
		for i, notification := range notificationsList {
			notificationItems[i] = notification
		}

		paginatedNotifications, total := h.paginateSlice(notificationItems, params)
		h.writePaginatedResponse(w, http.StatusOK, paginatedNotifications, total, params)
	} else {
		// Return original format for backward compatibility
		h.writeJSON(w, http.StatusOK, map[string]interface{}{
			"notifications": notificationsList,
			"count":         len(notificationsList),
		})
	}
}

// handleCreateNotification handles POST /api/v1/notifications
func (h *HTTPServer) handleCreateNotification(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title    string                             `json:"title"`
		Message  string                             `json:"message"`
		Level    notifications.NotificationLevel    `json:"level"`
		Category notifications.NotificationCategory `json:"category"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	// Validate required fields
	if req.Title == "" {
		h.writeError(w, http.StatusBadRequest, "Title is required")
		return
	}
	if req.Message == "" {
		h.writeError(w, http.StatusBadRequest, "Message is required")
		return
	}
	if req.Level == "" {
		req.Level = notifications.LevelInfo
	}
	if req.Category == "" {
		req.Category = notifications.CategoryCustom
	}

	notification, err := h.api.notifications.CreateNotification(req.Title, req.Message, req.Level, req.Category)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to create notification: %v", err))
		return
	}

	h.writeJSON(w, http.StatusCreated, notification)
}

// handleGetNotification handles GET /api/v1/notifications/{id}
func (h *HTTPServer) handleGetNotification(w http.ResponseWriter, r *http.Request, id string) {
	notification, err := h.api.notifications.GetNotification(id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.writeError(w, http.StatusNotFound, "Notification not found")
		} else {
			h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get notification: %v", err))
		}
		return
	}

	h.writeJSON(w, http.StatusOK, notification)
}

// handleUpdateNotification handles PUT /api/v1/notifications/{id}
func (h *HTTPServer) handleUpdateNotification(w http.ResponseWriter, r *http.Request, id string) {
	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	notification, err := h.api.notifications.UpdateNotification(id, updates)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.writeError(w, http.StatusNotFound, "Notification not found")
		} else {
			h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to update notification: %v", err))
		}
		return
	}

	h.writeJSON(w, http.StatusOK, notification)
}

// handleDeleteNotification handles DELETE /api/v1/notifications/{id}
func (h *HTTPServer) handleDeleteNotification(w http.ResponseWriter, r *http.Request, id string) {
	if err := h.api.notifications.DeleteNotification(id); err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.writeError(w, http.StatusNotFound, "Notification not found")
		} else {
			h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to delete notification: %v", err))
		}
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{"message": "Notification deleted"})
}

// Command Execution Handlers

// handleExecuteCommand handles POST /api/v1/execute/command
func (h *HTTPServer) handleExecuteCommand(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req command.CommandRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid JSON request")
		return
	}

	// Validate required fields
	if req.Command == "" {
		h.writeError(w, http.StatusBadRequest, "Command is required")
		return
	}

	// Execute command
	response, err := h.commandExecutor.ExecuteCommand(req)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to execute command: %v", err))
		return
	}

	// Return appropriate status code based on success
	statusCode := http.StatusOK
	if !response.Success {
		statusCode = http.StatusBadRequest
	}

	h.writeJSON(w, statusCode, response)
}

// handleExecuteContainer handles POST /api/v1/execute/container
func (h *HTTPServer) handleExecuteContainer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req command.ContainerCommandRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid JSON request")
		return
	}

	// Validate required fields
	if req.ContainerID == "" {
		h.writeError(w, http.StatusBadRequest, "Container ID is required")
		return
	}
	if req.Command == "" {
		h.writeError(w, http.StatusBadRequest, "Command is required")
		return
	}

	// Execute container command
	response, err := h.commandExecutor.ExecuteContainerCommand(req)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to execute container command: %v", err))
		return
	}

	// Return appropriate status code based on success
	statusCode := http.StatusOK
	if !response.Success {
		statusCode = http.StatusBadRequest
	}

	h.writeJSON(w, statusCode, response)
}

// handleAllowedCommands handles GET /api/v1/execute/allowed-commands
func (h *HTTPServer) handleAllowedCommands(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	allowedCommands := h.commandExecutor.GetAllowedCommands()

	response := map[string]interface{}{
		"allowed_commands": allowedCommands,
		"count":            len(allowedCommands),
		"message":          "List of allowed commands for secure execution",
	}

	h.writeJSON(w, http.StatusOK, response)
}
