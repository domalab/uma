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

	// Array control endpoints with enhanced orchestration
	r.mux.HandleFunc("/api/v1/storage/array/start", r.storageHandler.HandleArrayStart)
	r.mux.HandleFunc("/api/v1/storage/array/stop", r.storageHandler.HandleArrayStop)

	// Parity endpoints moved to system routes to avoid conflicts
}
