package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/domalab/uma/daemon/logger"
	"github.com/domalab/uma/daemon/services/api/utils"
	"github.com/domalab/uma/daemon/services/mcp"
	"github.com/gorilla/websocket"
)

// MCPHandler handles MCP-related API endpoints and WebSocket connections
type MCPHandler struct {
	api         utils.APIInterface
	registry    *mcp.ToolRegistry
	connections map[string]*MCPConnection
	mutex       sync.RWMutex
	upgrader    websocket.Upgrader
}

// MCPConnection represents an active MCP WebSocket connection
type MCPConnection struct {
	id     string
	conn   *websocket.Conn
	ctx    context.Context
	cancel context.CancelFunc
	mutex  sync.RWMutex
}

// NewMCPHandler creates a new MCP handler
func NewMCPHandler(api utils.APIInterface) *MCPHandler {
	return &MCPHandler{
		api:         api,
		registry:    mcp.NewToolRegistry(api),
		connections: make(map[string]*MCPConnection),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins for internal network usage
			},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
	}
}

// GetMCPStatus returns the status of the MCP server
func (h *MCPHandler) GetMCPStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// MCP is now integrated into the HTTP server, so it's always available when the handler exists
	h.mutex.RLock()
	activeConnections := len(h.connections)
	h.mutex.RUnlock()

	// Get total tools from registry
	tools, err := h.registry.GetTools()
	totalTools := 0
	if err == nil {
		totalTools = len(tools)
	}

	status := map[string]interface{}{
		"enabled":            true,
		"status":             "running",
		"max_connections":    100, // Default value
		"active_connections": activeConnections,
		"total_tools":        totalTools,
		"message":            "MCP server is integrated and available on HTTP port",
	}

	utils.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    status,
	})
}

// GetMCPTools returns the list of available MCP tools
func (h *MCPHandler) GetMCPTools(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Get tools from the integrated registry
	tools, err := h.registry.GetTools()
	if err != nil {
		logger.Yellow("Failed to get MCP tools: %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   "Failed to retrieve tools",
		})
		return
	}

	// Convert tools to response format
	toolsList := make([]interface{}, 0, len(tools))
	for _, tool := range tools {
		toolsList = append(toolsList, map[string]interface{}{
			"name":        tool.Name,
			"description": tool.Description,
			"inputSchema": tool.InputSchema,
		})
	}

	utils.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"tools": toolsList,
			"count": len(toolsList),
		},
	})
}

// GetMCPToolsByCategory returns MCP tools grouped by category
func (h *MCPHandler) GetMCPToolsByCategory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	utils.WriteJSON(w, http.StatusServiceUnavailable, map[string]interface{}{
		"success": false,
		"error":   "MCP server is not enabled or not available",
	})
}

// RefreshMCPTools refreshes the MCP tool registry
func (h *MCPHandler) RefreshMCPTools(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	utils.WriteJSON(w, http.StatusServiceUnavailable, map[string]interface{}{
		"success": false,
		"error":   "MCP server is not enabled or not available",
	})
}

// GetMCPConfig returns the current MCP configuration
func (h *MCPHandler) GetMCPConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Return integrated MCP configuration
	config := map[string]interface{}{
		"enabled":         true, // MCP is now integrated and always enabled
		"max_connections": 100,
	}

	utils.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    config,
	})
}

// UpdateMCPConfig updates the MCP configuration
func (h *MCPHandler) UpdateMCPConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Parse request body
	var updateRequest struct {
		Enabled        *bool `json:"enabled,omitempty"`
		MaxConnections *int  `json:"max_connections,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&updateRequest); err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   "Invalid request body",
		})
		return
	}

	// Validate max connections
	if updateRequest.MaxConnections != nil && *updateRequest.MaxConnections <= 0 {
		utils.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   "Max connections must be greater than 0",
		})
		return
	}

	// Get the configuration manager from the API adapter
	configManager, ok := h.api.GetConfigManager().(interface {
		SetMCPEnabled(bool) error
		SetMCPMaxConnections(int) error
	})

	if !ok {
		logger.Yellow("Configuration manager not available or doesn't support MCP configuration")
		utils.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   "Configuration management not available",
		})
		return
	}

	// Apply configuration changes
	if updateRequest.Enabled != nil {
		if err := configManager.SetMCPEnabled(*updateRequest.Enabled); err != nil {
			logger.Yellow("Failed to set MCP enabled: %v", err)
			utils.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
				"success": false,
				"error":   fmt.Sprintf("Failed to update MCP enabled setting: %v", err),
			})
			return
		}
	}

	if updateRequest.MaxConnections != nil {
		if err := configManager.SetMCPMaxConnections(*updateRequest.MaxConnections); err != nil {
			logger.Yellow("Failed to set MCP max connections: %v", err)
			utils.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
				"success": false,
				"error":   fmt.Sprintf("Failed to update MCP max connections: %v", err),
			})
			return
		}
	}

	logger.Green("MCP configuration updated successfully")
	utils.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"message": "MCP configuration updated successfully. Restart required for changes to take effect.",
		},
	})
}

// HandleMCPWebSocket handles MCP WebSocket connections for JSON-RPC 2.0 protocol
func (h *MCPHandler) HandleMCPWebSocket(w http.ResponseWriter, r *http.Request) {
	logger.Blue("MCP WebSocket connection attempt from %s", r.RemoteAddr)

	// Check connection limit
	h.mutex.RLock()
	connectionCount := len(h.connections)
	h.mutex.RUnlock()

	maxConnections := 100 // Default max connections
	if connectionCount >= maxConnections {
		logger.Yellow("MCP WebSocket connection rejected: maximum connections (%d) reached", maxConnections)
		http.Error(w, "Maximum connections reached", http.StatusServiceUnavailable)
		return
	}

	// Upgrade connection to WebSocket
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Red("MCP WebSocket upgrade failed: %v", err)
		return
	}

	// Create connection context
	ctx, cancel := context.WithCancel(context.Background())

	// Generate connection ID
	connectionID := fmt.Sprintf("mcp_%d", time.Now().UnixNano())

	// Create MCP connection object
	mcpConn := &MCPConnection{
		id:     connectionID,
		conn:   conn,
		ctx:    ctx,
		cancel: cancel,
	}

	// Store connection
	h.mutex.Lock()
	h.connections[connectionID] = mcpConn
	h.mutex.Unlock()

	logger.Green("MCP WebSocket connection established: %s", connectionID)

	// Handle connection cleanup
	defer func() {
		h.mutex.Lock()
		delete(h.connections, connectionID)
		h.mutex.Unlock()
		cancel()
		conn.Close()
		logger.Blue("MCP WebSocket connection closed: %s", connectionID)
	}()

	// Handle the MCP connection
	h.handleMCPConnection(mcpConn)
}

// handleMCPConnection handles the JSON-RPC 2.0 protocol for an MCP connection
func (h *MCPHandler) handleMCPConnection(mcpConn *MCPConnection) {
	for {
		select {
		case <-mcpConn.ctx.Done():
			return
		default:
			// Set read deadline
			mcpConn.conn.SetReadDeadline(time.Now().Add(60 * time.Second))

			// Read message
			_, messageBytes, err := mcpConn.conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					logger.Red("MCP WebSocket read error: %v", err)
				}
				return
			}

			// Process JSON-RPC 2.0 message
			go h.processMCPMessage(mcpConn, messageBytes)
		}
	}
}

// processMCPMessage processes a JSON-RPC 2.0 message
func (h *MCPHandler) processMCPMessage(mcpConn *MCPConnection, messageBytes []byte) {
	var request map[string]interface{}
	if err := json.Unmarshal(messageBytes, &request); err != nil {
		logger.Red("MCP message parse error: %v", err)
		h.sendMCPError(mcpConn, nil, -32700, "Parse error", nil)
		return
	}

	// Extract JSON-RPC 2.0 fields
	jsonrpc, _ := request["jsonrpc"].(string)
	method, _ := request["method"].(string)
	id := request["id"]

	// Validate JSON-RPC 2.0 format
	if jsonrpc != "2.0" {
		h.sendMCPError(mcpConn, id, -32600, "Invalid Request", nil)
		return
	}

	// Handle different MCP methods
	switch method {
	case "initialize":
		h.handleMCPInitialize(mcpConn, id, request)
	case "tools/list":
		h.handleMCPToolsList(mcpConn, id)
	case "tools/call":
		h.handleMCPToolCall(mcpConn, id, request)
	default:
		h.sendMCPError(mcpConn, id, -32601, "Method not found", nil)
	}
}

// sendMCPError sends a JSON-RPC 2.0 error response
func (h *MCPHandler) sendMCPError(mcpConn *MCPConnection, id interface{}, code int, message string, data interface{}) {
	response := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      id,
		"error": map[string]interface{}{
			"code":    code,
			"message": message,
		},
	}
	if data != nil {
		response["error"].(map[string]interface{})["data"] = data
	}

	h.sendMCPResponse(mcpConn, response)
}

// sendMCPResponse sends a JSON-RPC 2.0 response
func (h *MCPHandler) sendMCPResponse(mcpConn *MCPConnection, response map[string]interface{}) {
	mcpConn.mutex.Lock()
	defer mcpConn.mutex.Unlock()

	if err := mcpConn.conn.WriteJSON(response); err != nil {
		logger.Red("MCP response write error: %v", err)
	}
}

// handleMCPInitialize handles the MCP initialize method
func (h *MCPHandler) handleMCPInitialize(mcpConn *MCPConnection, id interface{}, request map[string]interface{}) {
	response := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      id,
		"result": map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]interface{}{
				"tools": map[string]interface{}{},
			},
			"serverInfo": map[string]interface{}{
				"name":    "UMA MCP Server",
				"version": "1.0.0",
			},
		},
	}
	h.sendMCPResponse(mcpConn, response)
}

// handleMCPToolsList handles the tools/list method
func (h *MCPHandler) handleMCPToolsList(mcpConn *MCPConnection, id interface{}) {
	tools, err := h.registry.GetTools()
	if err != nil {
		h.sendMCPError(mcpConn, id, -32603, "Failed to get tools", err.Error())
		return
	}

	toolsList := make([]interface{}, 0, len(tools))
	for _, tool := range tools {
		toolsList = append(toolsList, map[string]interface{}{
			"name":        tool.Name,
			"description": tool.Description,
			"inputSchema": tool.InputSchema,
		})
	}

	response := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      id,
		"result": map[string]interface{}{
			"tools": toolsList,
		},
	}
	h.sendMCPResponse(mcpConn, response)
}

// handleMCPToolCall handles the tools/call method
func (h *MCPHandler) handleMCPToolCall(mcpConn *MCPConnection, id interface{}, request map[string]interface{}) {
	params, ok := request["params"].(map[string]interface{})
	if !ok {
		h.sendMCPError(mcpConn, id, -32602, "Invalid params", nil)
		return
	}

	toolName, ok := params["name"].(string)
	if !ok {
		h.sendMCPError(mcpConn, id, -32602, "Missing tool name", nil)
		return
	}

	arguments, _ := params["arguments"].(map[string]interface{})
	if arguments == nil {
		arguments = make(map[string]interface{})
	}

	// Execute the tool
	result, err := h.registry.ExecuteTool(toolName, arguments)
	if err != nil {
		h.sendMCPError(mcpConn, id, -32603, "Tool execution failed", err.Error())
		return
	}

	response := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      id,
		"result":  result,
	}
	h.sendMCPResponse(mcpConn, response)
}
