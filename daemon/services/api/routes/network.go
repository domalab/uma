package routes

// registerNetworkRoutes registers all network-related endpoints
func (r *Router) registerNetworkRoutes() {
	// Network monitoring endpoints
	r.mux.HandleFunc("/api/v1/network/interfaces", r.networkHandler.HandleNetworkInterfaces)
	r.mux.HandleFunc("/api/v1/network/stats", r.networkHandler.HandleNetworkStats)
	r.mux.HandleFunc("/api/v1/network/connections", r.networkHandler.HandleNetworkConnections)
	r.mux.HandleFunc("/api/v1/network/ping", r.networkHandler.HandleNetworkPing)
}
