package routes

// registerDockerRoutes registers all Docker-related endpoints
func (r *Router) registerDockerRoutes() {
	// Docker management endpoints
	r.mux.HandleFunc("/api/v1/docker/containers", r.dockerHandler.HandleDockerContainers)
	r.mux.HandleFunc("/api/v1/docker/networks", r.dockerHandler.HandleDockerNetworks)
	r.mux.HandleFunc("/api/v1/docker/images", r.dockerHandler.HandleDockerImages)
	r.mux.HandleFunc("/api/v1/docker/info", r.dockerHandler.HandleDockerInfo)

	// Docker bulk operations
	r.mux.HandleFunc("/api/v1/docker/containers/bulk/start", r.dockerHandler.HandleDockerBulkStart)
	r.mux.HandleFunc("/api/v1/docker/containers/bulk/stop", r.dockerHandler.HandleDockerBulkStop)
	r.mux.HandleFunc("/api/v1/docker/containers/bulk/restart", r.dockerHandler.HandleDockerBulkRestart)

	// Individual Docker container control endpoints
	r.mux.HandleFunc("/api/v1/docker/containers/", r.dockerHandler.HandleDockerContainer) // Handles individual container operations
}
