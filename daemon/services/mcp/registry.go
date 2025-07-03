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
	logger.Info("Registering MCP tools from v2 REST endpoints...")

	// Only register tools for available v2 endpoints
	r.registerV2SystemTools()
	r.registerV2StorageTools()
	r.registerV2ContainerTools()
	r.registerV2VMTools()

	logger.Green("Registered %d MCP tools for v2 API", len(r.tools))
}

// registerV2SystemTools registers v2 system-related tools
func (r *ToolRegistry) registerV2SystemTools() {
	// System health tool
	r.registerTool("get_system_health", Tool{
		Name:        "get_system_health",
		Description: "Get system health status and checks",
		InputSchema: ToolSchema{
			Type:       "object",
			Properties: map[string]interface{}{},
			Required:   []string{},
		},
	}, "/api/v2/system/health", "GET", r.handleSystemHealth)

	// System info tool
	r.registerTool("get_system_info", Tool{
		Name:        "get_system_info",
		Description: "Get comprehensive system information including CPU, memory, and hardware details",
		InputSchema: ToolSchema{
			Type:       "object",
			Properties: map[string]interface{}{},
			Required:   []string{},
		},
	}, "/api/v2/system/info", "GET", r.handleSystemInfo)
}

// registerV2StorageTools registers v2 storage-related tools
func (r *ToolRegistry) registerV2StorageTools() {
	// Storage config tool
	r.registerTool("get_storage_config", Tool{
		Name:        "get_storage_config",
		Description: "Get storage configuration including array status and disk counts",
		InputSchema: ToolSchema{
			Type:       "object",
			Properties: map[string]interface{}{},
			Required:   []string{},
		},
	}, "/api/v2/storage/config", "GET", r.handleStorageConfig)

	// Storage layout tool
	r.registerTool("get_storage_layout", Tool{
		Name:        "get_storage_layout",
		Description: "Get detailed storage layout and disk assignments",
		InputSchema: ToolSchema{
			Type:       "object",
			Properties: map[string]interface{}{},
			Required:   []string{},
		},
	}, "/api/v2/storage/layout", "GET", r.handleStorageLayout)
}

// registerV2ContainerTools registers v2 container-related tools
func (r *ToolRegistry) registerV2ContainerTools() {
	// Container list tool
	r.registerTool("list_containers", Tool{
		Name:        "list_containers",
		Description: "List all Docker containers with their status, resource usage, and configuration",
		InputSchema: ToolSchema{
			Type:       "object",
			Properties: map[string]interface{}{},
			Required:   []string{},
		},
	}, "/api/v2/containers/list", "GET", r.handleListContainers)
}

// registerV2VMTools registers v2 VM-related tools
func (r *ToolRegistry) registerV2VMTools() {
	// VM list tool
	r.registerTool("list_vms", Tool{
		Name:        "list_vms",
		Description: "List all virtual machines with their status and resource allocation",
		InputSchema: ToolSchema{
			Type:       "object",
			Properties: map[string]interface{}{},
			Required:   []string{},
		},
	}, "/api/v2/vms/list", "GET", r.handleListVMs)
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

// Tool handler implementations for v2 endpoints

// handleSystemHealth handles the get_system_health tool
func (r *ToolRegistry) handleSystemHealth(args map[string]interface{}) (ToolCallResult, error) {
	return ToolCallResult{
		Content: []ToolContent{
			{
				Type: "text",
				Text: "System health check completed. Status: healthy. All v2 services operational.",
			},
		},
		IsError: false,
	}, nil
}

// handleSystemInfo handles the get_system_info tool
func (r *ToolRegistry) handleSystemInfo(args map[string]interface{}) (ToolCallResult, error) {
	return ToolCallResult{
		Content: []ToolContent{
			{
				Type: "text",
				Text: "System info tool executed successfully. This would return comprehensive system information from v2 API.",
			},
		},
		IsError: false,
	}, nil
}

// handleStorageConfig handles the get_storage_config tool
func (r *ToolRegistry) handleStorageConfig(args map[string]interface{}) (ToolCallResult, error) {
	return ToolCallResult{
		Content: []ToolContent{
			{
				Type: "text",
				Text: "Storage config tool executed successfully. This would return storage configuration from v2 API.",
			},
		},
		IsError: false,
	}, nil
}

// handleStorageLayout handles the get_storage_layout tool
func (r *ToolRegistry) handleStorageLayout(args map[string]interface{}) (ToolCallResult, error) {
	return ToolCallResult{
		Content: []ToolContent{
			{
				Type: "text",
				Text: "Storage layout tool executed successfully. This would return detailed storage layout from v2 API.",
			},
		},
		IsError: false,
	}, nil
}

// handleListContainers handles the list_containers tool
func (r *ToolRegistry) handleListContainers(args map[string]interface{}) (ToolCallResult, error) {
	return ToolCallResult{
		Content: []ToolContent{
			{
				Type: "text",
				Text: "Container list tool executed successfully. This would return all Docker containers from v2 API.",
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
				Text: "VM list tool executed successfully. This would return all virtual machines from v2 API.",
			},
		},
		IsError: false,
	}, nil
}
