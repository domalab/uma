package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/domalab/uma/daemon/logger"
)

// BulkOperationRequest represents a bulk operation request
type BulkOperationRequest struct {
	ContainerIDs []string `json:"container_ids"`
}

// BulkOperationResponse represents a bulk operation response
type BulkOperationResponse struct {
	Operation string                     `json:"operation"`
	Results   []ContainerOperationResult `json:"results"`
	Summary   BulkOperationSummary       `json:"summary"`
}

// ContainerOperationResult represents the result of an operation on a single container
type ContainerOperationResult struct {
	ContainerID   string `json:"container_id"`
	ContainerName string `json:"container_name,omitempty"`
	Success       bool   `json:"success"`
	Error         string `json:"error,omitempty"`
	Duration      string `json:"duration,omitempty"`
}

// BulkOperationSummary provides a summary of the bulk operation
type BulkOperationSummary struct {
	Total     int `json:"total"`
	Succeeded int `json:"succeeded"`
	Failed    int `json:"failed"`
}

// handleDockerBulkStart handles POST /api/v1/docker/containers/bulk/start
func (h *HTTPServer) handleDockerBulkStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	h.handleDockerBulkOperation(w, r, "start")
}

// handleDockerBulkStop handles POST /api/v1/docker/containers/bulk/stop
func (h *HTTPServer) handleDockerBulkStop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	h.handleDockerBulkOperation(w, r, "stop")
}

// handleDockerBulkRestart handles POST /api/v1/docker/containers/bulk/restart
func (h *HTTPServer) handleDockerBulkRestart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	h.handleDockerBulkOperation(w, r, "restart")
}

// handleDockerBulkOperation handles bulk Docker operations
func (h *HTTPServer) handleDockerBulkOperation(w http.ResponseWriter, r *http.Request, operation string) {
	// Parse request body
	var request BulkOperationRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid JSON request body")
		return
	}

	// Validate request using enhanced validation
	if err := h.validateBulkRequest(&request); err != nil {
		h.writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Get request metadata and track timing
	requestID := h.getRequestIDFromContext(r)
	startTime := time.Now()
	logger.Green("Starting bulk %s operation for %d containers [%s]", operation, len(request.ContainerIDs), requestID)

	// Perform bulk operation
	results := h.performBulkDockerOperation(request.ContainerIDs, operation)

	// Calculate summary
	summary := BulkOperationSummary{
		Total: len(results),
	}
	for _, result := range results {
		if result.Success {
			summary.Succeeded++
		} else {
			summary.Failed++
		}
	}

	// Create response
	response := BulkOperationResponse{
		Operation: operation,
		Results:   results,
		Summary:   summary,
	}

	// Record metrics for the bulk operation
	operationDuration := time.Since(startTime)
	RecordBulkOperation(operation, summary.Total, summary.Succeeded, summary.Failed, operationDuration, requestID)

	// Use standardized response format
	h.writeStandardResponse(w, http.StatusOK, response, nil)

	logger.Green("Bulk %s operation completed: %d succeeded, %d failed [%s]",
		operation, summary.Succeeded, summary.Failed, requestID)
}

// performBulkDockerOperation performs the actual bulk operation on containers
func (h *HTTPServer) performBulkDockerOperation(containerIDs []string, operation string) []ContainerOperationResult {
	results := make([]ContainerOperationResult, len(containerIDs))

	// Get container information first to validate IDs and get names
	containers, err := h.api.docker.ListContainers(true) // Include all containers
	if err != nil {
		// If we can't list containers, mark all as failed
		for i, containerID := range containerIDs {
			results[i] = ContainerOperationResult{
				ContainerID: containerID,
				Success:     false,
				Error:       "Failed to list containers: " + err.Error(),
			}
		}
		return results
	}

	// Create a map for quick container lookup
	containerMap := make(map[string]string) // ID -> Name
	for _, container := range containers {
		containerMap[container.ID] = container.Name
		// Also allow lookup by name
		containerMap[container.Name] = container.Name
	}

	// Perform operation on each container
	for i, containerID := range containerIDs {
		startTime := time.Now()

		// Check if container exists
		containerName, exists := containerMap[containerID]
		if !exists {
			results[i] = ContainerOperationResult{
				ContainerID: containerID,
				Success:     false,
				Error:       "Container not found",
				Duration:    time.Since(startTime).String(),
			}
			continue
		}

		// Perform the operation
		var err error
		switch operation {
		case "start":
			err = h.api.docker.StartContainer(containerID)
		case "stop":
			err = h.api.docker.StopContainer(containerID, 10) // 10 second timeout
		case "restart":
			err = h.api.docker.RestartContainer(containerID, 10) // 10 second timeout
		default:
			err = fmt.Errorf("unsupported operation: %s", operation)
		}

		// Record result
		results[i] = ContainerOperationResult{
			ContainerID:   containerID,
			ContainerName: containerName,
			Success:       err == nil,
			Duration:      time.Since(startTime).String(),
		}

		if err != nil {
			results[i].Error = err.Error()
		}
	}

	return results
}

// validateBulkOperationRequest validates a bulk operation request
func (h *HTTPServer) validateBulkOperationRequest(request *BulkOperationRequest) error {
	if len(request.ContainerIDs) == 0 {
		return fmt.Errorf("container_ids array cannot be empty")
	}

	if len(request.ContainerIDs) > 50 {
		return fmt.Errorf("maximum 50 containers allowed per bulk operation")
	}

	// Check for duplicate container IDs
	seen := make(map[string]bool)
	for _, containerID := range request.ContainerIDs {
		if containerID == "" {
			return fmt.Errorf("container ID cannot be empty")
		}
		if seen[containerID] {
			return fmt.Errorf("duplicate container ID: %s", containerID)
		}
		seen[containerID] = true
	}

	return nil
}

// Enhanced bulk operation with better error handling and validation
func (h *HTTPServer) handleEnhancedDockerBulkOperation(w http.ResponseWriter, r *http.Request, operation string) {
	// Parse request body
	var request BulkOperationRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid JSON request body: "+err.Error())
		return
	}

	// Validate request
	if err := h.validateBulkOperationRequest(&request); err != nil {
		h.writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Get request metadata
	requestID := h.getRequestIDFromContext(r)
	logger.Green("Starting enhanced bulk %s operation for %d containers [%s]", operation, len(request.ContainerIDs), requestID)

	// Perform bulk operation with enhanced error handling
	results := h.performEnhancedBulkDockerOperation(request.ContainerIDs, operation)

	// Calculate summary
	summary := BulkOperationSummary{
		Total: len(results),
	}
	for _, result := range results {
		if result.Success {
			summary.Succeeded++
		} else {
			summary.Failed++
		}
	}

	// Create response
	response := BulkOperationResponse{
		Operation: operation,
		Results:   results,
		Summary:   summary,
	}

	// Determine HTTP status code based on results
	statusCode := http.StatusOK
	if summary.Failed > 0 && summary.Succeeded == 0 {
		statusCode = http.StatusBadRequest // All operations failed
	} else if summary.Failed > 0 {
		statusCode = http.StatusMultiStatus // Partial success
	}

	// Use standardized response format
	h.writeStandardResponse(w, statusCode, response, nil)

	logger.Green("Enhanced bulk %s operation completed: %d succeeded, %d failed [%s]",
		operation, summary.Succeeded, summary.Failed, requestID)
}

// performEnhancedBulkDockerOperation performs bulk operations with enhanced error handling
func (h *HTTPServer) performEnhancedBulkDockerOperation(containerIDs []string, operation string) []ContainerOperationResult {
	results := make([]ContainerOperationResult, len(containerIDs))

	// Get container information first
	containers, err := h.api.docker.ListContainers(true)
	if err != nil {
		for i, containerID := range containerIDs {
			results[i] = ContainerOperationResult{
				ContainerID: containerID,
				Success:     false,
				Error:       "Failed to list containers: " + err.Error(),
			}
		}
		return results
	}

	// Create container lookup maps
	containerByID := make(map[string]interface{})
	containerByName := make(map[string]interface{})

	for _, container := range containers {
		containerByID[container.ID] = container
		containerByName[container.Name] = container
	}

	// Process each container
	for i, containerID := range containerIDs {
		startTime := time.Now()

		// Find container by ID or name
		var container interface{}
		var containerName string

		if c, exists := containerByID[containerID]; exists {
			container = c
			if containerData, ok := c.(map[string]interface{}); ok {
				if name, ok := containerData["name"].(string); ok {
					containerName = name
				}
			}
		} else if c, exists := containerByName[containerID]; exists {
			container = c
			containerName = containerID
		}

		if container == nil {
			results[i] = ContainerOperationResult{
				ContainerID: containerID,
				Success:     false,
				Error:       "Container not found",
				Duration:    time.Since(startTime).String(),
			}
			continue
		}

		// Perform operation
		var err error
		switch operation {
		case "start":
			err = h.api.docker.StartContainer(containerID)
		case "stop":
			err = h.api.docker.StopContainer(containerID, 10) // 10 second timeout
		case "restart":
			err = h.api.docker.RestartContainer(containerID, 10) // 10 second timeout
		default:
			err = fmt.Errorf("unsupported operation: %s", operation)
		}

		results[i] = ContainerOperationResult{
			ContainerID:   containerID,
			ContainerName: containerName,
			Success:       err == nil,
			Duration:      time.Since(startTime).String(),
		}

		if err != nil {
			results[i].Error = err.Error()
		}
	}

	return results
}
