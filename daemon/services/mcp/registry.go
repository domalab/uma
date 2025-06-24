package mcp

import (
	"fmt"
	"sync"

	"github.com/domalab/uma/daemon/logger"
	"github.com/domalab/uma/daemon/services/api/utils"
)

// ToolRegistry manages the mapping between REST endpoints and MCP tools
type ToolRegistry struct {
	api   utils.APIInterface
	tools map[string]*ToolDefinition
	mutex sync.RWMutex
}

// ToolDefinition represents a tool definition with execution details
type ToolDefinition struct {
	Tool     Tool
	Endpoint string
	Method   string
	Handler  ToolHandler
}

// ToolHandler is a function that executes a tool
type ToolHandler func(args map[string]interface{}) (ToolCallResult, error)

// NewToolRegistry creates a new tool registry
func NewToolRegistry(api utils.APIInterface) *ToolRegistry {
	registry := &ToolRegistry{
		api:   api,
		tools: make(map[string]*ToolDefinition),
	}

	// Register all available tools
	registry.registerTools()

	return registry
}

// registerTools registers all available tools from REST endpoints
func (r *ToolRegistry) registerTools() {
	logger.Info("Registering MCP tools from REST endpoints...")

	// System tools
	r.registerSystemTools()

	// Storage tools
	r.registerStorageTools()

	// Extended storage tools
	r.registerExtendedStorageTools()

	// Docker tools
	r.registerDockerTools()

	// VM tools
	r.registerVMTools()

	// Health tools
	r.registerHealthTools()

	// Monitoring tools
	r.registerMonitoringTools()

	// Network tools
	r.registerNetworkTools()

	// Security tools
	r.registerSecurityTools()

	// Performance tools
	r.registerPerformanceTools()

	// Configuration tools
	r.registerConfigurationTools()

	// Backup tools
	r.registerBackupTools()

	// User management tools
	r.registerUserTools()

	// Log management tools
	r.registerLogTools()

	// Plugin management tools
	r.registerPluginTools()

	// Notification tools
	r.registerNotificationTools()

	logger.Green("Registered %d MCP tools", len(r.tools))
}

// registerSystemTools registers system-related tools
func (r *ToolRegistry) registerSystemTools() {
	// System info tool
	r.registerTool("get_system_info", Tool{
		Name:        "get_system_info",
		Description: "Get comprehensive system information including CPU, memory, and hardware details",
		InputSchema: ToolSchema{
			Type:       "object",
			Properties: map[string]interface{}{},
			Required:   []string{},
		},
	}, "/api/v1/system/info", "GET", r.handleSystemInfo)

	// System stats tool
	r.registerTool("get_system_stats", Tool{
		Name:        "get_system_stats",
		Description: "Get real-time system statistics including CPU usage, memory usage, and load averages",
		InputSchema: ToolSchema{
			Type:       "object",
			Properties: map[string]interface{}{},
			Required:   []string{},
		},
	}, "/api/v1/system/stats", "GET", r.handleSystemStats)

	// System processes tool
	r.registerTool("get_system_processes", Tool{
		Name:        "get_system_processes",
		Description: "Get list of running system processes with CPU and memory usage",
		InputSchema: ToolSchema{
			Type:       "object",
			Properties: map[string]interface{}{},
			Required:   []string{},
		},
	}, "/api/v1/system/processes", "GET", r.handleSystemProcesses)
}

// registerStorageTools registers storage-related tools
func (r *ToolRegistry) registerStorageTools() {
	// Storage overview tool
	r.registerTool("get_storage_overview", Tool{
		Name:        "get_storage_overview",
		Description: "Get storage overview including disk usage, array status, and pool information",
		InputSchema: ToolSchema{
			Type:       "object",
			Properties: map[string]interface{}{},
			Required:   []string{},
		},
	}, "/api/v1/storage/overview", "GET", r.handleStorageOverview)

	// Disk details tool
	r.registerTool("get_disk_details", Tool{
		Name:        "get_disk_details",
		Description: "Get detailed information about a specific disk including SMART data",
		InputSchema: ToolSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"disk_id": map[string]interface{}{
					"type":        "string",
					"description": "The disk identifier (e.g., 'sda', 'nvme0n1')",
				},
			},
			Required: []string{"disk_id"},
		},
	}, "/api/v1/storage/disks/{disk_id}", "GET", r.handleDiskDetails)
}

// registerExtendedStorageTools registers additional storage-related tools
func (r *ToolRegistry) registerExtendedStorageTools() {
	// Get disk SMART data
	r.registerTool("get_disk_smart", Tool{
		Name:        "get_disk_smart",
		Description: "Get SMART health data for a specific disk",
		InputSchema: ToolSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"disk_id": map[string]interface{}{
					"type":        "string",
					"description": "The disk identifier (e.g., 'sda', 'nvme0n1')",
				},
			},
			Required: []string{"disk_id"},
		},
	}, "/api/v1/storage/disks/{disk_id}/smart", "GET", r.handleDiskSMART)

	// Get storage pools
	r.registerTool("get_storage_pools", Tool{
		Name:        "get_storage_pools",
		Description: "Get information about storage pools and cache drives",
		InputSchema: ToolSchema{
			Type:       "object",
			Properties: map[string]interface{}{},
			Required:   []string{},
		},
	}, "/api/v1/storage/pools", "GET", r.handleStoragePools)

	// Get array status
	r.registerTool("get_array_status", Tool{
		Name:        "get_array_status",
		Description: "Get Unraid array status and parity information",
		InputSchema: ToolSchema{
			Type:       "object",
			Properties: map[string]interface{}{},
			Required:   []string{},
		},
	}, "/api/v1/storage/array", "GET", r.handleArrayStatus)

	// Get disk usage
	r.registerTool("get_disk_usage", Tool{
		Name:        "get_disk_usage",
		Description: "Get disk space usage for all mounted filesystems",
		InputSchema: ToolSchema{
			Type:       "object",
			Properties: map[string]interface{}{},
			Required:   []string{},
		},
	}, "/api/v1/storage/usage", "GET", r.handleDiskUsage)

	// Get share details
	r.registerTool("get_share_details", Tool{
		Name:        "get_share_details",
		Description: "Get detailed information about a specific user share",
		InputSchema: ToolSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"share_name": map[string]interface{}{
					"type":        "string",
					"description": "The share name to get details for",
				},
			},
			Required: []string{"share_name"},
		},
	}, "/api/v1/storage/shares/{share_name}", "GET", r.handleShareDetails)

	// Get filesystem info
	r.registerTool("get_filesystem_info", Tool{
		Name:        "get_filesystem_info",
		Description: "Get filesystem information and mount points",
		InputSchema: ToolSchema{
			Type:       "object",
			Properties: map[string]interface{}{},
			Required:   []string{},
		},
	}, "/api/v1/storage/filesystems", "GET", r.handleFilesystemInfo)

	// Get I/O statistics
	r.registerTool("get_io_stats", Tool{
		Name:        "get_io_stats",
		Description: "Get disk I/O statistics for all storage devices",
		InputSchema: ToolSchema{
			Type:       "object",
			Properties: map[string]interface{}{},
			Required:   []string{},
		},
	}, "/api/v1/storage/iostats", "GET", r.handleIOStats)

	// Get temperature data
	r.registerTool("get_disk_temperatures", Tool{
		Name:        "get_disk_temperatures",
		Description: "Get temperature readings for all storage devices",
		InputSchema: ToolSchema{
			Type:       "object",
			Properties: map[string]interface{}{},
			Required:   []string{},
		},
	}, "/api/v1/storage/temperatures", "GET", r.handleDiskTemperatures)

	// Get parity check status
	r.registerTool("get_parity_check_status", Tool{
		Name:        "get_parity_check_status",
		Description: "Get status of parity check operations",
		InputSchema: ToolSchema{
			Type:       "object",
			Properties: map[string]interface{}{},
			Required:   []string{},
		},
	}, "/api/v1/storage/parity", "GET", r.handleParityCheckStatus)

	// Get cache status
	r.registerTool("get_cache_status", Tool{
		Name:        "get_cache_status",
		Description: "Get cache drive status and usage information",
		InputSchema: ToolSchema{
			Type:       "object",
			Properties: map[string]interface{}{},
			Required:   []string{},
		},
	}, "/api/v1/storage/cache", "GET", r.handleCacheStatus)

	// Get ZFS pools (if available)
	r.registerTool("get_zfs_pools", Tool{
		Name:        "get_zfs_pools",
		Description: "Get ZFS pool information and status",
		InputSchema: ToolSchema{
			Type:       "object",
			Properties: map[string]interface{}{},
			Required:   []string{},
		},
	}, "/api/v1/storage/zfs", "GET", r.handleZFSPools)

	// Get RAID status
	r.registerTool("get_raid_status", Tool{
		Name:        "get_raid_status",
		Description: "Get software RAID status and configuration",
		InputSchema: ToolSchema{
			Type:       "object",
			Properties: map[string]interface{}{},
			Required:   []string{},
		},
	}, "/api/v1/storage/raid", "GET", r.handleRAIDStatus)

	// Get disk health summary
	r.registerTool("get_disk_health_summary", Tool{
		Name:        "get_disk_health_summary",
		Description: "Get overall disk health summary for all drives",
		InputSchema: ToolSchema{
			Type:       "object",
			Properties: map[string]interface{}{},
			Required:   []string{},
		},
	}, "/api/v1/storage/health", "GET", r.handleDiskHealthSummary)

	// Get storage alerts
	r.registerTool("get_storage_alerts", Tool{
		Name:        "get_storage_alerts",
		Description: "Get storage-related alerts and warnings",
		InputSchema: ToolSchema{
			Type:       "object",
			Properties: map[string]interface{}{},
			Required:   []string{},
		},
	}, "/api/v1/storage/alerts", "GET", r.handleStorageAlerts)
}

// registerDockerTools registers Docker-related tools
func (r *ToolRegistry) registerDockerTools() {
	// List containers tool
	r.registerTool("list_containers", Tool{
		Name:        "list_containers",
		Description: "List all Docker containers with their status, resource usage, and configuration",
		InputSchema: ToolSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"all": map[string]interface{}{
					"type":        "boolean",
					"description": "Include stopped containers (default: false)",
					"default":     false,
				},
			},
			Required: []string{},
		},
	}, "/api/v1/docker/containers", "GET", r.handleListContainers)

	// Container details tool
	r.registerTool("get_container_details", Tool{
		Name:        "get_container_details",
		Description: "Get detailed information about a specific Docker container",
		InputSchema: ToolSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"container_id": map[string]interface{}{
					"type":        "string",
					"description": "The container ID or name",
				},
			},
			Required: []string{"container_id"},
		},
	}, "/api/v1/docker/containers/{container_id}", "GET", r.handleContainerDetails)
}

// registerVMTools registers VM-related tools
func (r *ToolRegistry) registerVMTools() {
	// List VMs tool
	r.registerTool("list_vms", Tool{
		Name:        "list_vms",
		Description: "List all virtual machines with their status and resource allocation",
		InputSchema: ToolSchema{
			Type:       "object",
			Properties: map[string]interface{}{},
			Required:   []string{},
		},
	}, "/api/v1/vms", "GET", r.handleListVMs)

	// VM details tool
	r.registerTool("get_vm_details", Tool{
		Name:        "get_vm_details",
		Description: "Get detailed information about a specific virtual machine",
		InputSchema: ToolSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"vm_id": map[string]interface{}{
					"type":        "string",
					"description": "The VM identifier or name",
				},
			},
			Required: []string{"vm_id"},
		},
	}, "/api/v1/vms/{vm_id}", "GET", r.handleVMDetails)
}

// registerHealthTools registers health check tools
func (r *ToolRegistry) registerHealthTools() {
	// Health check tool
	r.registerTool("health_check", Tool{
		Name:        "health_check",
		Description: "Perform comprehensive health check of the Unraid system",
		InputSchema: ToolSchema{
			Type:       "object",
			Properties: map[string]interface{}{},
			Required:   []string{},
		},
	}, "/api/v1/health", "GET", r.handleHealthCheck)
}

// registerMonitoringTools registers monitoring tools
func (r *ToolRegistry) registerMonitoringTools() {
	// Get metrics tool
	r.registerTool("get_metrics", Tool{
		Name:        "get_metrics",
		Description: "Get Prometheus metrics for system monitoring",
		InputSchema: ToolSchema{
			Type:       "object",
			Properties: map[string]interface{}{},
			Required:   []string{},
		},
	}, "/metrics", "GET", r.handleMetrics)

	// Get system sensors
	r.registerTool("get_system_sensors", Tool{
		Name:        "get_system_sensors",
		Description: "Get temperature, fan speed, and voltage sensor readings",
		InputSchema: ToolSchema{
			Type:       "object",
			Properties: map[string]interface{}{},
			Required:   []string{},
		},
	}, "/api/v1/system/sensors", "GET", r.handleSystemSensors)

	// Get UPS status
	r.registerTool("get_ups_status", Tool{
		Name:        "get_ups_status",
		Description: "Get UPS (Uninterruptible Power Supply) status and battery information",
		InputSchema: ToolSchema{
			Type:       "object",
			Properties: map[string]interface{}{},
			Required:   []string{},
		},
	}, "/api/v1/system/ups", "GET", r.handleUPSStatus)
}

// registerNetworkTools registers network-related tools
func (r *ToolRegistry) registerNetworkTools() {
	// Get network interfaces
	r.registerTool("get_network_interfaces", Tool{
		Name:        "get_network_interfaces",
		Description: "Get information about all network interfaces",
		InputSchema: ToolSchema{
			Type:       "object",
			Properties: map[string]interface{}{},
			Required:   []string{},
		},
	}, "/api/v1/network/interfaces", "GET", r.handleNetworkInterfaces)

	// Get network statistics
	r.registerTool("get_network_stats", Tool{
		Name:        "get_network_stats",
		Description: "Get network traffic statistics and bandwidth usage",
		InputSchema: ToolSchema{
			Type:       "object",
			Properties: map[string]interface{}{},
			Required:   []string{},
		},
	}, "/api/v1/network/stats", "GET", r.handleNetworkStats)

	// Get network connections
	r.registerTool("get_network_connections", Tool{
		Name:        "get_network_connections",
		Description: "Get active network connections and listening ports",
		InputSchema: ToolSchema{
			Type:       "object",
			Properties: map[string]interface{}{},
			Required:   []string{},
		},
	}, "/api/v1/network/connections", "GET", r.handleNetworkConnections)

	// Test network connectivity
	r.registerTool("test_network_connectivity", Tool{
		Name:        "test_network_connectivity",
		Description: "Test network connectivity to a specific host",
		InputSchema: ToolSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"host": map[string]interface{}{
					"type":        "string",
					"description": "The hostname or IP address to test",
				},
			},
			Required: []string{"host"},
		},
	}, "/api/v1/network/ping", "POST", r.handleNetworkPing)
}

// registerSecurityTools registers security-related tools
func (r *ToolRegistry) registerSecurityTools() {
	// Get security status
	r.registerTool("get_security_status", Tool{
		Name:        "get_security_status",
		Description: "Get overall security status and recommendations",
		InputSchema: ToolSchema{
			Type:       "object",
			Properties: map[string]interface{}{},
			Required:   []string{},
		},
	}, "/api/v1/security/status", "GET", r.handleSecurityStatus)

	// Get firewall status
	r.registerTool("get_firewall_status", Tool{
		Name:        "get_firewall_status",
		Description: "Get firewall configuration and active rules",
		InputSchema: ToolSchema{
			Type:       "object",
			Properties: map[string]interface{}{},
			Required:   []string{},
		},
	}, "/api/v1/security/firewall", "GET", r.handleFirewallStatus)

	// Get SSH status
	r.registerTool("get_ssh_status", Tool{
		Name:        "get_ssh_status",
		Description: "Get SSH service status and configuration",
		InputSchema: ToolSchema{
			Type:       "object",
			Properties: map[string]interface{}{},
			Required:   []string{},
		},
	}, "/api/v1/security/ssh", "GET", r.handleSSHStatus)
}

// registerPerformanceTools registers performance monitoring tools
func (r *ToolRegistry) registerPerformanceTools() {
	// Get CPU performance
	r.registerTool("get_cpu_performance", Tool{
		Name:        "get_cpu_performance",
		Description: "Get detailed CPU performance metrics and usage history",
		InputSchema: ToolSchema{
			Type:       "object",
			Properties: map[string]interface{}{},
			Required:   []string{},
		},
	}, "/api/v1/performance/cpu", "GET", r.handleCPUPerformance)

	// Get memory performance
	r.registerTool("get_memory_performance", Tool{
		Name:        "get_memory_performance",
		Description: "Get detailed memory usage and performance metrics",
		InputSchema: ToolSchema{
			Type:       "object",
			Properties: map[string]interface{}{},
			Required:   []string{},
		},
	}, "/api/v1/performance/memory", "GET", r.handleMemoryPerformance)

	// Get disk I/O performance
	r.registerTool("get_disk_io_performance", Tool{
		Name:        "get_disk_io_performance",
		Description: "Get disk I/O performance metrics and statistics",
		InputSchema: ToolSchema{
			Type:       "object",
			Properties: map[string]interface{}{},
			Required:   []string{},
		},
	}, "/api/v1/performance/disk-io", "GET", r.handleDiskIOPerformance)

	// Get network performance
	r.registerTool("get_network_performance", Tool{
		Name:        "get_network_performance",
		Description: "Get network performance metrics and bandwidth usage",
		InputSchema: ToolSchema{
			Type:       "object",
			Properties: map[string]interface{}{},
			Required:   []string{},
		},
	}, "/api/v1/performance/network", "GET", r.handleNetworkPerformance)
}

// registerConfigurationTools registers configuration management tools
func (r *ToolRegistry) registerConfigurationTools() {
	// Get system configuration
	r.registerTool("get_system_config", Tool{
		Name:        "get_system_config",
		Description: "Get current system configuration settings",
		InputSchema: ToolSchema{
			Type:       "object",
			Properties: map[string]interface{}{},
			Required:   []string{},
		},
	}, "/api/v1/config/system", "GET", r.handleSystemConfig)

	// Get array configuration
	r.registerTool("get_array_config", Tool{
		Name:        "get_array_config",
		Description: "Get Unraid array configuration and disk assignments",
		InputSchema: ToolSchema{
			Type:       "object",
			Properties: map[string]interface{}{},
			Required:   []string{},
		},
	}, "/api/v1/config/array", "GET", r.handleArrayConfig)

	// Get share configuration
	r.registerTool("get_share_config", Tool{
		Name:        "get_share_config",
		Description: "Get user share configuration and settings",
		InputSchema: ToolSchema{
			Type:       "object",
			Properties: map[string]interface{}{},
			Required:   []string{},
		},
	}, "/api/v1/config/shares", "GET", r.handleShareConfig)
}

// registerBackupTools registers backup and restore tools
func (r *ToolRegistry) registerBackupTools() {
	// Get backup status
	r.registerTool("get_backup_status", Tool{
		Name:        "get_backup_status",
		Description: "Get status of backup operations and schedules",
		InputSchema: ToolSchema{
			Type:       "object",
			Properties: map[string]interface{}{},
			Required:   []string{},
		},
	}, "/api/v1/backup/status", "GET", r.handleBackupStatus)

	// List backup jobs
	r.registerTool("list_backup_jobs", Tool{
		Name:        "list_backup_jobs",
		Description: "List all configured backup jobs and their schedules",
		InputSchema: ToolSchema{
			Type:       "object",
			Properties: map[string]interface{}{},
			Required:   []string{},
		},
	}, "/api/v1/backup/jobs", "GET", r.handleBackupJobs)
}

// registerUserTools registers user management tools
func (r *ToolRegistry) registerUserTools() {
	// List users
	r.registerTool("list_users", Tool{
		Name:        "list_users",
		Description: "List all system users and their permissions",
		InputSchema: ToolSchema{
			Type:       "object",
			Properties: map[string]interface{}{},
			Required:   []string{},
		},
	}, "/api/v1/users", "GET", r.handleListUsers)

	// Get user details
	r.registerTool("get_user_details", Tool{
		Name:        "get_user_details",
		Description: "Get detailed information about a specific user",
		InputSchema: ToolSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"username": map[string]interface{}{
					"type":        "string",
					"description": "The username to get details for",
				},
			},
			Required: []string{"username"},
		},
	}, "/api/v1/users/{username}", "GET", r.handleUserDetails)
}

// registerLogTools registers log management tools
func (r *ToolRegistry) registerLogTools() {
	// Get system logs
	r.registerTool("get_system_logs", Tool{
		Name:        "get_system_logs",
		Description: "Get system log entries with optional filtering",
		InputSchema: ToolSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"lines": map[string]interface{}{
					"type":        "integer",
					"description": "Number of log lines to retrieve (default: 100)",
					"default":     100,
				},
				"level": map[string]interface{}{
					"type":        "string",
					"description": "Log level filter (error, warn, info, debug)",
				},
			},
			Required: []string{},
		},
	}, "/api/v1/logs/system", "GET", r.handleSystemLogs)

	// Get application logs
	r.registerTool("get_application_logs", Tool{
		Name:        "get_application_logs",
		Description: "Get application-specific log entries",
		InputSchema: ToolSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"application": map[string]interface{}{
					"type":        "string",
					"description": "Application name to get logs for",
				},
				"lines": map[string]interface{}{
					"type":        "integer",
					"description": "Number of log lines to retrieve (default: 100)",
					"default":     100,
				},
			},
			Required: []string{"application"},
		},
	}, "/api/v1/logs/application", "GET", r.handleApplicationLogs)
}

// registerPluginTools registers plugin management tools
func (r *ToolRegistry) registerPluginTools() {
	// List plugins
	r.registerTool("list_plugins", Tool{
		Name:        "list_plugins",
		Description: "List all installed Unraid plugins",
		InputSchema: ToolSchema{
			Type:       "object",
			Properties: map[string]interface{}{},
			Required:   []string{},
		},
	}, "/api/v1/plugins", "GET", r.handleListPlugins)

	// Get plugin details
	r.registerTool("get_plugin_details", Tool{
		Name:        "get_plugin_details",
		Description: "Get detailed information about a specific plugin",
		InputSchema: ToolSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"plugin_name": map[string]interface{}{
					"type":        "string",
					"description": "The plugin name to get details for",
				},
			},
			Required: []string{"plugin_name"},
		},
	}, "/api/v1/plugins/{plugin_name}", "GET", r.handlePluginDetails)
}

// registerNotificationTools registers notification management tools
func (r *ToolRegistry) registerNotificationTools() {
	// Get notifications
	r.registerTool("get_notifications", Tool{
		Name:        "get_notifications",
		Description: "Get system notifications and alerts",
		InputSchema: ToolSchema{
			Type:       "object",
			Properties: map[string]interface{}{},
			Required:   []string{},
		},
	}, "/api/v1/notifications", "GET", r.handleNotifications)

	// Get notification settings
	r.registerTool("get_notification_settings", Tool{
		Name:        "get_notification_settings",
		Description: "Get notification configuration and settings",
		InputSchema: ToolSchema{
			Type:       "object",
			Properties: map[string]interface{}{},
			Required:   []string{},
		},
	}, "/api/v1/notifications/settings", "GET", r.handleNotificationSettings)
}

// registerTool registers a single tool
func (r *ToolRegistry) registerTool(name string, tool Tool, endpoint string, method string, handler ToolHandler) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.tools[name] = &ToolDefinition{
		Tool:     tool,
		Endpoint: endpoint,
		Method:   method,
		Handler:  handler,
	}

	logger.Debug("Registered MCP tool: %s -> %s %s", name, method, endpoint)
}

// GetTools returns all registered tools
func (r *ToolRegistry) GetTools() ([]Tool, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	tools := make([]Tool, 0, len(r.tools))
	for _, toolDef := range r.tools {
		tools = append(tools, toolDef.Tool)
	}

	return tools, nil
}

// ExecuteTool executes a tool by name with the given arguments
func (r *ToolRegistry) ExecuteTool(name string, args map[string]interface{}) (ToolCallResult, error) {
	r.mutex.RLock()
	toolDef, exists := r.tools[name]
	r.mutex.RUnlock()

	if !exists {
		return ToolCallResult{}, fmt.Errorf("tool not found")
	}

	logger.Info("Executing MCP tool: %s", name)
	return toolDef.Handler(args)
}

// GetRegistryStats returns statistics about the tool registry
func (r *ToolRegistry) GetRegistryStats() map[string]interface{} {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	return map[string]interface{}{
		"total_tools": len(r.tools),
		"tools":       r.getToolNames(),
	}
}

// getToolNames returns a list of all tool names
func (r *ToolRegistry) getToolNames() []string {
	names := make([]string, 0, len(r.tools))
	for name := range r.tools {
		names = append(names, name)
	}
	return names
}

// Tool handler implementations

// handleSystemInfo handles the get_system_info tool
func (r *ToolRegistry) handleSystemInfo(args map[string]interface{}) (ToolCallResult, error) {
	// This would call the actual API endpoint
	// For now, return a placeholder response
	return ToolCallResult{
		Content: []ToolContent{
			{
				Type: "text",
				Text: "System info tool executed successfully. This would return comprehensive system information.",
			},
		},
		IsError: false,
	}, nil
}

// handleSystemStats handles the get_system_stats tool
func (r *ToolRegistry) handleSystemStats(args map[string]interface{}) (ToolCallResult, error) {
	return ToolCallResult{
		Content: []ToolContent{
			{
				Type: "text",
				Text: "System stats tool executed successfully. This would return real-time system statistics.",
			},
		},
		IsError: false,
	}, nil
}

// handleSystemProcesses handles the get_system_processes tool
func (r *ToolRegistry) handleSystemProcesses(args map[string]interface{}) (ToolCallResult, error) {
	return ToolCallResult{
		Content: []ToolContent{
			{
				Type: "text",
				Text: "System processes tool executed successfully. This would return running processes.",
			},
		},
		IsError: false,
	}, nil
}

// handleStorageOverview handles the get_storage_overview tool
func (r *ToolRegistry) handleStorageOverview(args map[string]interface{}) (ToolCallResult, error) {
	return ToolCallResult{
		Content: []ToolContent{
			{
				Type: "text",
				Text: "Storage overview tool executed successfully. This would return storage information.",
			},
		},
		IsError: false,
	}, nil
}

// handleDiskDetails handles the get_disk_details tool
func (r *ToolRegistry) handleDiskDetails(args map[string]interface{}) (ToolCallResult, error) {
	diskID, ok := args["disk_id"].(string)
	if !ok {
		return ToolCallResult{
			Content: []ToolContent{
				{
					Type: "text",
					Text: "Error: disk_id parameter is required and must be a string",
				},
			},
			IsError: true,
		}, nil
	}

	return ToolCallResult{
		Content: []ToolContent{
			{
				Type: "text",
				Text: fmt.Sprintf("Disk details tool executed for disk: %s. This would return detailed disk information.", diskID),
			},
		},
		IsError: false,
	}, nil
}

// handleListContainers handles the list_containers tool
func (r *ToolRegistry) handleListContainers(args map[string]interface{}) (ToolCallResult, error) {
	all := false
	if allVal, ok := args["all"].(bool); ok {
		all = allVal
	}

	text := "Container list tool executed successfully."
	if all {
		text += " Including stopped containers."
	} else {
		text += " Showing only running containers."
	}

	return ToolCallResult{
		Content: []ToolContent{
			{
				Type: "text",
				Text: text,
			},
		},
		IsError: false,
	}, nil
}

// handleContainerDetails handles the get_container_details tool
func (r *ToolRegistry) handleContainerDetails(args map[string]interface{}) (ToolCallResult, error) {
	containerID, ok := args["container_id"].(string)
	if !ok {
		return ToolCallResult{
			Content: []ToolContent{
				{
					Type: "text",
					Text: "Error: container_id parameter is required and must be a string",
				},
			},
			IsError: true,
		}, nil
	}

	return ToolCallResult{
		Content: []ToolContent{
			{
				Type: "text",
				Text: fmt.Sprintf("Container details tool executed for container: %s. This would return detailed container information.", containerID),
			},
		},
		IsError: false,
	}, nil
}

// handleListVMs handles the list_vms tool
func (r *ToolRegistry) handleListVMs(args map[string]interface{}) (ToolCallResult, error) {
	return ToolCallResult{
		Content: []ToolContent{
			{
				Type: "text",
				Text: "VM list tool executed successfully. This would return all virtual machines.",
			},
		},
		IsError: false,
	}, nil
}

// handleVMDetails handles the get_vm_details tool
func (r *ToolRegistry) handleVMDetails(args map[string]interface{}) (ToolCallResult, error) {
	vmID, ok := args["vm_id"].(string)
	if !ok {
		return ToolCallResult{
			Content: []ToolContent{
				{
					Type: "text",
					Text: "Error: vm_id parameter is required and must be a string",
				},
			},
			IsError: true,
		}, nil
	}

	return ToolCallResult{
		Content: []ToolContent{
			{
				Type: "text",
				Text: fmt.Sprintf("VM details tool executed for VM: %s. This would return detailed VM information.", vmID),
			},
		},
		IsError: false,
	}, nil
}

// handleHealthCheck handles the health_check tool
func (r *ToolRegistry) handleHealthCheck(args map[string]interface{}) (ToolCallResult, error) {
	return ToolCallResult{
		Content: []ToolContent{
			{
				Type: "text",
				Text: "Health check tool executed successfully. This would return comprehensive system health status.",
			},
		},
		IsError: false,
	}, nil
}

// handleMetrics handles the get_metrics tool
func (r *ToolRegistry) handleMetrics(args map[string]interface{}) (ToolCallResult, error) {
	return ToolCallResult{
		Content: []ToolContent{
			{
				Type: "text",
				Text: "Metrics tool executed successfully. This would return Prometheus metrics.",
			},
		},
		IsError: false,
	}, nil
}

// Additional handler implementations for new tools

// handleSystemSensors handles the get_system_sensors tool
func (r *ToolRegistry) handleSystemSensors(args map[string]interface{}) (ToolCallResult, error) {
	return ToolCallResult{
		Content: []ToolContent{
			{
				Type: "text",
				Text: "System sensors tool executed successfully. This would return temperature, fan, and voltage readings.",
			},
		},
		IsError: false,
	}, nil
}

// handleUPSStatus handles the get_ups_status tool
func (r *ToolRegistry) handleUPSStatus(args map[string]interface{}) (ToolCallResult, error) {
	return ToolCallResult{
		Content: []ToolContent{
			{
				Type: "text",
				Text: "UPS status tool executed successfully. This would return UPS battery and power information.",
			},
		},
		IsError: false,
	}, nil
}

// handleNetworkInterfaces handles the get_network_interfaces tool
func (r *ToolRegistry) handleNetworkInterfaces(args map[string]interface{}) (ToolCallResult, error) {
	return ToolCallResult{
		Content: []ToolContent{
			{
				Type: "text",
				Text: "Network interfaces tool executed successfully. This would return network interface information.",
			},
		},
		IsError: false,
	}, nil
}

// handleNetworkStats handles the get_network_stats tool
func (r *ToolRegistry) handleNetworkStats(args map[string]interface{}) (ToolCallResult, error) {
	return ToolCallResult{
		Content: []ToolContent{
			{
				Type: "text",
				Text: "Network stats tool executed successfully. This would return network traffic statistics.",
			},
		},
		IsError: false,
	}, nil
}

// handleNetworkConnections handles the get_network_connections tool
func (r *ToolRegistry) handleNetworkConnections(args map[string]interface{}) (ToolCallResult, error) {
	return ToolCallResult{
		Content: []ToolContent{
			{
				Type: "text",
				Text: "Network connections tool executed successfully. This would return active network connections.",
			},
		},
		IsError: false,
	}, nil
}

// handleNetworkPing handles the test_network_connectivity tool
func (r *ToolRegistry) handleNetworkPing(args map[string]interface{}) (ToolCallResult, error) {
	host, ok := args["host"].(string)
	if !ok {
		return ToolCallResult{
			Content: []ToolContent{
				{
					Type: "text",
					Text: "Error: host parameter is required and must be a string",
				},
			},
			IsError: true,
		}, nil
	}

	return ToolCallResult{
		Content: []ToolContent{
			{
				Type: "text",
				Text: fmt.Sprintf("Network ping tool executed for host: %s. This would test connectivity.", host),
			},
		},
		IsError: false,
	}, nil
}

// Security tool handlers

// handleSecurityStatus handles the get_security_status tool
func (r *ToolRegistry) handleSecurityStatus(args map[string]interface{}) (ToolCallResult, error) {
	return ToolCallResult{
		Content: []ToolContent{
			{
				Type: "text",
				Text: "Security status tool executed successfully. This would return overall security status and recommendations.",
			},
		},
		IsError: false,
	}, nil
}

// handleFirewallStatus handles the get_firewall_status tool
func (r *ToolRegistry) handleFirewallStatus(args map[string]interface{}) (ToolCallResult, error) {
	return ToolCallResult{
		Content: []ToolContent{
			{
				Type: "text",
				Text: "Firewall status tool executed successfully. This would return firewall configuration and rules.",
			},
		},
		IsError: false,
	}, nil
}

// handleSSHStatus handles the get_ssh_status tool
func (r *ToolRegistry) handleSSHStatus(args map[string]interface{}) (ToolCallResult, error) {
	return ToolCallResult{
		Content: []ToolContent{
			{
				Type: "text",
				Text: "SSH status tool executed successfully. This would return SSH service status and configuration.",
			},
		},
		IsError: false,
	}, nil
}

// Performance tool handlers

// handleCPUPerformance handles the get_cpu_performance tool
func (r *ToolRegistry) handleCPUPerformance(args map[string]interface{}) (ToolCallResult, error) {
	return ToolCallResult{
		Content: []ToolContent{
			{
				Type: "text",
				Text: "CPU performance tool executed successfully. This would return detailed CPU performance metrics.",
			},
		},
		IsError: false,
	}, nil
}

// handleMemoryPerformance handles the get_memory_performance tool
func (r *ToolRegistry) handleMemoryPerformance(args map[string]interface{}) (ToolCallResult, error) {
	return ToolCallResult{
		Content: []ToolContent{
			{
				Type: "text",
				Text: "Memory performance tool executed successfully. This would return detailed memory usage metrics.",
			},
		},
		IsError: false,
	}, nil
}

// handleDiskIOPerformance handles the get_disk_io_performance tool
func (r *ToolRegistry) handleDiskIOPerformance(args map[string]interface{}) (ToolCallResult, error) {
	return ToolCallResult{
		Content: []ToolContent{
			{
				Type: "text",
				Text: "Disk I/O performance tool executed successfully. This would return disk I/O performance metrics.",
			},
		},
		IsError: false,
	}, nil
}

// handleNetworkPerformance handles the get_network_performance tool
func (r *ToolRegistry) handleNetworkPerformance(args map[string]interface{}) (ToolCallResult, error) {
	return ToolCallResult{
		Content: []ToolContent{
			{
				Type: "text",
				Text: "Network performance tool executed successfully. This would return network performance metrics.",
			},
		},
		IsError: false,
	}, nil
}

// Configuration tool handlers

// handleSystemConfig handles the get_system_config tool
func (r *ToolRegistry) handleSystemConfig(args map[string]interface{}) (ToolCallResult, error) {
	return ToolCallResult{
		Content: []ToolContent{
			{
				Type: "text",
				Text: "System config tool executed successfully. This would return current system configuration settings.",
			},
		},
		IsError: false,
	}, nil
}

// handleArrayConfig handles the get_array_config tool
func (r *ToolRegistry) handleArrayConfig(args map[string]interface{}) (ToolCallResult, error) {
	return ToolCallResult{
		Content: []ToolContent{
			{
				Type: "text",
				Text: "Array config tool executed successfully. This would return Unraid array configuration and disk assignments.",
			},
		},
		IsError: false,
	}, nil
}

// handleShareConfig handles the get_share_config tool
func (r *ToolRegistry) handleShareConfig(args map[string]interface{}) (ToolCallResult, error) {
	return ToolCallResult{
		Content: []ToolContent{
			{
				Type: "text",
				Text: "Share config tool executed successfully. This would return user share configuration and settings.",
			},
		},
		IsError: false,
	}, nil
}

// Backup tool handlers

// handleBackupStatus handles the get_backup_status tool
func (r *ToolRegistry) handleBackupStatus(args map[string]interface{}) (ToolCallResult, error) {
	return ToolCallResult{
		Content: []ToolContent{
			{
				Type: "text",
				Text: "Backup status tool executed successfully. This would return status of backup operations and schedules.",
			},
		},
		IsError: false,
	}, nil
}

// handleBackupJobs handles the list_backup_jobs tool
func (r *ToolRegistry) handleBackupJobs(args map[string]interface{}) (ToolCallResult, error) {
	return ToolCallResult{
		Content: []ToolContent{
			{
				Type: "text",
				Text: "Backup jobs tool executed successfully. This would return all configured backup jobs and schedules.",
			},
		},
		IsError: false,
	}, nil
}

// User management tool handlers

// handleListUsers handles the list_users tool
func (r *ToolRegistry) handleListUsers(args map[string]interface{}) (ToolCallResult, error) {
	return ToolCallResult{
		Content: []ToolContent{
			{
				Type: "text",
				Text: "List users tool executed successfully. This would return all system users and their permissions.",
			},
		},
		IsError: false,
	}, nil
}

// handleUserDetails handles the get_user_details tool
func (r *ToolRegistry) handleUserDetails(args map[string]interface{}) (ToolCallResult, error) {
	username, ok := args["username"].(string)
	if !ok {
		return ToolCallResult{
			Content: []ToolContent{
				{
					Type: "text",
					Text: "Error: username parameter is required and must be a string",
				},
			},
			IsError: true,
		}, nil
	}

	return ToolCallResult{
		Content: []ToolContent{
			{
				Type: "text",
				Text: fmt.Sprintf("User details tool executed for user: %s. This would return detailed user information.", username),
			},
		},
		IsError: false,
	}, nil
}

// Log management tool handlers

// handleSystemLogs handles the get_system_logs tool
func (r *ToolRegistry) handleSystemLogs(args map[string]interface{}) (ToolCallResult, error) {
	lines := 100
	if linesVal, ok := args["lines"].(float64); ok {
		lines = int(linesVal)
	}

	level := ""
	if levelVal, ok := args["level"].(string); ok {
		level = levelVal
	}

	text := fmt.Sprintf("System logs tool executed successfully. Retrieving %d lines", lines)
	if level != "" {
		text += fmt.Sprintf(" with level filter: %s", level)
	}
	text += ". This would return system log entries."

	return ToolCallResult{
		Content: []ToolContent{
			{
				Type: "text",
				Text: text,
			},
		},
		IsError: false,
	}, nil
}

// handleApplicationLogs handles the get_application_logs tool
func (r *ToolRegistry) handleApplicationLogs(args map[string]interface{}) (ToolCallResult, error) {
	application, ok := args["application"].(string)
	if !ok {
		return ToolCallResult{
			Content: []ToolContent{
				{
					Type: "text",
					Text: "Error: application parameter is required and must be a string",
				},
			},
			IsError: true,
		}, nil
	}

	lines := 100
	if linesVal, ok := args["lines"].(float64); ok {
		lines = int(linesVal)
	}

	return ToolCallResult{
		Content: []ToolContent{
			{
				Type: "text",
				Text: fmt.Sprintf("Application logs tool executed for %s. Retrieving %d lines. This would return application-specific log entries.", application, lines),
			},
		},
		IsError: false,
	}, nil
}

// Plugin management tool handlers

// handleListPlugins handles the list_plugins tool
func (r *ToolRegistry) handleListPlugins(args map[string]interface{}) (ToolCallResult, error) {
	return ToolCallResult{
		Content: []ToolContent{
			{
				Type: "text",
				Text: "List plugins tool executed successfully. This would return all installed Unraid plugins.",
			},
		},
		IsError: false,
	}, nil
}

// handlePluginDetails handles the get_plugin_details tool
func (r *ToolRegistry) handlePluginDetails(args map[string]interface{}) (ToolCallResult, error) {
	pluginName, ok := args["plugin_name"].(string)
	if !ok {
		return ToolCallResult{
			Content: []ToolContent{
				{
					Type: "text",
					Text: "Error: plugin_name parameter is required and must be a string",
				},
			},
			IsError: true,
		}, nil
	}

	return ToolCallResult{
		Content: []ToolContent{
			{
				Type: "text",
				Text: fmt.Sprintf("Plugin details tool executed for plugin: %s. This would return detailed plugin information.", pluginName),
			},
		},
		IsError: false,
	}, nil
}

// Notification tool handlers

// handleNotifications handles the get_notifications tool
func (r *ToolRegistry) handleNotifications(args map[string]interface{}) (ToolCallResult, error) {
	return ToolCallResult{
		Content: []ToolContent{
			{
				Type: "text",
				Text: "Notifications tool executed successfully. This would return system notifications and alerts.",
			},
		},
		IsError: false,
	}, nil
}

// handleNotificationSettings handles the get_notification_settings tool
func (r *ToolRegistry) handleNotificationSettings(args map[string]interface{}) (ToolCallResult, error) {
	return ToolCallResult{
		Content: []ToolContent{
			{
				Type: "text",
				Text: "Notification settings tool executed successfully. This would return notification configuration and settings.",
			},
		},
		IsError: false,
	}, nil
}

// Extended storage tool handlers

// handleDiskSMART handles the get_disk_smart tool
func (r *ToolRegistry) handleDiskSMART(args map[string]interface{}) (ToolCallResult, error) {
	diskID, ok := args["disk_id"].(string)
	if !ok {
		return ToolCallResult{
			Content: []ToolContent{
				{
					Type: "text",
					Text: "Error: disk_id parameter is required and must be a string",
				},
			},
			IsError: true,
		}, nil
	}

	return ToolCallResult{
		Content: []ToolContent{
			{
				Type: "text",
				Text: fmt.Sprintf("Disk SMART tool executed for disk: %s. This would return SMART health data.", diskID),
			},
		},
		IsError: false,
	}, nil
}

// handleStoragePools handles the get_storage_pools tool
func (r *ToolRegistry) handleStoragePools(args map[string]interface{}) (ToolCallResult, error) {
	return ToolCallResult{
		Content: []ToolContent{
			{
				Type: "text",
				Text: "Storage pools tool executed successfully. This would return storage pool information.",
			},
		},
		IsError: false,
	}, nil
}

// handleArrayStatus handles the get_array_status tool
func (r *ToolRegistry) handleArrayStatus(args map[string]interface{}) (ToolCallResult, error) {
	return ToolCallResult{
		Content: []ToolContent{
			{
				Type: "text",
				Text: "Array status tool executed successfully. This would return Unraid array status and parity information.",
			},
		},
		IsError: false,
	}, nil
}

// handleDiskUsage handles the get_disk_usage tool
func (r *ToolRegistry) handleDiskUsage(args map[string]interface{}) (ToolCallResult, error) {
	return ToolCallResult{
		Content: []ToolContent{
			{
				Type: "text",
				Text: "Disk usage tool executed successfully. This would return disk space usage for all filesystems.",
			},
		},
		IsError: false,
	}, nil
}

// handleShareDetails handles the get_share_details tool
func (r *ToolRegistry) handleShareDetails(args map[string]interface{}) (ToolCallResult, error) {
	shareName, ok := args["share_name"].(string)
	if !ok {
		return ToolCallResult{
			Content: []ToolContent{
				{
					Type: "text",
					Text: "Error: share_name parameter is required and must be a string",
				},
			},
			IsError: true,
		}, nil
	}

	return ToolCallResult{
		Content: []ToolContent{
			{
				Type: "text",
				Text: fmt.Sprintf("Share details tool executed for share: %s. This would return detailed share information.", shareName),
			},
		},
		IsError: false,
	}, nil
}

// handleFilesystemInfo handles the get_filesystem_info tool
func (r *ToolRegistry) handleFilesystemInfo(args map[string]interface{}) (ToolCallResult, error) {
	return ToolCallResult{
		Content: []ToolContent{
			{
				Type: "text",
				Text: "Filesystem info tool executed successfully. This would return filesystem information and mount points.",
			},
		},
		IsError: false,
	}, nil
}

// handleIOStats handles the get_io_stats tool
func (r *ToolRegistry) handleIOStats(args map[string]interface{}) (ToolCallResult, error) {
	return ToolCallResult{
		Content: []ToolContent{
			{
				Type: "text",
				Text: "I/O stats tool executed successfully. This would return disk I/O statistics for all devices.",
			},
		},
		IsError: false,
	}, nil
}

// handleDiskTemperatures handles the get_disk_temperatures tool
func (r *ToolRegistry) handleDiskTemperatures(args map[string]interface{}) (ToolCallResult, error) {
	return ToolCallResult{
		Content: []ToolContent{
			{
				Type: "text",
				Text: "Disk temperatures tool executed successfully. This would return temperature readings for all storage devices.",
			},
		},
		IsError: false,
	}, nil
}

// handleParityCheckStatus handles the get_parity_check_status tool
func (r *ToolRegistry) handleParityCheckStatus(args map[string]interface{}) (ToolCallResult, error) {
	return ToolCallResult{
		Content: []ToolContent{
			{
				Type: "text",
				Text: "Parity check status tool executed successfully. This would return parity check operation status.",
			},
		},
		IsError: false,
	}, nil
}

// handleCacheStatus handles the get_cache_status tool
func (r *ToolRegistry) handleCacheStatus(args map[string]interface{}) (ToolCallResult, error) {
	return ToolCallResult{
		Content: []ToolContent{
			{
				Type: "text",
				Text: "Cache status tool executed successfully. This would return cache drive status and usage.",
			},
		},
		IsError: false,
	}, nil
}

// handleZFSPools handles the get_zfs_pools tool
func (r *ToolRegistry) handleZFSPools(args map[string]interface{}) (ToolCallResult, error) {
	return ToolCallResult{
		Content: []ToolContent{
			{
				Type: "text",
				Text: "ZFS pools tool executed successfully. This would return ZFS pool information and status.",
			},
		},
		IsError: false,
	}, nil
}

// handleRAIDStatus handles the get_raid_status tool
func (r *ToolRegistry) handleRAIDStatus(args map[string]interface{}) (ToolCallResult, error) {
	return ToolCallResult{
		Content: []ToolContent{
			{
				Type: "text",
				Text: "RAID status tool executed successfully. This would return software RAID status and configuration.",
			},
		},
		IsError: false,
	}, nil
}

// handleDiskHealthSummary handles the get_disk_health_summary tool
func (r *ToolRegistry) handleDiskHealthSummary(args map[string]interface{}) (ToolCallResult, error) {
	return ToolCallResult{
		Content: []ToolContent{
			{
				Type: "text",
				Text: "Disk health summary tool executed successfully. This would return overall disk health for all drives.",
			},
		},
		IsError: false,
	}, nil
}

// handleStorageAlerts handles the get_storage_alerts tool
func (r *ToolRegistry) handleStorageAlerts(args map[string]interface{}) (ToolCallResult, error) {
	return ToolCallResult{
		Content: []ToolContent{
			{
				Type: "text",
				Text: "Storage alerts tool executed successfully. This would return storage-related alerts and warnings.",
			},
		},
		IsError: false,
	}, nil
}
