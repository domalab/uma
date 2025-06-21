package routes

import (
	"net/http"

	"github.com/domalab/uma/daemon/services/api/handlers"
	"github.com/domalab/uma/daemon/services/api/utils"
)

// RegisterMCPRoutes registers MCP-related routes
func RegisterMCPRoutes(mux *http.ServeMux, apiAdapter utils.APIInterface) {
	mcpHandler := handlers.NewMCPHandler(apiAdapter)

	// MCP server status and management
	mux.HandleFunc("GET /api/v1/mcp/status", mcpHandler.GetMCPStatus)
	mux.HandleFunc("GET /api/v1/mcp/config", mcpHandler.GetMCPConfig)
	mux.HandleFunc("PUT /api/v1/mcp/config", mcpHandler.UpdateMCPConfig)

	// MCP tools management
	mux.HandleFunc("GET /api/v1/mcp/tools", mcpHandler.GetMCPTools)
	mux.HandleFunc("GET /api/v1/mcp/tools/categories", mcpHandler.GetMCPToolsByCategory)
	mux.HandleFunc("POST /api/v1/mcp/tools/refresh", mcpHandler.RefreshMCPTools)
}
