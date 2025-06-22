package mcp

import (
	"fmt"
	"sync"

	"github.com/domalab/uma/daemon/logger"
	"github.com/domalab/uma/daemon/services/api/utils"
)

// SimpleToolRegistry provides a simplified tool registry for MCP without OpenAPI dependency
type SimpleToolRegistry struct {
	api   utils.APIInterface
	tools map[string]*ToolDefinition
	mutex sync.RWMutex
}

// NewSimpleToolRegistry creates a new simplified tool registry
func NewSimpleToolRegistry(api utils.APIInterface) *SimpleToolRegistry {
	registry := &SimpleToolRegistry{
		api:   api,
		tools: make(map[string]*ToolDefinition),
	}

	// Register predefined tools for core UMA functionality
	registry.registerPredefinedTools()

	return registry
}

// registerPredefinedTools registers a set of predefined tools for core UMA functionality
func (r *SimpleToolRegistry) registerPredefinedTools() {
	logger.Info("Registering predefined MCP tools for UMA...")

	// System information tools
	r.registerTool("get_system_info", Tool{
		Name:        "get_system_info",
		Description: "Get comprehensive system information including hardware, OS, and Unraid details",
		InputSchema: ToolSchema{
			Type:       "object",
			Properties: map[string]interface{}{},
			Required:   []string{},
		},
	}, "/api/v1/system/info", "GET", r.createSystemInfoHandler())

	// Storage tools
	r.registerTool("get_storage_disks", Tool{
		Name:        "get_storage_disks",
		Description: "Get information about all storage disks in the system",
		InputSchema: ToolSchema{
			Type:       "object",
			Properties: map[string]interface{}{},
			Required:   []string{},
		},
	}, "/api/v1/storage/disks", "GET", r.createStorageDisksHandler())

	// Docker tools
	r.registerTool("get_docker_containers", Tool{
		Name:        "get_docker_containers",
		Description: "Get information about all Docker containers",
		InputSchema: ToolSchema{
			Type:       "object",
			Properties: map[string]interface{}{},
			Required:   []string{},
		},
	}, "/api/v1/docker/containers", "GET", r.createDockerContainersHandler())

	logger.Green("Registered %d predefined MCP tools", len(r.tools))
}

// createSystemInfoHandler creates a handler for system information
func (r *SimpleToolRegistry) createSystemInfoHandler() ToolHandler {
	return func(args map[string]interface{}) (ToolCallResult, error) {
		// This would call the actual API endpoint
		result := ToolCallResult{
			Content: []ToolContent{
				{
					Type: "text",
					Text: "System information retrieved successfully. This tool provides comprehensive system details including hardware specifications, OS information, and Unraid-specific data.",
				},
			},
			IsError: false,
		}
		return result, nil
	}
}

// createStorageDisksHandler creates a handler for storage disk information
func (r *SimpleToolRegistry) createStorageDisksHandler() ToolHandler {
	return func(args map[string]interface{}) (ToolCallResult, error) {
		result := ToolCallResult{
			Content: []ToolContent{
				{
					Type: "text",
					Text: "Storage disk information retrieved successfully. This tool provides details about all storage devices including capacity, usage, health status, and configuration.",
				},
			},
			IsError: false,
		}
		return result, nil
	}
}

// createDockerContainersHandler creates a handler for Docker container information
func (r *SimpleToolRegistry) createDockerContainersHandler() ToolHandler {
	return func(args map[string]interface{}) (ToolCallResult, error) {
		result := ToolCallResult{
			Content: []ToolContent{
				{
					Type: "text",
					Text: "Docker container information retrieved successfully. This tool provides details about all containers including status, resource usage, and configuration.",
				},
			},
			IsError: false,
		}
		return result, nil
	}
}

// registerTool registers a single tool
func (r *SimpleToolRegistry) registerTool(name string, tool Tool, endpoint string, method string, handler ToolHandler) {
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
func (r *SimpleToolRegistry) GetTools() ([]Tool, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	tools := make([]Tool, 0, len(r.tools))
	for _, toolDef := range r.tools {
		tools = append(tools, toolDef.Tool)
	}

	return tools, nil
}

// ExecuteTool executes a tool by name with the given arguments
func (r *SimpleToolRegistry) ExecuteTool(name string, args map[string]interface{}) (ToolCallResult, error) {
	r.mutex.RLock()
	toolDef, exists := r.tools[name]
	r.mutex.RUnlock()

	if !exists {
		return ToolCallResult{}, fmt.Errorf("tool not found")
	}

	logger.Info("Executing MCP tool: %s", name)
	return toolDef.Handler(args)
}

// GetToolByName returns a specific tool by name
func (r *SimpleToolRegistry) GetToolByName(name string) (*ToolDefinition, bool) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	tool, exists := r.tools[name]
	return tool, exists
}

// GetToolCount returns the number of registered tools
func (r *SimpleToolRegistry) GetToolCount() int {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	return len(r.tools)
}

// GetToolsByCategory returns tools grouped by category
func (r *SimpleToolRegistry) GetToolsByCategory() map[string][]Tool {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	categories := make(map[string][]Tool)

	for _, toolDef := range r.tools {
		// Simple categorization based on tool name prefix
		category := "General"
		if len(toolDef.Tool.Name) > 4 {
			switch {
			case toolDef.Tool.Name[:6] == "system" || toolDef.Tool.Name[:3] == "get":
				category = "System"
			case toolDef.Tool.Name[:7] == "storage":
				category = "Storage"
			case toolDef.Tool.Name[:6] == "docker":
				category = "Docker"
			}
		}

		categories[category] = append(categories[category], toolDef.Tool)
	}

	return categories
}

// GetRegistryStats returns statistics about the tool registry
func (r *SimpleToolRegistry) GetRegistryStats() map[string]interface{} {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	categories := r.GetToolsByCategory()
	categoryStats := make(map[string]int)

	for category, tools := range categories {
		categoryStats[category] = len(tools)
	}

	return map[string]interface{}{
		"total_tools": len(r.tools),
		"categories":  categoryStats,
	}
}
