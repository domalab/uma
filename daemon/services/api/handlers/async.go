package handlers

import (
	"net/http"
	"strings"

	"github.com/domalab/uma/daemon/services/api/utils"
)

// AsyncHandler handles async operation-related HTTP requests
type AsyncHandler struct {
	api utils.APIInterface
}

// NewAsyncHandler creates a new async handler
func NewAsyncHandler(api utils.APIInterface) *AsyncHandler {
	return &AsyncHandler{
		api: api,
	}
}

// HandleAsyncOperations handles GET /api/v1/operations
func (h *AsyncHandler) HandleAsyncOperations(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Return placeholder async operations data with required fields
	operations := []map[string]interface{}{
		{
			"id":          "op-001",
			"type":        "bulk_container", // Use valid enum value instead of "docker_bulk_start"
			"status":      "completed",
			"progress":    100,
			"description": "Bulk container start operation", // Add required field
			"started":     "2024-01-01T00:00:00Z",           // Add required field (renamed from created_at)
			"completed":   "2024-01-01T00:01:00Z",           // Rename from completed_at
			"cancellable": false,
		},
		{
			"id":          "op-002",
			"type":        "array_start",
			"status":      "running",
			"progress":    75,
			"description": "Array start operation", // Add required field
			"started":     "2024-01-01T00:02:00Z",  // Add required field (renamed from created_at)
			"cancellable": true,
		},
	}

	// Return array directly to match schema expectation
	utils.WriteJSON(w, http.StatusOK, operations)
}

// HandleAsyncOperation handles GET /api/v1/operations/{id}
func (h *AsyncHandler) HandleAsyncOperation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Extract operation ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/operations/")
	if path == "" || path == "stats" {
		utils.WriteError(w, http.StatusBadRequest, "Operation ID required")
		return
	}

	// Return placeholder operation data with required fields
	operation := map[string]interface{}{
		"id":          path,
		"type":        "bulk_container", // Use valid enum value
		"status":      "completed",
		"progress":    100,
		"description": "Bulk container operation", // Add required field
		"started":     "2024-01-01T00:00:00Z",     // Add required field (renamed from created_at)
		"completed":   "2024-01-01T00:01:00Z",     // Rename from completed_at
		"cancellable": false,
		"result": map[string]interface{}{
			"success": true,
			"message": "Operation completed successfully",
		},
	}

	utils.WriteJSON(w, http.StatusOK, operation)
}

// HandleAsyncStats handles GET /api/v1/operations/stats
func (h *AsyncHandler) HandleAsyncStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Return stats with required fields matching schema
	stats := map[string]interface{}{
		// Required fields per schema
		"total":     10,
		"active":    1,
		"completed": 8,
		"failed":    1,
		// Additional fields (keeping existing data)
		"total_operations":     10,
		"running_operations":   1,
		"completed_operations": 8,
		"failed_operations":    1,
		"average_duration":     "45s",
		"operations_by_type": map[string]int{
			"docker_bulk_start": 3,
			"docker_bulk_stop":  2,
			"array_start":       2,
			"array_stop":        1,
			"vm_start":          1,
			"vm_stop":           1,
		},
	}

	utils.WriteJSON(w, http.StatusOK, stats)
}
