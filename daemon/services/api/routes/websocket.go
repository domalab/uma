package routes

// registerWebSocketRoutes registers all WebSocket-related endpoints
func (r *Router) registerWebSocketRoutes() {
	// Unified WebSocket endpoint with subscription management
	r.mux.HandleFunc("/api/v1/ws", r.enhancedWebSocketHandler.HandleUnifiedWebSocket)
}
