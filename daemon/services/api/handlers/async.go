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

	// Return placeholder async operations data
	operations := []map[string]interface{}{
		{
			"id":           "op-001",
			"type":         "docker_bulk_start",
			"status":       "completed",
			"progress":     100,
			"created_at":   "2024-01-01T00:00:00Z",
			"completed_at": "2024-01-01T00:01:00Z",
		},
		{
			"id":         "op-002",
			"type":       "array_start",
			"status":     "running",
			"progress":   75,
			"created_at": "2024-01-01T00:02:00Z",
		},
	}

	utils.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"operations": operations,
		"total":      len(operations),
	})
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

	// Return placeholder operation data
	operation := map[string]interface{}{
		"id":           path,
		"type":         "docker_bulk_start",
		"status":       "completed",
		"progress":     100,
		"created_at":   "2024-01-01T00:00:00Z",
		"completed_at": "2024-01-01T00:01:00Z",
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

	// Return placeholder async operation stats
	stats := map[string]interface{}{
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
