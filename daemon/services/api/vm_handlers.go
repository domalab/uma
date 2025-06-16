package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// handleVMGet handles GET /api/v1/vms/{name}
func (h *HTTPServer) handleVMGet(w http.ResponseWriter, r *http.Request) {
	vmName := chi.URLParam(r, "name")
	if vmName == "" {
		h.writeError(w, http.StatusBadRequest, "VM name is required")
		return
	}

	vm, err := h.api.vm.GetVM(vmName)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "VM not found: "+err.Error())
		return
	}

	h.writeStandardResponse(w, http.StatusOK, vm, nil)
}

// handleVMStats handles GET /api/v1/vms/{name}/stats
func (h *HTTPServer) handleVMStats(w http.ResponseWriter, r *http.Request) {
	vmName := chi.URLParam(r, "name")
	if vmName == "" {
		h.writeError(w, http.StatusBadRequest, "VM name is required")
		return
	}

	stats, err := h.api.vm.GetVMStats(vmName)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to get VM stats: "+err.Error())
		return
	}

	h.writeStandardResponse(w, http.StatusOK, stats, nil)
}

// handleVMConsole handles GET /api/v1/vms/{name}/console
func (h *HTTPServer) handleVMConsole(w http.ResponseWriter, r *http.Request) {
	vmName := chi.URLParam(r, "name")
	if vmName == "" {
		h.writeError(w, http.StatusBadRequest, "VM name is required")
		return
	}

	console, err := h.api.vm.GetVMConsole(vmName)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to get VM console: "+err.Error())
		return
	}

	response := map[string]string{
		"console": console,
		"vm_name": vmName,
	}

	h.writeStandardResponse(w, http.StatusOK, response, nil)
}

// handleVMStart handles POST /api/v1/vms/{name}/start
func (h *HTTPServer) handleVMStart(w http.ResponseWriter, r *http.Request) {
	vmName := chi.URLParam(r, "name")
	if vmName == "" {
		h.writeError(w, http.StatusBadRequest, "VM name is required")
		return
	}

	err := h.api.vm.StartVM(vmName)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to start VM: "+err.Error())
		return
	}

	response := map[string]string{
		"message": "VM started successfully",
		"vm_name": vmName,
		"action":  "start",
	}

	h.writeStandardResponse(w, http.StatusOK, response, nil)
}

// handleVMStop handles POST /api/v1/vms/{name}/stop
func (h *HTTPServer) handleVMStop(w http.ResponseWriter, r *http.Request) {
	vmName := chi.URLParam(r, "name")
	if vmName == "" {
		h.writeError(w, http.StatusBadRequest, "VM name is required")
		return
	}

	force := r.URL.Query().Get("force") == "true"
	err := h.api.vm.StopVM(vmName, force)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to stop VM: "+err.Error())
		return
	}

	response := map[string]interface{}{
		"message": "VM stopped successfully",
		"vm_name": vmName,
		"action":  "stop",
		"force":   force,
	}

	h.writeStandardResponse(w, http.StatusOK, response, nil)
}

// handleVMRestart handles POST /api/v1/vms/{name}/restart
func (h *HTTPServer) handleVMRestart(w http.ResponseWriter, r *http.Request) {
	vmName := chi.URLParam(r, "name")
	if vmName == "" {
		h.writeError(w, http.StatusBadRequest, "VM name is required")
		return
	}

	err := h.api.vm.RestartVM(vmName)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to restart VM: "+err.Error())
		return
	}

	response := map[string]string{
		"message": "VM restarted successfully",
		"vm_name": vmName,
		"action":  "restart",
	}

	h.writeStandardResponse(w, http.StatusOK, response, nil)
}

// handleVMPause handles POST /api/v1/vms/{name}/pause
func (h *HTTPServer) handleVMPause(w http.ResponseWriter, r *http.Request) {
	vmName := chi.URLParam(r, "name")
	if vmName == "" {
		h.writeError(w, http.StatusBadRequest, "VM name is required")
		return
	}

	err := h.api.vm.PauseVM(vmName)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to pause VM: "+err.Error())
		return
	}

	response := map[string]string{
		"message": "VM paused successfully",
		"vm_name": vmName,
		"action":  "pause",
	}

	h.writeStandardResponse(w, http.StatusOK, response, nil)
}

// handleVMResume handles POST /api/v1/vms/{name}/resume
func (h *HTTPServer) handleVMResume(w http.ResponseWriter, r *http.Request) {
	vmName := chi.URLParam(r, "name")
	if vmName == "" {
		h.writeError(w, http.StatusBadRequest, "VM name is required")
		return
	}

	err := h.api.vm.ResumeVM(vmName)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to resume VM: "+err.Error())
		return
	}

	response := map[string]string{
		"message": "VM resumed successfully",
		"vm_name": vmName,
		"action":  "resume",
	}

	h.writeStandardResponse(w, http.StatusOK, response, nil)
}

// handleVMHibernate handles POST /api/v1/vms/{name}/hibernate
func (h *HTTPServer) handleVMHibernate(w http.ResponseWriter, r *http.Request) {
	vmName := chi.URLParam(r, "name")
	if vmName == "" {
		h.writeError(w, http.StatusBadRequest, "VM name is required")
		return
	}

	err := h.api.vm.HibernateVM(vmName)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to hibernate VM: "+err.Error())
		return
	}

	response := map[string]string{
		"message": "VM hibernated successfully",
		"vm_name": vmName,
		"action":  "hibernate",
	}

	h.writeStandardResponse(w, http.StatusOK, response, nil)
}

// handleVMRestore handles POST /api/v1/vms/{name}/restore
func (h *HTTPServer) handleVMRestore(w http.ResponseWriter, r *http.Request) {
	vmName := chi.URLParam(r, "name")
	if vmName == "" {
		h.writeError(w, http.StatusBadRequest, "VM name is required")
		return
	}

	err := h.api.vm.RestoreVM(vmName)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to restore VM: "+err.Error())
		return
	}

	response := map[string]string{
		"message": "VM restored successfully",
		"vm_name": vmName,
		"action":  "restore",
	}

	h.writeStandardResponse(w, http.StatusOK, response, nil)
}

// handleVMAutostart handles POST /api/v1/vms/{name}/autostart
func (h *HTTPServer) handleVMAutostart(w http.ResponseWriter, r *http.Request) {
	vmName := chi.URLParam(r, "name")
	if vmName == "" {
		h.writeError(w, http.StatusBadRequest, "VM name is required")
		return
	}

	enable := r.URL.Query().Get("enable") == "true"
	err := h.api.vm.SetVMAutostart(vmName, enable)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to set VM autostart: "+err.Error())
		return
	}

	response := map[string]interface{}{
		"message":   "VM autostart updated successfully",
		"vm_name":   vmName,
		"action":    "autostart",
		"autostart": enable,
	}

	h.writeStandardResponse(w, http.StatusOK, response, nil)
}
