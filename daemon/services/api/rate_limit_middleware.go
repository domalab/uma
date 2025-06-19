package api

import (
	"net/http"
	"strings"

	"github.com/domalab/uma/daemon/dto"
	"github.com/domalab/uma/daemon/logger"
	"github.com/domalab/uma/daemon/services/auth"
)

// OperationRateLimitMiddleware creates middleware for operation-specific rate limiting
func (h *HTTPServer) OperationRateLimitMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get client IP
		clientIP := getClientIPFromRequest(r)

		// Determine operation type from request path and method
		operationType := getOperationTypeFromRequest(r)

		// Check operation-specific rate limit
		if !h.api.operationRateLimiter.Allow(clientIP, operationType) {
			// Create structured error response
			apiError := dto.NewAPIError(
				dto.ErrCodeRateLimitExceeded,
				"Operation-specific rate limit exceeded",
				http.StatusTooManyRequests,
			).WithDetails(map[string]interface{}{
				"operation_type": string(operationType),
				"client_ip":      clientIP,
				"limit":          h.api.operationRateLimiter.GetLimit(operationType),
			})

			h.writeAPIError(w, apiError)
			return
		}

		// Log rate limit check for expensive operations
		if isExpensiveOperation(operationType) {
			logger.Blue("Rate limit check passed for %s operation from %s", operationType, clientIP)
		}

		// Continue to next handler
		next(w, r)
	}
}

// getOperationTypeFromRequest determines the operation type from the request
func getOperationTypeFromRequest(r *http.Request) auth.OperationType {
	path := r.URL.Path
	method := r.Method

	// Handle method-specific operations
	switch {
	// Async operations - method-specific
	case path == "/api/v1/operations" && method == "POST":
		return auth.OpTypeAsyncCreate
	case path == "/api/v1/operations" && method == "GET":
		return auth.OpTypeAsyncList
	case strings.HasPrefix(path, "/api/v1/operations/") && method == "DELETE":
		return auth.OpTypeAsyncCancel

	// Docker operations - method-specific
	case strings.HasPrefix(path, "/api/v1/docker/containers/") && (method == "POST" || method == "PUT"):
		return auth.OpTypeDockerControl
	case path == "/api/v1/docker/bulk":
		return auth.OpTypeDockerBulk

	// VM operations - method-specific
	case strings.HasPrefix(path, "/api/v1/vms/") && (method == "POST" || method == "PUT"):
		return auth.OpTypeVMControl
	case path == "/api/v1/vms/bulk":
		return auth.OpTypeVMBulk

	// Array control operations
	case (path == "/api/v1/storage/array/start" || path == "/api/v1/storage/array/stop") && method == "POST":
		return auth.OpTypeArrayControl

	// Parity operations
	case strings.Contains(path, "/parity") && method == "POST":
		return auth.OpTypeParityCheck

	// System control operations
	case (path == "/api/v1/system/reboot" || path == "/api/v1/system/shutdown") && method == "POST":
		return auth.OpTypeSystemControl

	default:
		// Fall back to path-based detection
		return auth.GetOperationTypeFromPath(path)
	}
}

// isExpensiveOperation checks if an operation is considered expensive
func isExpensiveOperation(operationType auth.OperationType) bool {
	expensiveOps := []auth.OperationType{
		auth.OpTypeSMARTData,
		auth.OpTypeParityCheck,
		auth.OpTypeArrayControl,
		auth.OpTypeDockerBulk,
		auth.OpTypeVMBulk,
		auth.OpTypeSystemControl,
	}

	for _, expensiveOp := range expensiveOps {
		if operationType == expensiveOp {
			return true
		}
	}
	return false
}

// getClientIPFromRequest extracts client IP from request
func getClientIPFromRequest(r *http.Request) string {
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

// writeAPIError writes a structured API error response
func (h *HTTPServer) writeAPIError(w http.ResponseWriter, apiError *dto.APIError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(apiError.HTTPStatus)

	// Create standard error response
	errorResponse := map[string]interface{}{
		"error": map[string]interface{}{
			"code":    apiError.Code,
			"message": apiError.Message,
			"details": apiError.Details,
		},
		"meta": map[string]interface{}{
			"timestamp":   getCurrentTimestamp(),
			"api_version": "v1",
		},
	}

	h.WriteJSON(w, apiError.HTTPStatus, errorResponse)
}

// RateLimitStatsHandler handles GET /api/v1/rate-limits/stats
func (h *HTTPServer) RateLimitStatsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		if r.Method == "OPTIONS" {
			h.handleCORS(w, r)
			return
		}
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get stats from both rate limiters
	generalStats := map[string]interface{}{
		"type": "general",
		// Add general rate limiter stats here if available
	}

	operationStats := h.api.operationRateLimiter.GetStats()
	operationStats["type"] = "operation_specific"

	response := map[string]interface{}{
		"general_rate_limiter":   generalStats,
		"operation_rate_limiter": operationStats,
	}

	// Create standard response
	standardResponse := map[string]interface{}{
		"data": response,
		"meta": map[string]interface{}{
			"timestamp":   getCurrentTimestamp(),
			"api_version": "v1",
		},
	}

	h.WriteJSON(w, http.StatusOK, standardResponse)
}

// RateLimitConfigHandler handles GET/PUT /api/v1/rate-limits/config
func (h *HTTPServer) RateLimitConfigHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		h.getRateLimitConfig(w, r)
	case "PUT":
		h.updateRateLimitConfig(w, r)
	case "OPTIONS":
		h.handleCORS(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// getRateLimitConfig returns current rate limit configuration
func (h *HTTPServer) getRateLimitConfig(w http.ResponseWriter, r *http.Request) {
	// Get current limits for all operation types
	operationTypes := []auth.OperationType{
		auth.OpTypeGeneral,
		auth.OpTypeHealthCheck,
		auth.OpTypeSMARTData,
		auth.OpTypeParityCheck,
		auth.OpTypeArrayControl,
		auth.OpTypeDiskInfo,
		auth.OpTypeDockerList,
		auth.OpTypeDockerControl,
		auth.OpTypeDockerBulk,
		auth.OpTypeVMList,
		auth.OpTypeVMControl,
		auth.OpTypeVMBulk,
		auth.OpTypeSystemInfo,
		auth.OpTypeSystemControl,
		auth.OpTypeSensorData,
		auth.OpTypeAsyncCreate,
		auth.OpTypeAsyncList,
		auth.OpTypeAsyncCancel,
	}

	config := make(map[string]interface{})
	for _, opType := range operationTypes {
		limit := h.api.operationRateLimiter.GetLimit(opType)
		config[string(opType)] = map[string]interface{}{
			"requests": limit.Requests,
			"window":   limit.Window.String(),
		}
	}

	// Create standard response
	standardResponse := map[string]interface{}{
		"data": config,
		"meta": map[string]interface{}{
			"timestamp":   getCurrentTimestamp(),
			"api_version": "v1",
		},
	}

	h.WriteJSON(w, http.StatusOK, standardResponse)
}

// updateRateLimitConfig updates rate limit configuration
func (h *HTTPServer) updateRateLimitConfig(w http.ResponseWriter, r *http.Request) {
	// This would require admin authentication in a real implementation
	// For now, just return a placeholder response

	response := map[string]interface{}{
		"message": "Rate limit configuration update not implemented yet",
		"note":    "This endpoint would require admin authentication",
	}

	standardResponse := map[string]interface{}{
		"data": response,
		"meta": map[string]interface{}{
			"timestamp":   getCurrentTimestamp(),
			"api_version": "v1",
		},
	}

	h.WriteJSON(w, http.StatusNotImplemented, standardResponse)
}
