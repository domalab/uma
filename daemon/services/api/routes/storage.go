package routes

// registerStorageRoutes registers all storage-related endpoints
func (r *Router) registerStorageRoutes() {
	// Storage management endpoints
	r.mux.HandleFunc("/api/v1/storage/disks", r.storageHandler.HandleStorageDisks)
	r.mux.HandleFunc("/api/v1/storage/array", r.storageHandler.HandleStorageArray)
	r.mux.HandleFunc("/api/v1/storage/cache", r.storageHandler.HandleStorageCache)
	r.mux.HandleFunc("/api/v1/storage/boot", r.storageHandler.HandleStorageBoot)
	r.mux.HandleFunc("/api/v1/storage/zfs", r.storageHandler.HandleStorageZFS)
	r.mux.HandleFunc("/api/v1/storage/general", r.storageHandler.HandleStorageGeneral)

	// Missing endpoints implementation
	r.mux.HandleFunc("/api/v1/storage/smart", r.storageHandler.HandleStorageSMART)
	r.mux.HandleFunc("/api/v1/storage/array/status", r.storageHandler.HandleArrayStatus)
	r.mux.HandleFunc("/api/v1/storage/shares", r.shareHandler.HandleShares)

	// Array control endpoints with enhanced orchestration
	r.mux.HandleFunc("/api/v1/storage/array/start", r.storageHandler.HandleArrayStart)
	r.mux.HandleFunc("/api/v1/storage/array/stop", r.storageHandler.HandleArrayStop)

	// Storage usage monitoring endpoints
	r.mux.HandleFunc("/api/v1/storage/docker", r.storageHandler.HandleDockerStorage)
	r.mux.HandleFunc("/api/v1/storage/logs", r.storageHandler.HandleLogStorage)

	// Parity endpoints moved to system routes to avoid conflicts
}
