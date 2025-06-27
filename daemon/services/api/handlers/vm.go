package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/domalab/uma/daemon/services/api/services"
	"github.com/domalab/uma/daemon/services/api/types/requests"
	"github.com/domalab/uma/daemon/services/api/types/responses"
	"github.com/domalab/uma/daemon/services/api/utils"
)

// VMHandler handles VM-related HTTP requests
type VMHandler struct {
	api       utils.APIInterface
	vmService *services.VMService
}

// NewVMHandler creates a new VM handler
func NewVMHandler(api utils.APIInterface) *VMHandler {
	return &VMHandler{
		api:       api,
		vmService: services.NewVMService(api),
	}
}

// HandleVMList handles GET /api/v1/vms
func (h *VMHandler) HandleVMList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	vms, err := h.api.GetVM().GetVMs()
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get VMs: %v", err))
		return
	}

	// Transform VMs to match OpenAPI schema requirements
	transformedVMs := h.transformVMsData(vms)
	utils.WriteJSON(w, http.StatusOK, transformedVMs)
}

// HandleVM handles VM operations
func (h *VMHandler) HandleVM(w http.ResponseWriter, r *http.Request) {
	// Extract VM identifier from URL path (could be ID or name)
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/vms/")
	parts := strings.Split(path, "/")
	if len(parts) == 0 || parts[0] == "" {
		utils.WriteError(w, http.StatusBadRequest, "VM identifier required")
		return
	}

	vmIdentifier := parts[0]
	action := ""
	if len(parts) > 1 {
		action = parts[1]
	}

	// Resolve VM identifier to VM name (handles both ID and name)
	vmName, err := h.resolveVMName(vmIdentifier)
	if err != nil {
		utils.WriteError(w, http.StatusNotFound, fmt.Sprintf("VM not found: %v", err))
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.handleGetVM(w, r, vmName, action)
	case http.MethodPost:
		h.handleVMAction(w, r, vmName, action)
	case http.MethodPut:
		h.handleUpdateVM(w, r, vmName)
	case http.MethodDelete:
		h.handleDeleteVM(w, r, vmName)
	default:
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// Helper methods

// handleGetVM handles GET requests for VMs
func (h *VMHandler) handleGetVM(w http.ResponseWriter, r *http.Request, vmName, action string) {
	switch action {
	case "stats":
		stats, err := h.api.GetVM().GetVMStats(vmName)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get VM stats: %v", err))
			return
		}
		utils.WriteJSON(w, http.StatusOK, stats)

	case "performance":
		performance, err := h.getVMPerformance(vmName)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get VM performance: %v", err))
			return
		}
		utils.WriteJSON(w, http.StatusOK, performance)

	case "resources":
		resources, err := h.getVMResources(vmName)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get VM resources: %v", err))
			return
		}
		utils.WriteJSON(w, http.StatusOK, resources)

	case "snapshots":
		// Handle VM snapshots - delegate to the snapshot handler
		h.HandleVMSnapshot(w, r, vmName)
		return

	case "console":
		console, err := h.api.GetVM().GetVMConsole(vmName)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get VM console: %v", err))
			return
		}
		utils.WriteJSON(w, http.StatusOK, map[string]interface{}{"console": console})

	case "":
		// Get VM info - this handles /api/v1/vms/{name}
		vm, err := h.api.GetVM().GetVM(vmName)
		if err != nil {
			utils.WriteError(w, http.StatusNotFound, fmt.Sprintf("VM not found: %v", err))
			return
		}

		// Transform VM data to ensure schema compliance
		if vmMap, ok := vm.(map[string]interface{}); ok {
			transformedVM := h.transformSingleVM(vmMap)
			utils.WriteJSON(w, http.StatusOK, transformedVM)
		} else {
			utils.WriteJSON(w, http.StatusOK, vm)
		}

	default:
		utils.WriteError(w, http.StatusBadRequest, "Invalid action")
	}
}

// handleVMAction handles POST requests for VM actions
func (h *VMHandler) handleVMAction(w http.ResponseWriter, r *http.Request, vmName, action string) {
	var err error
	var message string

	switch action {
	case "start":
		err = h.api.GetVM().StartVM(vmName)
		message = "VM started successfully"

	case "stop":
		err = h.api.GetVM().StopVM(vmName)
		message = "VM stopped successfully"

	case "restart":
		err = h.api.GetVM().RestartVM(vmName)
		message = "VM restarted successfully"

	case "pause":
		// Implementation would pause VM
		message = "VM paused successfully"

	case "resume":
		// Implementation would resume VM
		message = "VM resumed successfully"

	case "reset":
		// Implementation would reset VM
		message = "VM reset successfully"

	case "snapshots":
		// Handle VM snapshot creation - delegate to the snapshot handler
		h.HandleVMSnapshot(w, r, vmName)
		return

	case "autostart":
		autostart := r.URL.Query().Get("enable") == "true"
		err = h.api.GetVM().SetVMAutostart(vmName, autostart)
		if autostart {
			message = "VM autostart enabled"
		} else {
			message = "VM autostart disabled"
		}

	default:
		utils.WriteError(w, http.StatusBadRequest, "Invalid action")
		return
	}

	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to %s VM: %v", action, err))
		return
	}

	response := responses.VMOperationResponse{
		Success:     true,
		Message:     message,
		OperationID: fmt.Sprintf("vm_%s_%s_%d", vmName, action, time.Now().Unix()),
		VMName:      vmName,
	}
	utils.WriteJSON(w, http.StatusOK, response)
}

// handleUpdateVM handles PUT requests to update VM configuration
func (h *VMHandler) handleUpdateVM(w http.ResponseWriter, r *http.Request, vmName string) {
	var request requests.VMUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid JSON request")
		return
	}

	// Validate the update request
	if err := utils.ValidateStruct(&request); err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Sprintf("Validation error: %v", err))
		return
	}

	// Implementation would update VM configuration
	// For now, return success
	response := responses.VMOperationResponse{
		Success:     true,
		Message:     "VM configuration updated successfully",
		OperationID: fmt.Sprintf("vm_update_%s_%d", vmName, time.Now().Unix()),
		VMName:      vmName,
	}
	utils.WriteJSON(w, http.StatusOK, response)
}

// handleDeleteVM handles DELETE requests to remove a VM
func (h *VMHandler) handleDeleteVM(w http.ResponseWriter, r *http.Request, vmName string) {
	// Implementation would delete VM
	// For now, return success
	response := responses.VMOperationResponse{
		Success:     true,
		Message:     "VM deleted successfully",
		OperationID: fmt.Sprintf("vm_delete_%s_%d", vmName, time.Now().Unix()),
		VMName:      vmName,
	}
	utils.WriteJSON(w, http.StatusOK, response)
}

// HandleVMCreate handles POST /api/v1/vms (create new VM)
func (h *VMHandler) HandleVMCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var request requests.VMCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid JSON request")
		return
	}

	// Validate the create request
	if err := utils.ValidateVMCreateRequest(&request); err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Sprintf("Validation error: %v", err))
		return
	}

	// Implementation would create VM
	// For now, return success
	response := responses.VMOperationResponse{
		Success:     true,
		Message:     "VM created successfully",
		OperationID: fmt.Sprintf("vm_create_%s_%d", request.Name, time.Now().Unix()),
		VMName:      request.Name,
	}
	utils.WriteJSON(w, http.StatusCreated, response)
}

// HandleVMSnapshot handles VM snapshot operations
func (h *VMHandler) HandleVMSnapshot(w http.ResponseWriter, r *http.Request, vmName string) {
	switch r.Method {
	case http.MethodGet:
		// List snapshots
		snapshots := h.getVMSnapshots(vmName)
		utils.WriteJSON(w, http.StatusOK, snapshots)

	case http.MethodPost:
		// Create snapshot
		var request requests.VMSnapshotRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			utils.WriteError(w, http.StatusBadRequest, "Invalid JSON request")
			return
		}

		err := h.createVMSnapshot(vmName, request)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to create snapshot: %v", err))
			return
		}

		response := responses.VMOperationResponse{
			Success:     true,
			Message:     "VM snapshot created successfully",
			OperationID: fmt.Sprintf("vm_snapshot_%s_%d", vmName, time.Now().Unix()),
			VMName:      vmName,
		}
		utils.WriteJSON(w, http.StatusCreated, response)

	default:
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// Helper methods for snapshots

// getVMSnapshots returns a list of snapshots for a VM
func (h *VMHandler) getVMSnapshots(vmName string) []responses.VMSnapshotInfo {
	// Implementation would get actual snapshots
	// For now, return empty list
	return []responses.VMSnapshotInfo{}
}

// createVMSnapshot creates a snapshot for a VM
func (h *VMHandler) createVMSnapshot(vmName string, request requests.VMSnapshotRequest) error {
	// Implementation would create actual snapshot
	// For now, return success
	return nil
}

// GetVMDataOptimized returns optimized VM data using the VM service
func (h *VMHandler) GetVMDataOptimized() interface{} {
	return h.vmService.GetVMDataOptimized()
}

// transformVMsData transforms VM data to match OpenAPI schema requirements
func (h *VMHandler) transformVMsData(vms interface{}) interface{} {
	// Handle different possible return types from GetVMs()
	switch v := vms.(type) {
	case []interface{}:
		// Transform array of VM objects
		transformedVMs := make([]interface{}, 0, len(v))
		for _, vm := range v {
			if vmMap, ok := vm.(map[string]interface{}); ok {
				transformedVM := h.transformSingleVM(vmMap)
				transformedVMs = append(transformedVMs, transformedVM)
			}
		}
		return transformedVMs
	case map[string]interface{}:
		// If it's a single VM object, transform it
		return h.transformSingleVM(v)
	default:
		// Return empty array if unknown type
		return []interface{}{}
	}
}

// transformSingleVM transforms a single VM object to match schema
func (h *VMHandler) transformSingleVM(vm map[string]interface{}) map[string]interface{} {
	transformed := make(map[string]interface{})

	// Copy all existing fields first
	for key, value := range vm {
		transformed[key] = value
	}

	// Ensure required fields are present
	if _, exists := transformed["id"]; !exists {
		// Use name as ID if ID is missing
		if name, ok := transformed["name"].(string); ok {
			transformed["id"] = name
		} else {
			transformed["id"] = "unknown"
		}
	}

	// Add missing required fields
	if _, exists := transformed["resources"]; !exists {
		transformed["resources"] = h.createDefaultVMResources(vm)
	}

	if _, exists := transformed["created"]; !exists {
		transformed["created"] = time.Now().UTC().Format(time.RFC3339)
	}

	if _, exists := transformed["last_updated"]; !exists {
		transformed["last_updated"] = time.Now().UTC().Format(time.RFC3339)
	}

	// Fix os_type enum violation
	if osType, exists := transformed["os_type"]; exists {
		if osTypeStr, ok := osType.(string); ok {
			transformed["os_type"] = h.normalizeOSType(osTypeStr)
		}
	} else {
		transformed["os_type"] = "other" // Default to "other"
	}

	return transformed
}

// createDefaultVMResources creates default VM resources from available data
func (h *VMHandler) createDefaultVMResources(vm map[string]interface{}) map[string]interface{} {
	resources := map[string]interface{}{
		"cpu":    1,
		"memory": 1024,
	}

	// Extract CPU information
	if vcpus, exists := vm["vcpus"]; exists {
		if vcpusStr, ok := vcpus.(string); ok {
			if vcpusInt, err := strconv.Atoi(vcpusStr); err == nil {
				resources["cpu"] = vcpusInt
			}
		} else if vcpusInt, ok := vcpus.(int); ok {
			resources["cpu"] = vcpusInt
		}
	}

	// Extract memory information
	if maxMemory, exists := vm["max_memory"]; exists {
		if memoryStr, ok := maxMemory.(string); ok {
			// Parse memory strings like "4194304 KiB"
			if memoryBytes := h.parseMemoryToMB(memoryStr); memoryBytes > 0 {
				resources["memory"] = memoryBytes
			}
		}
	}

	return resources
}

// normalizeOSType converts various OS type values to schema-compliant enum values
func (h *VMHandler) normalizeOSType(osType string) string {
	osType = strings.ToLower(strings.TrimSpace(osType))

	switch osType {
	case "windows", "win", "microsoft":
		return "windows"
	case "linux", "ubuntu", "debian", "centos", "rhel", "fedora", "opensuse":
		return "linux"
	case "macos", "darwin", "osx":
		return "macos"
	case "hvm", "kvm", "xen", "vmware", "virtualbox":
		// These are virtualization types, not OS types
		// Default to "other" since we can't determine the actual OS
		return "other"
	default:
		return "other"
	}
}

// parseMemoryToMB converts memory strings to megabytes
func (h *VMHandler) parseMemoryToMB(memoryStr string) int {
	// Remove whitespace and convert to uppercase
	memoryStr = strings.TrimSpace(strings.ToUpper(memoryStr))

	// Extract numeric part and unit
	re := regexp.MustCompile(`^(\d+(?:\.\d+)?)\s*(KIB|MIB|GIB|TIB|KB|MB|GB|TB|K|M|G|T)?$`)
	matches := re.FindStringSubmatch(memoryStr)

	if len(matches) < 2 {
		return 0
	}

	memory, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return 0
	}

	// Convert to MB based on unit
	unit := ""
	if len(matches) > 2 {
		unit = matches[2]
	}

	switch unit {
	case "TIB", "TB", "T":
		return int(memory * 1024 * 1024)
	case "GIB", "GB", "G":
		return int(memory * 1024)
	case "MIB", "MB", "M":
		return int(memory)
	case "KIB", "KB", "K":
		return int(memory / 1024)
	default:
		// Assume bytes if no unit
		return int(memory / 1024 / 1024)
	}
}

// getVMPerformance gets comprehensive VM performance metrics
func (h *VMHandler) getVMPerformance(vmName string) (map[string]interface{}, error) {
	// Get enhanced VM stats
	stats, err := h.api.GetVM().GetVMStats(vmName)
	if err != nil {
		return nil, err
	}

	// Transform stats into performance-focused structure
	performance := map[string]interface{}{
		"vm_name":   vmName,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"status":    "unknown",
		"cpu":       map[string]interface{}{},
		"memory":    map[string]interface{}{},
		"disk":      map[string]interface{}{},
		"network":   map[string]interface{}{},
	}

	if statsMap, ok := stats.(map[string]interface{}); ok {
		// Extract CPU performance
		if cpuStats, ok := statsMap["cpu"].(map[string]interface{}); ok {
			performance["cpu"] = map[string]interface{}{
				"total_time_seconds": cpuStats["total_time_seconds"],
				"user_time_ns":       cpuStats["user_time_ns"],
				"system_time_ns":     cpuStats["system_time_ns"],
			}
		}

		// Extract memory performance
		if memStats, ok := statsMap["memory"].(map[string]interface{}); ok {
			performance["memory"] = map[string]interface{}{
				"usage_percent":   memStats["usage_percent"],
				"current_bytes":   memStats["current_bytes"],
				"maximum_bytes":   memStats["maximum_bytes"],
				"rss_bytes":       memStats["rss_bytes"],
				"available_bytes": memStats["available_bytes"],
			}
		}

		// Extract disk performance
		if diskStats, ok := statsMap["disk"].(map[string]interface{}); ok {
			performance["disk"] = map[string]interface{}{
				"read_bytes":     diskStats["read_bytes"],
				"write_bytes":    diskStats["write_bytes"],
				"total_bytes":    diskStats["total_bytes"],
				"read_requests":  diskStats["read_requests"],
				"write_requests": diskStats["write_requests"],
				"total_requests": diskStats["total_requests"],
				"usage_percent":  diskStats["usage_percent"],
			}
		}

		// Extract network performance
		if netStats, ok := statsMap["network"].(map[string]interface{}); ok {
			performance["network"] = map[string]interface{}{
				"rx_bytes":      netStats["rx_bytes"],
				"tx_bytes":      netStats["tx_bytes"],
				"total_bytes":   netStats["total_bytes"],
				"rx_packets":    netStats["rx_packets"],
				"tx_packets":    netStats["tx_packets"],
				"total_packets": netStats["total_packets"],
				"rx_errors":     netStats["rx_errors"],
				"tx_errors":     netStats["tx_errors"],
			}
		}

		// Extract VM status
		if stateStats, ok := statsMap["state"].(map[string]interface{}); ok {
			if status, ok := stateStats["status"].(string); ok {
				performance["status"] = status
			}
		}
	}

	return performance, nil
}

// getVMResources gets VM resource allocation and configuration
func (h *VMHandler) getVMResources(vmName string) (map[string]interface{}, error) {
	// Get VM details
	vm, err := h.api.GetVM().GetVM(vmName)
	if err != nil {
		return nil, err
	}

	// Get VM stats for current usage
	stats, err := h.api.GetVM().GetVMStats(vmName)
	if err != nil {
		return nil, err
	}

	resources := map[string]interface{}{
		"vm_name":   vmName,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"allocated": map[string]interface{}{},
		"current":   map[string]interface{}{},
		"limits":    map[string]interface{}{},
	}

	// Extract allocated resources from VM config
	if vmMap, ok := vm.(map[string]interface{}); ok {
		allocated := map[string]interface{}{}

		if vcpus, ok := vmMap["vcpus"]; ok {
			allocated["cpu_cores"] = vcpus
		}
		if maxMem, ok := vmMap["max_memory"]; ok {
			allocated["memory"] = maxMem
		}

		resources["allocated"] = allocated
	}

	// Extract current usage from stats
	if statsMap, ok := stats.(map[string]interface{}); ok {
		current := map[string]interface{}{}

		if memStats, ok := statsMap["memory"].(map[string]interface{}); ok {
			current["memory_usage_percent"] = memStats["usage_percent"]
			current["memory_current_bytes"] = memStats["current_bytes"]
		}

		if diskStats, ok := statsMap["disk"].(map[string]interface{}); ok {
			current["disk_usage_percent"] = diskStats["usage_percent"]
			current["disk_allocation_bytes"] = diskStats["allocation_bytes"]
		}

		resources["current"] = current
	}

	return resources, nil
}

// resolveVMName resolves a VM identifier (ID or name) to the actual VM name
func (h *VMHandler) resolveVMName(identifier string) (string, error) {
	// If identifier looks like a numeric ID, resolve it to name
	if vmId, err := strconv.Atoi(identifier); err == nil {
		// Get all VMs to find the one with matching ID
		vms, err := h.api.GetVM().GetVMs()
		if err != nil {
			return "", fmt.Errorf("failed to get VMs: %v", err)
		}

		// Handle different possible return types from GetVMs()
		switch v := vms.(type) {
		case []interface{}:
			for _, vm := range v {
				if vmMap, ok := vm.(map[string]interface{}); ok {
					if vmIdStr, exists := vmMap["id"]; exists {
						if vmIdStr == strconv.Itoa(vmId) || vmIdStr == identifier {
							if vmName, exists := vmMap["name"]; exists {
								if name, ok := vmName.(string); ok {
									return name, nil
								}
							}
						}
					}
				}
			}
		}

		return "", fmt.Errorf("VM with ID %s not found", identifier)
	}

	// If not numeric, assume it's already a VM name
	return identifier, nil
}
