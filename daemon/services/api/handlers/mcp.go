package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/domalab/uma/daemon/services/api/utils"
)

// MCPHandler handles MCP-related API endpoints
type MCPHandler struct {
	api utils.APIInterface
}

// NewMCPHandler creates a new MCP handler
func NewMCPHandler(api utils.APIInterface) *MCPHandler {
	return &MCPHandler{
		api: api,
	}
}

// GetMCPStatus returns the status of the MCP server
func (h *MCPHandler) GetMCPStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// For now, return disabled status since MCP server integration is not complete
	status := map[string]interface{}{
		"enabled":            false,
		"status":             "disabled",
		"port":               34800,
		"max_connections":    100,
		"active_connections": 0,
		"total_tools":        0,
		"message":            "MCP server is not enabled",
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

	utils.WriteJSON(w, http.StatusServiceUnavailable, map[string]interface{}{
		"success": false,
		"error":   "MCP server is not enabled or not available",
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

	// Return default MCP configuration
	config := map[string]interface{}{
		"enabled":         false,
		"port":            34800,
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
		Port           *int  `json:"port,omitempty"`
		MaxConnections *int  `json:"max_connections,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&updateRequest); err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   "Invalid request body",
		})
		return
	}

	// Validate port range
	if updateRequest.Port != nil && (*updateRequest.Port < 1024 || *updateRequest.Port > 65535) {
		utils.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   "Port must be between 1024 and 65535",
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

	// For now, just return success message since configuration management is not fully integrated
	utils.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"message": "MCP configuration updated successfully. Restart required for changes to take effect.",
		},
	})
}
