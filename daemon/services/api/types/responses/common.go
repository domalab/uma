package responses

import "time"

// Common response types used across multiple domains

// StandardResponse represents the standard API response format
type StandardResponse struct {
	Data       interface{}     `json:"data,omitempty"`
	Error      string          `json:"error,omitempty"`
	Message    string          `json:"message,omitempty"`
	Pagination *PaginationInfo `json:"pagination,omitempty"`
	Meta       *ResponseMeta   `json:"meta,omitempty"`
}

// ResponseMeta contains metadata about the response
type ResponseMeta struct {
	RequestID string    `json:"request_id,omitempty"`
	Version   string    `json:"version,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// PaginationInfo contains pagination information
type PaginationInfo struct {
	Page       int  `json:"page"`
	PageSize   int  `json:"page_size"`
	TotalPages int  `json:"total_pages"`
	TotalItems int  `json:"total_items"`
	HasNext    bool `json:"has_next"`
	HasPrev    bool `json:"has_prev"`
}

// OperationResponse represents a generic operation response
type OperationResponse struct {
	Success     bool   `json:"success"`
	Message     string `json:"message"`
	OperationID string `json:"operation_id,omitempty"`
}

// BulkOperationResponse represents a bulk operation response
type BulkOperationResponse struct {
	Success   bool                  `json:"success"`
	Message   string                `json:"message"`
	Operation string                `json:"operation,omitempty"`
	Results   []BulkOperationResult `json:"results"`
	Summary   BulkOperationSummary  `json:"summary"`
}

// BulkOperationResult represents the result of a single operation in a bulk request
type BulkOperationResult struct {
	ID      string `json:"id"`
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

// BulkOperationSummary provides a summary of bulk operation results
type BulkOperationSummary struct {
	Total     int `json:"total"`
	Succeeded int `json:"succeeded"`
	Failed    int `json:"failed"`
}

// HealthResponse represents a health check response
type HealthResponse struct {
	Status    string                 `json:"status"` // "healthy", "degraded", "unhealthy"
	Version   string                 `json:"version"`
	Uptime    int                    `json:"uptime"` // uptime in seconds
	Timestamp time.Time              `json:"timestamp"`
	Checks    map[string]HealthCheck `json:"checks,omitempty"`
}

// HealthCheck represents an individual health check
type HealthCheck struct {
	Status    string    `json:"status"` // "pass", "fail", "warn"
	Message   string    `json:"message,omitempty"`
	Timestamp time.Time `json:"timestamp"`
	Duration  string    `json:"duration,omitempty"`
}
