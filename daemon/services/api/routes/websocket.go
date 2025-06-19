package routes

// registerWebSocketRoutes registers all WebSocket-related endpoints
func (r *Router) registerWebSocketRoutes() {
	// WebSocket endpoints
	r.mux.HandleFunc("/api/v1/ws/system/stats", r.websocketHandler.HandleSystemStatsWebSocket)
	r.mux.HandleFunc("/api/v1/ws/docker/events", r.websocketHandler.HandleDockerEventsWebSocket)
	r.mux.HandleFunc("/api/v1/ws/storage/status", r.websocketHandler.HandleStorageStatusWebSocket)
}
