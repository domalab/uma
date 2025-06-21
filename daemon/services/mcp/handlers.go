package mcp

import (
	"encoding/json"
	"fmt"

	"github.com/domalab/uma/daemon/logger"
)

// handleInitialize handles the MCP initialize method
func (c *Connection) handleInitialize(request *JSONRPCRequest) error {
	logger.Info("MCP initialize request from connection %s", c.id)

	// Parse initialize parameters
	var params InitializeParams
	if request.Params != nil {
		paramsData, err := json.Marshal(request.Params)
		if err != nil {
			return c.sendError(request.ID, InvalidParams, "Invalid parameters", nil)
		}
		if err := json.Unmarshal(paramsData, &params); err != nil {
			return c.sendError(request.ID, InvalidParams, "Invalid parameters", nil)
		}
	}

	// Validate protocol version
	if params.ProtocolVersion != "" && params.ProtocolVersion != MCPProtocolVersion {
		logger.Warn("Client requested protocol version %s, server supports %s",
			params.ProtocolVersion, MCPProtocolVersion)
	}

	// Create initialize result
	result := InitializeResult{
		ProtocolVersion: MCPProtocolVersion,
		Capabilities: ServerCapabilities{
			Tools: &ToolsCapability{
				ListChanged: false, // We don't support dynamic tool changes yet
			},
			Logging: &LoggingCapability{},
		},
		ServerInfo: ServerInfo{
			Name:    "UMA MCP Server",
			Version: "1.0.0",
		},
	}

	logger.Green("MCP connection %s initialized successfully", c.id)
	return c.sendResponse(request.ID, result)
}

// handleToolsList handles the tools/list method
func (c *Connection) handleToolsList(request *JSONRPCRequest) error {
	logger.Info("MCP tools/list request from connection %s", c.id)

	// Get tools from registry
	tools, err := c.server.registry.GetTools()
	if err != nil {
		logger.Yellow("Error getting tools: %v", err)
		return c.sendError(request.ID, InternalError, "Failed to get tools", nil)
	}

	result := ToolsListResult{
		Tools: tools,
	}

	logger.Green("Returning %d tools to connection %s", len(tools), c.id)
	return c.sendResponse(request.ID, result)
}

// handleToolsCall handles the tools/call method
func (c *Connection) handleToolsCall(request *JSONRPCRequest) error {
	logger.Info("MCP tools/call request from connection %s", c.id)

	// Parse tool call parameters
	var params ToolCallParams
	if request.Params == nil {
		return c.sendError(request.ID, InvalidParams, "Missing parameters", nil)
	}

	paramsData, err := json.Marshal(request.Params)
	if err != nil {
		return c.sendError(request.ID, InvalidParams, "Invalid parameters", nil)
	}
	if err := json.Unmarshal(paramsData, &params); err != nil {
		return c.sendError(request.ID, InvalidParams, "Invalid parameters", nil)
	}

	if params.Name == "" {
		return c.sendError(request.ID, InvalidParams, "Tool name is required", nil)
	}

	logger.Info("Executing tool '%s' for connection %s", params.Name, c.id)

	// Execute tool
	result, err := c.server.registry.ExecuteTool(params.Name, params.Arguments)
	if err != nil {
		logger.Yellow("Tool execution error: %v", err)

		// Check if it's a tool not found error
		if err.Error() == "tool not found" {
			return c.sendError(request.ID, ToolNotFound, fmt.Sprintf("Tool '%s' not found", params.Name), nil)
		}

		// Return tool execution error
		return c.sendResponse(request.ID, ToolCallResult{
			Content: []ToolContent{
				{
					Type: "text",
					Text: fmt.Sprintf("Error executing tool: %v", err),
				},
			},
			IsError: true,
		})
	}

	logger.Green("Tool '%s' executed successfully for connection %s", params.Name, c.id)
	return c.sendResponse(request.ID, result)
}

// Additional helper methods for error handling and logging

// validateJSONRPCRequest validates basic JSON-RPC request structure
func (c *Connection) validateJSONRPCRequest(request *JSONRPCRequest) error {
	if request.JSONRPC != "2.0" {
		return fmt.Errorf("invalid JSON-RPC version: %s", request.JSONRPC)
	}

	if request.Method == "" {
		return fmt.Errorf("method is required")
	}

	return nil
}

// GetConnectionStats returns statistics about the connection
func (c *Connection) GetConnectionStats() map[string]interface{} {
	return map[string]interface{}{
		"id":          c.id,
		"remote_addr": c.conn.RemoteAddr().String(),
		"connected":   c.ctx.Err() == nil,
	}
}

// GetServerStats returns statistics about the MCP server
func (s *Server) GetServerStats() map[string]interface{} {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	connectionStats := make([]map[string]interface{}, 0, len(s.connections))
	for _, conn := range s.connections {
		connectionStats = append(connectionStats, conn.GetConnectionStats())
	}

	// Get registry stats
	registryStats := s.registry.GetRegistryStats()

	stats := map[string]interface{}{
		"enabled":            s.config.Enabled,
		"port":               s.config.Port,
		"max_connections":    s.config.MaxConnections,
		"active_connections": len(s.connections),
		"connections":        connectionStats,
	}

	// Merge registry stats
	for key, value := range registryStats {
		stats[key] = value
	}

	return stats
}
