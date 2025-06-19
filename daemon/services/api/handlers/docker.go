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

// DockerHandler handles Docker-related HTTP requests
type DockerHandler struct {
	api           utils.APIInterface
	dockerService *services.DockerService
}

// NewDockerHandler creates a new Docker handler
func NewDockerHandler(api utils.APIInterface) *DockerHandler {
	return &DockerHandler{
		api:           api,
		dockerService: services.NewDockerService(api),
	}
}

// HandleDockerContainers handles GET /api/v1/docker/containers
func (h *DockerHandler) HandleDockerContainers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	containers, err := h.api.GetDocker().GetContainers()
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get containers: %v", err))
		return
	}

	utils.WriteJSON(w, http.StatusOK, containers)
}

// HandleDockerContainer handles individual Docker container operations
func (h *DockerHandler) HandleDockerContainer(w http.ResponseWriter, r *http.Request) {
	// Extract container ID and operation from URL path
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/docker/containers/")
	parts := strings.Split(path, "/")

	if len(parts) < 2 {
		utils.WriteError(w, http.StatusBadRequest, "Container ID and operation are required")
		return
	}

	containerID := parts[0]
	operation := parts[1]

	if containerID == "" {
		utils.WriteError(w, http.StatusBadRequest, "Container ID is required")
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.handleGetContainer(w, r, containerID, operation)
	case http.MethodPost:
		h.handleContainerAction(w, r, containerID, operation)
	default:
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// HandleDockerImages handles GET /api/v1/docker/images
func (h *DockerHandler) HandleDockerImages(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	images, err := h.api.GetDocker().GetImages()
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get images: %v", err))
		return
	}

	utils.WriteJSON(w, http.StatusOK, images)
}

// HandleDockerNetworks handles GET /api/v1/docker/networks
func (h *DockerHandler) HandleDockerNetworks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	networks, err := h.api.GetDocker().GetNetworks()
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get networks: %v", err))
		return
	}

	utils.WriteJSON(w, http.StatusOK, networks)
}

// HandleDockerInfo handles GET /api/v1/docker/info
func (h *DockerHandler) HandleDockerInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	info, err := h.api.GetDocker().GetSystemInfo()
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get Docker info: %v", err))
		return
	}

	utils.WriteJSON(w, http.StatusOK, info)
}

// HandleDockerBulkStart handles POST /api/v1/docker/containers/bulk/start
func (h *DockerHandler) HandleDockerBulkStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var request requests.DockerBulkActionRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid JSON request")
		return
	}

	if len(request.ContainerIDs) == 0 {
		utils.WriteError(w, http.StatusBadRequest, "Container IDs list cannot be empty")
		return
	}

	results := h.performBulkAction(request.ContainerIDs, "start", request.Force)
	utils.WriteBulkOperationResponse(w, http.StatusOK, results)
}

// HandleDockerBulkStop handles POST /api/v1/docker/containers/bulk/stop
func (h *DockerHandler) HandleDockerBulkStop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var request requests.DockerBulkActionRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid JSON request")
		return
	}

	if len(request.ContainerIDs) == 0 {
		utils.WriteError(w, http.StatusBadRequest, "Container IDs list cannot be empty")
		return
	}

	results := h.performBulkAction(request.ContainerIDs, "stop", request.Force)
	utils.WriteBulkOperationResponse(w, http.StatusOK, results)
}

// HandleDockerBulkRestart handles POST /api/v1/docker/containers/bulk/restart
func (h *DockerHandler) HandleDockerBulkRestart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var request requests.DockerBulkActionRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid JSON request")
		return
	}

	if len(request.ContainerIDs) == 0 {
		utils.WriteError(w, http.StatusBadRequest, "Container IDs list cannot be empty")
		return
	}

	results := h.performBulkAction(request.ContainerIDs, "restart", request.Force)
	utils.WriteBulkOperationResponse(w, http.StatusOK, results)
}

// Helper methods

// handleGetContainer handles GET requests for individual containers
func (h *DockerHandler) handleGetContainer(w http.ResponseWriter, r *http.Request, containerID, operation string) {
	switch operation {
	case "info", "":
		container, err := h.api.GetDocker().GetContainer(containerID)
		if err != nil {
			utils.WriteError(w, http.StatusNotFound, fmt.Sprintf("Container not found: %v", err))
			return
		}
		utils.WriteJSON(w, http.StatusOK, container)

	case "logs":
		// Implementation would get container logs
		logs := map[string]interface{}{
			"container_id": containerID,
			"logs":         []string{"Log functionality not implemented"},
			"timestamp":    time.Now().UTC().Format(time.RFC3339),
		}
		utils.WriteJSON(w, http.StatusOK, logs)

	case "stats":
		// Implementation would get container stats
		stats := map[string]interface{}{
			"container_id": containerID,
			"cpu_percent":  0.0,
			"memory_usage": 0,
			"timestamp":    time.Now().UTC().Format(time.RFC3339),
		}
		utils.WriteJSON(w, http.StatusOK, stats)

	default:
		utils.WriteError(w, http.StatusBadRequest, "Invalid operation")
	}
}

// handleContainerAction handles POST requests for container actions
func (h *DockerHandler) handleContainerAction(w http.ResponseWriter, r *http.Request, containerID, operation string) {
	var err error

	switch operation {
	case "start":
		err = h.api.GetDocker().StartContainer(containerID)
	case "stop":
		err = h.api.GetDocker().StopContainer(containerID)
	case "restart":
		err = h.api.GetDocker().RestartContainer(containerID)
	default:
		utils.WriteError(w, http.StatusBadRequest, "Invalid operation")
		return
	}

	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to %s container: %v", operation, err))
		return
	}

	response := map[string]interface{}{
		"message":      fmt.Sprintf("Container %s %sed successfully", containerID, operation),
		"container_id": containerID,
		"operation":    operation,
		"timestamp":    time.Now().UTC().Format(time.RFC3339),
	}
	utils.WriteJSON(w, http.StatusOK, response)
}

// performBulkAction performs bulk actions on multiple containers
func (h *DockerHandler) performBulkAction(containerIDs []string, action string, force bool) []responses.BulkOperationResult {
	results := make([]responses.BulkOperationResult, len(containerIDs))

	for i, containerID := range containerIDs {
		var err error

		switch action {
		case "start":
			err = h.api.GetDocker().StartContainer(containerID)
		case "stop":
			err = h.api.GetDocker().StopContainer(containerID)
		case "restart":
			err = h.api.GetDocker().RestartContainer(containerID)
		default:
			err = fmt.Errorf("invalid action: %s", action)
		}

		if err != nil {
			results[i] = responses.BulkOperationResult{
				ID:      containerID,
				Success: false,
				Error:   err.Error(),
			}
		} else {
			results[i] = responses.BulkOperationResult{
				ID:      containerID,
				Success: true,
				Message: fmt.Sprintf("Container %s %sed successfully", containerID, action),
			}
		}
	}

	return results
}

// GetDockerDataOptimized returns optimized Docker data using the Docker service
func (h *DockerHandler) GetDockerDataOptimized() interface{} {
	return h.dockerService.GetDockerDataOptimized()
}
