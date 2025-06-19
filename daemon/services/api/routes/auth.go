package routes

// registerAuthRoutes registers all authentication-related endpoints
func (r *Router) registerAuthRoutes() {
	// Authentication endpoints
	r.mux.HandleFunc("/api/v1/auth/login", r.authHandler.HandleAuthLogin)
	r.mux.HandleFunc("/api/v1/auth/users", r.authHandler.HandleAuthUsers)
	r.mux.HandleFunc("/api/v1/auth/stats", r.authHandler.HandleAuthStats)
}
