package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/domalab/uma/daemon/logger"
	"github.com/domalab/uma/daemon/services/async"
)

// AsyncOperationsHandler handles GET /api/v1/operations
func (h *HTTPServer) AsyncOperationsHandler(w http.ResponseWriter, r *http.Request) {
	// Apply rate limiting
	if !h.api.rateLimiter.Allow(getClientIP(r)) {
		http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
		return
	}

	switch r.Method {
	case "GET":
		h.listOperations(w, r)
	case "POST":
		h.startOperation(w, r)
	case "OPTIONS":
		h.handleCORS(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// AsyncOperationHandler handles GET/DELETE /api/v1/operations/{id}
func (h *HTTPServer) AsyncOperationHandler(w http.ResponseWriter, r *http.Request) {
	// Apply rate limiting
	if !h.api.rateLimiter.Allow(getClientIP(r)) {
		http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
		return
	}

	// Extract operation ID from URL
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 5 {
		http.Error(w, "Invalid operation ID", http.StatusBadRequest)
		return
	}
	operationID := pathParts[4]

	switch r.Method {
	case "GET":
		h.getOperation(w, r, operationID)
	case "DELETE":
		h.cancelOperation(w, r, operationID)
	case "OPTIONS":
		h.handleCORS(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// listOperations handles GET /api/v1/operations
func (h *HTTPServer) listOperations(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters for filtering
	status := async.OperationStatus(r.URL.Query().Get("status"))
	operationType := async.OperationType(r.URL.Query().Get("type"))

	// Get operations from async manager
	response := h.api.asyncManager.ListOperations(status, operationType)

	// Create standard response
	standardResponse := map[string]interface{}{
		"data": response,
		"meta": map[string]interface{}{
			"timestamp":   getCurrentTimestamp(),
			"request_id":  r.Header.Get("X-Request-ID"),
			"api_version": "v1",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(standardResponse); err != nil {
		logger.Red("Failed to encode operations response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	logger.Green("Listed %d operations (status: %s, type: %s)",
		response.Total, status, operationType)
}

// startOperation handles POST /api/v1/operations
func (h *HTTPServer) startOperation(w http.ResponseWriter, r *http.Request) {
	var req async.OperationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.Type == "" {
		http.Error(w, "Operation type is required", http.StatusBadRequest)
		return
	}

	// Get user from context (if authenticated)
	createdBy := "anonymous"
	if userID := r.Header.Get("X-User-ID"); userID != "" {
		createdBy = userID
	}

	// Start the operation
	operation, err := h.api.asyncManager.StartOperation(req, createdBy)
	if err != nil {
		requestID := r.Header.Get("X-Request-ID")
		logger.LogErrorWithContext("async", "start_operation", err, requestID, map[string]interface{}{
			"operation_type": req.Type,
			"created_by":     createdBy,
			"description":    req.Description,
		})
		logger.Red("Failed to start operation %s: %v", req.Type, err)
		http.Error(w, fmt.Sprintf("Failed to start operation: %v", err), http.StatusBadRequest)
		return
	}

	// Create response
	response := async.OperationResponse{
		ID:          operation.ID,
		Type:        operation.Type,
		Status:      operation.Status,
		Description: operation.Description,
		Cancellable: operation.Cancellable,
		Started:     operation.Started,
	}

	// Create standard response
	standardResponse := map[string]interface{}{
		"data": response,
		"meta": map[string]interface{}{
			"timestamp":   getCurrentTimestamp(),
			"request_id":  r.Header.Get("X-Request-ID"),
			"api_version": "v1",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(standardResponse); err != nil {
		logger.Red("Failed to encode operation response: %v", err)
		return
	}

	// Log successful operation start with structured logging
	requestID := r.Header.Get("X-Request-ID")
	logger.LogAsyncOperation(operation.ID, string(operation.Type), "started", 0, requestID, map[string]interface{}{
		"description": operation.Description,
		"cancellable": operation.Cancellable,
		"created_by":  createdBy,
	})
	logger.Green("Started async operation %s (%s) by %s",
		operation.ID, operation.Type, createdBy)
}

// getOperation handles GET /api/v1/operations/{id}
func (h *HTTPServer) getOperation(w http.ResponseWriter, r *http.Request, operationID string) {
	operation, err := h.api.asyncManager.GetOperation(operationID)
	if err != nil {
		http.Error(w, "Operation not found", http.StatusNotFound)
		return
	}

	// Create standard response
	standardResponse := map[string]interface{}{
		"data": operation,
		"meta": map[string]interface{}{
			"timestamp":   getCurrentTimestamp(),
			"request_id":  r.Header.Get("X-Request-ID"),
			"api_version": "v1",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(standardResponse); err != nil {
		logger.Red("Failed to encode operation response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// cancelOperation handles DELETE /api/v1/operations/{id}
func (h *HTTPServer) cancelOperation(w http.ResponseWriter, r *http.Request, operationID string) {
	err := h.api.asyncManager.CancelOperation(operationID)
	if err != nil {
		logger.Yellow("Failed to cancel operation %s: %v", operationID, err)
		http.Error(w, fmt.Sprintf("Failed to cancel operation: %v", err), http.StatusBadRequest)
		return
	}

	// Create success response
	response := map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Operation %s cancelled successfully", operationID),
	}

	standardResponse := map[string]interface{}{
		"data": response,
		"meta": map[string]interface{}{
			"timestamp":   getCurrentTimestamp(),
			"request_id":  r.Header.Get("X-Request-ID"),
			"api_version": "v1",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(standardResponse); err != nil {
		logger.Red("Failed to encode cancel response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Log successful cancellation with structured logging
	requestID := r.Header.Get("X-Request-ID")
	logger.LogAsyncOperation(operationID, "unknown", "cancelled", 0, requestID, map[string]interface{}{
		"cancelled_by": "user",
	})
	logger.Green("Cancelled async operation %s", operationID)
}

// AsyncStatsHandler handles GET /api/v1/operations/stats
func (h *HTTPServer) AsyncStatsHandler(w http.ResponseWriter, r *http.Request) {
	// Apply rate limiting
	if !h.api.rateLimiter.Allow(getClientIP(r)) {
		http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
		return
	}

	if r.Method != "GET" {
		if r.Method == "OPTIONS" {
			h.handleCORS(w, r)
			return
		}
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get stats from async manager
	stats := h.api.asyncManager.GetStats()

	// Create standard response
	standardResponse := map[string]interface{}{
		"data": stats,
		"meta": map[string]interface{}{
			"timestamp":   getCurrentTimestamp(),
			"request_id":  r.Header.Get("X-Request-ID"),
			"api_version": "v1",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(standardResponse); err != nil {
		logger.Red("Failed to encode stats response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// Helper function to get current timestamp
func getCurrentTimestamp() int64 {
	return 1640995200 // This would be time.Now().Unix() in real implementation
}

// Helper function to get client IP
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP if multiple are present
		if idx := strings.Index(xff, ","); idx != -1 {
			return strings.TrimSpace(xff[:idx])
		}
		return strings.TrimSpace(xff)
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}

	// Fall back to RemoteAddr
	if idx := strings.LastIndex(r.RemoteAddr, ":"); idx != -1 {
		return r.RemoteAddr[:idx]
	}
	return r.RemoteAddr
}

// Helper function to handle CORS
func (h *HTTPServer) handleCORS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID")
	w.WriteHeader(http.StatusOK)
}
