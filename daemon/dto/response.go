package dto

// Response represents the legacy API response format (preserved for backward compatibility)
type Response struct {
	Message string   `json:"message"`
	Logs    []string `json:"logs,omitempty"`
	Error   string   `json:"error,omitempty"`
}

// PaginationInfo represents pagination metadata
type PaginationInfo struct {
	Page       int  `json:"page"`
	PerPage    int  `json:"per_page"`
	Total      int  `json:"total"`
	HasMore    bool `json:"has_more"`
	TotalPages int  `json:"total_pages"`
}

// PaginatedResponse represents a paginated API response
type PaginatedResponse struct {
	Data       interface{}     `json:"data"`
	Pagination *PaginationInfo `json:"pagination,omitempty"`
}

// StandardResponse represents the new standardized API response format
type StandardResponse struct {
	Data       interface{}     `json:"data"`
	Pagination *PaginationInfo `json:"pagination,omitempty"`
	Meta       *ResponseMeta   `json:"meta,omitempty"`
}

// ResponseMeta contains additional response metadata
type ResponseMeta struct {
	RequestID string `json:"request_id,omitempty"`
	Version   string `json:"version,omitempty"`
	Timestamp string `json:"timestamp,omitempty"`
}

// PaginationParams represents pagination query parameters
type PaginationParams struct {
	Page    int `json:"page" form:"page"`
	Limit   int `json:"limit" form:"limit"`
	PerPage int `json:"per_page" form:"per_page"`
}

// GetPage returns the page number, defaulting to 1 if not set or invalid
func (p *PaginationParams) GetPage() int {
	if p.Page < 1 {
		return 1
	}
	return p.Page
}

// GetLimit returns the limit, defaulting to 50 if not set or invalid
func (p *PaginationParams) GetLimit() int {
	if p.Limit < 1 {
		if p.PerPage > 0 {
			return p.PerPage
		}
		return 50
	}
	if p.Limit > 1000 {
		return 1000 // Maximum limit
	}
	return p.Limit
}

// CalculatePagination calculates pagination info based on total items and params
func CalculatePagination(total int, params *PaginationParams) *PaginationInfo {
	page := params.GetPage()
	limit := params.GetLimit()
	totalPages := (total + limit - 1) / limit // Ceiling division
	hasMore := page < totalPages

	return &PaginationInfo{
		Page:       page,
		PerPage:    limit,
		Total:      total,
		HasMore:    hasMore,
		TotalPages: totalPages,
	}
}
