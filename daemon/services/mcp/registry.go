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

	// Docker tools
	r.registerDockerTools()

	// VM tools
	r.registerVMTools()

	// Health tools
	r.registerHealthTools()

	// Monitoring tools
	r.registerMonitoringTools()

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
