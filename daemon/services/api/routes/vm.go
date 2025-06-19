package routes

// registerVMRoutes registers all VM-related endpoints
func (r *Router) registerVMRoutes() {
	// VM management endpoints
	r.mux.HandleFunc("/api/v1/vms", r.vmHandler.HandleVMList)
	r.mux.HandleFunc("/api/v1/vms/", r.vmHandler.HandleVM) // Handles all VM operations with path parsing
}
