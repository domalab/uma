package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/domalab/uma/daemon/dto"
	"github.com/domalab/uma/daemon/logger"
	"github.com/domalab/uma/daemon/services/api/types/responses"
)

// WriteJSON writes a JSON response
func WriteJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")

	// Try to marshal the data first to detect encoding errors
	jsonData, err := json.Marshal(data)
	if err != nil {
		logger.Yellow("Error encoding JSON response: %v", err)
		// If encoding fails, write an error response instead
		w.WriteHeader(http.StatusInternalServerError)
		errorResponse := dto.Response{
			Error:   "Failed to encode response",
			Message: "Internal Server Error",
		}
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	// If encoding succeeded, set the status code and write the data
	w.WriteHeader(status)
	w.Write(jsonData)
}

// WriteError writes an error response
func WriteError(w http.ResponseWriter, status int, message string) {
	// Ensure error field is always present, even if empty
	if message == "" {
		message = "Error occurred"
	}

	errorResponse := dto.Response{
		Error:   message,
		Message: http.StatusText(status),
	}
	WriteJSON(w, status, errorResponse)
}

// WriteStandardResponse writes a standardized API response using object pooling
func WriteStandardResponse(w http.ResponseWriter, status int, data interface{}, pagination *responses.PaginationInfo, requestID string, version string) {
	response := GetStandardResponse()
	defer PutStandardResponse(response)

	meta := GetResponseMeta()
	defer PutResponseMeta(meta)

	meta.RequestID = requestID
	meta.Version = version
	meta.Timestamp = time.Now().UTC()

	response.Data = data
	response.Pagination = pagination
	response.Meta = meta

	WriteJSON(w, status, response)
}

// WritePaginatedResponse writes a paginated API response using object pooling
func WritePaginatedResponse(w http.ResponseWriter, status int, data interface{}, total int, params *dto.PaginationParams, requestID string, version string) {
	pagination := dto.CalculatePagination(total, params)

	// Convert dto.PaginationInfo to responses.PaginationInfo using object pool
	responsePagination := GetPaginationInfo()
	defer PutPaginationInfo(responsePagination)

	responsePagination.Page = pagination.Page
	responsePagination.PageSize = pagination.PerPage
	responsePagination.TotalPages = pagination.TotalPages
	responsePagination.TotalItems = pagination.Total
	responsePagination.HasNext = pagination.HasMore
	responsePagination.HasPrev = pagination.Page > 1

	WriteStandardResponse(w, status, data, responsePagination, requestID, version)
}

// WriteVersionedResponse writes a response with version-specific formatting
func WriteVersionedResponse(w http.ResponseWriter, r *http.Request, status int, data interface{}, pagination *responses.PaginationInfo, requestID string, version string) {
	switch version {
	case "v1":
		// Current v1 format with standardized response structure
		WriteStandardResponse(w, status, data, pagination, requestID, version)
	default:
		// Future versions can have different response formats
		WriteStandardResponse(w, status, data, pagination, requestID, version)
	}
}

// WriteOperationResponse writes a generic operation response using object pooling
func WriteOperationResponse(w http.ResponseWriter, status int, success bool, message string, operationID string) {
	response := GetOperationResponse()
	defer PutOperationResponse(response)

	response.Success = success
	response.Message = message
	response.OperationID = operationID

	WriteJSON(w, status, response)
}

// WriteBulkOperationResponse writes a bulk operation response using object pooling
func WriteBulkOperationResponse(w http.ResponseWriter, status int, results []responses.BulkOperationResult) {
	total := len(results)
	succeeded := 0
	failed := 0

	for _, result := range results {
		if result.Success {
			succeeded++
		} else {
			failed++
		}
	}

	response := GetBulkOperationResponse()
	defer PutBulkOperationResponse(response)

	response.Success = failed == 0
	response.Message = "Bulk operation completed"
	response.Results = results
	response.Summary.Total = total
	response.Summary.Succeeded = succeeded
	response.Summary.Failed = failed

	WriteJSON(w, status, response)
}

// WriteHealthResponse writes a health check response
func WriteHealthResponse(w http.ResponseWriter, status string, version string, uptime int, checks map[string]responses.HealthCheck) {
	response := responses.HealthResponse{
		Status:    status,
		Version:   version,
		Uptime:    uptime,
		Timestamp: time.Now().UTC(),
		Checks:    checks,
	}

	var httpStatus int
	switch status {
	case "healthy":
		httpStatus = http.StatusOK
	case "degraded":
		httpStatus = http.StatusOK // Still return 200 for degraded
	case "unhealthy":
		httpStatus = http.StatusServiceUnavailable
	default:
		httpStatus = http.StatusInternalServerError
	}

	WriteJSON(w, httpStatus, response)
}

// GetRequestID gets the request ID from response header or generates one
func GetRequestID(w http.ResponseWriter) string {
	// Check if request ID was set in response headers by middleware
	if requestID := w.Header().Get("X-Request-ID"); requestID != "" {
		return requestID
	}

	// Generate a simple request ID as fallback
	return GenerateRequestID()
}

// GenerateRequestID generates a simple request ID
func GenerateRequestID() string {
	return fmt.Sprintf("req_%d", time.Now().UnixNano())
}

// ExtractURLParam extracts a URL parameter from the path (replacement for chi.URLParam)
// Example: ExtractURLParam(r, "/api/v1/vms/", "name") extracts "myvm" from "/api/v1/vms/myvm/action"
func ExtractURLParam(r *http.Request, prefix string, paramName string) string {
	path := r.URL.Path
	if !strings.HasPrefix(path, prefix) {
		return ""
	}

	// Remove the prefix to get the remaining path
	remaining := strings.TrimPrefix(path, prefix)
	parts := strings.Split(remaining, "/")

	if len(parts) == 0 || parts[0] == "" {
		return ""
	}

	// For now, we assume the first part is the parameter we want
	// This can be extended to support multiple named parameters if needed
	return parts[0]
}
