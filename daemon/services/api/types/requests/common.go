package requests

// Common request types used across multiple domains

// PaginationRequest represents pagination parameters
type PaginationRequest struct {
	Page     int `json:"page,omitempty"`
	PageSize int `json:"page_size,omitempty"`
	Offset   int `json:"offset,omitempty"`
	Limit    int `json:"limit,omitempty"`
}

// FilterRequest represents common filtering parameters
type FilterRequest struct {
	Search   string            `json:"search,omitempty"`
	Filters  map[string]string `json:"filters,omitempty"`
	SortBy   string            `json:"sort_by,omitempty"`
	SortDesc bool              `json:"sort_desc,omitempty"`
}

// BulkOperationRequest represents a bulk operation request
type BulkOperationRequest struct {
	IDs          []string `json:"ids,omitempty"`
	ContainerIDs []string `json:"container_ids,omitempty"`
	Operation    string   `json:"operation,omitempty"`
	Force        bool     `json:"force,omitempty"`
}
