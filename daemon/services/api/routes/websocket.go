package routes

// registerWebSocketRoutes registers all WebSocket-related endpoints
func (r *Router) registerWebSocketRoutes() {
	// Unified WebSocket endpoint with subscription management
	// Note: WebSocket endpoints need to handle GET requests for upgrade, so we don't restrict the method
	r.mux.HandleFunc("/api/v1/ws", r.webSocketHandler.HandleWebSocket)

	// MCP WebSocket endpoint for JSON-RPC 2.0 protocol
	// Note: WebSocket endpoints need to handle GET requests for upgrade, so we don't restrict the method
	r.mux.HandleFunc("/api/v1/mcp", r.mcpHandler.HandleMCPWebSocket)
}
