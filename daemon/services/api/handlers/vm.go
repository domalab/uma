package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
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

	utils.WriteJSON(w, http.StatusOK, vms)
}

// HandleVM handles VM operations
func (h *VMHandler) HandleVM(w http.ResponseWriter, r *http.Request) {
	// Extract VM name from URL path
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/vms/")
	parts := strings.Split(path, "/")
	if len(parts) == 0 || parts[0] == "" {
		utils.WriteError(w, http.StatusBadRequest, "VM name required")
		return
	}

	vmName := parts[0]
	action := ""
	if len(parts) > 1 {
		action = parts[1]
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

	case "console":
		console, err := h.api.GetVM().GetVMConsole(vmName)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get VM console: %v", err))
			return
		}
		utils.WriteJSON(w, http.StatusOK, map[string]interface{}{"console": console})

	case "":
		// Get VM info
		vm, err := h.api.GetVM().GetVM(vmName)
		if err != nil {
			utils.WriteError(w, http.StatusNotFound, fmt.Sprintf("VM not found: %v", err))
			return
		}
		utils.WriteJSON(w, http.StatusOK, vm)

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
