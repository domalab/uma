package routes

// registerSystemRoutes registers all system-related endpoints
func (r *Router) registerSystemRoutes() {
	// System monitoring endpoints
	r.mux.HandleFunc("/api/v1/system/info", r.systemHandler.HandleSystemInfo)
	r.mux.HandleFunc("/api/v1/system/cpu", r.systemHandler.HandleSystemCPU)
	r.mux.HandleFunc("/api/v1/system/memory", r.systemHandler.HandleSystemMemory)
	r.mux.HandleFunc("/api/v1/system/fans", r.systemHandler.HandleSystemFans)
	r.mux.HandleFunc("/api/v1/system/temperature", r.systemHandler.HandleSystemTemperature)
	r.mux.HandleFunc("/api/v1/system/gpu", r.systemHandler.HandleSystemGPU)
	r.mux.HandleFunc("/api/v1/system/ups", r.systemHandler.HandleSystemUPS)
	r.mux.HandleFunc("/api/v1/system/network", r.systemHandler.HandleSystemNetwork)
	r.mux.HandleFunc("/api/v1/system/resources", r.systemHandler.HandleSystemResources)
	r.mux.HandleFunc("/api/v1/system/filesystems", r.systemHandler.HandleSystemFilesystems)

	// Parity monitoring endpoints
	r.mux.HandleFunc("/api/v1/system/parity/disk", r.systemHandler.HandleParityDisk)
	r.mux.HandleFunc("/api/v1/system/parity/check", r.systemHandler.HandleParityCheck)

	// Legacy GPU endpoint
	r.mux.HandleFunc("/api/v1/gpu", r.systemHandler.HandleGPU)

	// System control endpoints
	r.mux.HandleFunc("/api/v1/system/scripts", r.systemHandler.HandleSystemScripts)
	r.mux.HandleFunc("/api/v1/system/execute", r.systemHandler.HandleSystemExecute)
	r.mux.HandleFunc("/api/v1/system/reboot", r.systemHandler.HandleSystemReboot)
	r.mux.HandleFunc("/api/v1/system/shutdown", r.systemHandler.HandleSystemShutdown)
	r.mux.HandleFunc("/api/v1/system/logs", r.systemHandler.HandleSystemLogs)
	r.mux.HandleFunc("/api/v1/system/logs/all", r.systemHandler.HandleSystemLogsAll)
}
