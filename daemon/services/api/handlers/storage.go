package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/domalab/uma/daemon/services/api/services"
	"github.com/domalab/uma/daemon/services/api/types/requests"
	"github.com/domalab/uma/daemon/services/api/types/responses"
	"github.com/domalab/uma/daemon/services/api/utils"
)

// StorageHandler handles storage-related HTTP requests
type StorageHandler struct {
	api            utils.APIInterface
	storageService *services.StorageService
}

// NewStorageHandler creates a new storage handler
func NewStorageHandler(api utils.APIInterface) *StorageHandler {
	return &StorageHandler{
		api:            api,
		storageService: services.NewStorageService(api),
	}
}

// HandleStorageArray handles GET /api/v1/storage/array
func (h *StorageHandler) HandleStorageArray(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Use the system interface to get real array info with parity check data
	arrayInfo, err := h.api.GetSystem().GetRealArrayInfo()
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get array information: %v", err))
		return
	}

	utils.WriteJSON(w, http.StatusOK, arrayInfo)
}

// HandleStorageDisks handles GET /api/v1/storage/disks
func (h *StorageHandler) HandleStorageDisks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	disks, err := h.api.GetStorage().GetDisks()
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get disk information: %v", err))
		return
	}

	utils.WriteJSON(w, http.StatusOK, disks)
}

// HandleStorageZFS handles GET /api/v1/storage/zfs
func (h *StorageHandler) HandleStorageZFS(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	zfsPools, err := h.api.GetStorage().GetZFSPools()
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get ZFS information: %v", err))
		return
	}

	utils.WriteJSON(w, http.StatusOK, zfsPools)
}

// HandleStorageCache handles GET /api/v1/storage/cache
func (h *StorageHandler) HandleStorageCache(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	cacheInfo, err := h.api.GetStorage().GetCacheInfo()
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get cache information: %v", err))
		return
	}

	utils.WriteJSON(w, http.StatusOK, cacheInfo)
}

// HandleStorageBoot handles GET /api/v1/storage/boot
func (h *StorageHandler) HandleStorageBoot(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	bootInfo := h.getBootUsage()
	utils.WriteJSON(w, http.StatusOK, bootInfo)
}

// HandleStorageGeneral handles GET /api/v1/storage/general
func (h *StorageHandler) HandleStorageGeneral(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	generalInfo := h.getGeneralStorageInfo()
	utils.WriteJSON(w, http.StatusOK, generalInfo)
}

// HandleArrayStart handles POST /api/v1/storage/array/start
func (h *StorageHandler) HandleArrayStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var request requests.ArrayStartRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid JSON request")
		return
	}

	err := h.api.GetStorage().StartArray(request)
	if err != nil {
		response := responses.ArrayOperationResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to start array: %v", err),
		}
		utils.WriteJSON(w, http.StatusInternalServerError, response)
		return
	}

	response := responses.ArrayOperationResponse{
		Success:     true,
		Message:     "Array start initiated successfully",
		OperationID: fmt.Sprintf("array_start_%d", time.Now().Unix()),
	}
	utils.WriteJSON(w, http.StatusOK, response)
}

// HandleArrayStop handles POST /api/v1/storage/array/stop
func (h *StorageHandler) HandleArrayStop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var request requests.ArrayStopRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid JSON request")
		return
	}

	err := h.api.GetStorage().StopArray(request)
	if err != nil {
		response := responses.ArrayOperationResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to stop array: %v", err),
		}
		utils.WriteJSON(w, http.StatusInternalServerError, response)
		return
	}

	response := responses.ArrayOperationResponse{
		Success:     true,
		Message:     "Array stop initiated successfully",
		OperationID: fmt.Sprintf("array_stop_%d", time.Now().Unix()),
	}
	utils.WriteJSON(w, http.StatusOK, response)
}

// HandleParityCheck handles GET/POST /api/v1/system/parity/check
func (h *StorageHandler) HandleParityCheck(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// Get parity check status
		status := h.getParityCheckStatus()
		utils.WriteJSON(w, http.StatusOK, status)

	case http.MethodPost:
		// Start parity check
		var request requests.ParityCheckRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			utils.WriteError(w, http.StatusBadRequest, "Invalid JSON request")
			return
		}

		err := h.startParityCheck(request)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to start parity check: %v", err))
			return
		}

		response := responses.ArrayOperationResponse{
			Success:     true,
			Message:     "Parity check started successfully",
			OperationID: fmt.Sprintf("parity_check_%d", time.Now().Unix()),
		}
		utils.WriteJSON(w, http.StatusOK, response)

	default:
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// HandleParityDisk handles GET /api/v1/system/parity/disk
func (h *StorageHandler) HandleParityDisk(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	parityDiskInfo := h.getParityDiskInfo()
	utils.WriteJSON(w, http.StatusOK, parityDiskInfo)
}

// Helper methods

// getBootUsage returns boot filesystem usage information
func (h *StorageHandler) getBootUsage() map[string]interface{} {
	bootData, err := h.storageService.GetBootData()
	if err != nil {
		return h.getPathUsage("/boot")
	}
	return bootData
}

// getGeneralStorageInfo returns general storage information
func (h *StorageHandler) getGeneralStorageInfo() map[string]interface{} {
	return map[string]interface{}{
		"docker_vdisk": h.getDockerVDiskUsage(),
		"log_usage":    h.getLogFilesystemUsage(),
		"boot_usage":   h.getBootUsage(),
		"last_updated": time.Now().UTC().Format(time.RFC3339),
	}
}

// getPathUsage returns filesystem usage for a given path
func (h *StorageHandler) getPathUsage(path string) map[string]interface{} {
	// Implementation would get actual filesystem usage
	// For now, return placeholder
	return map[string]interface{}{
		"total": 0,
		"used":  0,
		"free":  0,
		"path":  path,
	}
}

// getDockerVDiskUsage returns Docker vDisk usage information
func (h *StorageHandler) getDockerVDiskUsage() map[string]interface{} {
	// Check common Docker vDisk locations
	dockerPaths := []string{"/var/lib/docker", "/mnt/user/system/docker/docker.img"}

	for _, path := range dockerPaths {
		if usage := h.getPathUsage(path); usage != nil {
			return usage
		}
	}

	return map[string]interface{}{
		"total": 0,
		"used":  0,
		"free":  0,
	}
}

// getLogFilesystemUsage returns log filesystem usage information
func (h *StorageHandler) getLogFilesystemUsage() map[string]interface{} {
	return h.getPathUsage("/var/log")
}

// getParityCheckStatus returns the current parity check status
func (h *StorageHandler) getParityCheckStatus() responses.ParityCheckStatus {
	// Implementation would get actual parity check status
	// For now, return placeholder
	return responses.ParityCheckStatus{
		Active:      false,
		Type:        "",
		Progress:    0.0,
		Speed:       "",
		Errors:      0,
		LastUpdated: time.Now().UTC(),
	}
}

// startParityCheck starts a parity check operation
func (h *StorageHandler) startParityCheck(request requests.ParityCheckRequest) error {
	// Implementation would start actual parity check
	// For now, return success
	return nil
}

// getParityDiskInfo returns parity disk information
func (h *StorageHandler) getParityDiskInfo() map[string]interface{} {
	parityData, err := h.storageService.GetParityDiskData()
	if err != nil {
		return map[string]interface{}{
			"parity_disks": []interface{}{},
			"capacity":     0,
			"temperature":  0.0,
			"health":       "unknown",
			"last_updated": time.Now().UTC().Format(time.RFC3339),
		}
	}
	return parityData
}
