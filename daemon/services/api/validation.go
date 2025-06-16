package api

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

// BulkOperationRequestValidated represents a validated bulk operation request
type BulkOperationRequestValidated struct {
	ContainerIDs []string `json:"container_ids" validate:"required,min=1,max=50,dive,required,min=1"`
}

// validateBulkRequest validates bulk operation requests with comprehensive checks
func (h *HTTPServer) validateBulkRequest(req *BulkOperationRequest) error {
	// Convert to validated struct for validation
	validatedReq := BulkOperationRequestValidated{
		ContainerIDs: req.ContainerIDs,
	}

	// Perform struct validation
	if err := validate.Struct(validatedReq); err != nil {
		// Convert validation errors to user-friendly messages
		var errorMessages []string
		for _, err := range err.(validator.ValidationErrors) {
			switch err.Tag() {
			case "required":
				if err.Field() == "ContainerIDs" {
					errorMessages = append(errorMessages, "container_ids field is required")
				} else {
					errorMessages = append(errorMessages, "container ID cannot be empty")
				}
			case "min":
				if err.Field() == "ContainerIDs" {
					errorMessages = append(errorMessages, "at least 1 container ID is required")
				} else {
					errorMessages = append(errorMessages, "container ID cannot be empty")
				}
			case "max":
				errorMessages = append(errorMessages, "maximum 50 containers allowed per bulk operation")
			}
		}
		return fmt.Errorf("validation failed: %s", strings.Join(errorMessages, ", "))
	}

	// Custom validation for duplicate IDs
	seen := make(map[string]bool)
	for _, id := range req.ContainerIDs {
		if seen[id] {
			return fmt.Errorf("duplicate container ID: %s", id)
		}
		seen[id] = true
	}

	// Custom validation for container ID format (basic check)
	for _, id := range req.ContainerIDs {
		if strings.TrimSpace(id) != id {
			return fmt.Errorf("container ID contains leading/trailing whitespace: '%s'", id)
		}
		if strings.Contains(id, " ") {
			return fmt.Errorf("container ID contains spaces: '%s'", id)
		}
	}

	return nil
}

// validateHealthCheckRequest validates health check requests (future use)
func (h *HTTPServer) validateHealthCheckRequest() error {
	// Placeholder for future health check validation
	return nil
}

// validatePaginationParams validates pagination parameters
func (h *HTTPServer) validatePaginationParams(page, limit int) error {
	if page < 1 {
		return fmt.Errorf("page must be >= 1, got %d", page)
	}
	if limit < 1 {
		return fmt.Errorf("limit must be >= 1, got %d", limit)
	}
	if limit > 1000 {
		return fmt.Errorf("limit must be <= 1000, got %d", limit)
	}
	return nil
}

// validateRequestID validates request ID format
func (h *HTTPServer) validateRequestID(requestID string) error {
	if len(requestID) > 255 {
		return fmt.Errorf("request ID too long (max 255 characters)")
	}
	
	// Check for invalid characters (basic validation)
	for _, char := range requestID {
		if char < 32 || char > 126 {
			return fmt.Errorf("request ID contains invalid characters")
		}
	}
	
	return nil
}

// Enhanced validation for API versioning
func (h *HTTPServer) validateAPIVersion(version string) error {
	supportedVersions := []string{"v1"}
	
	for _, supported := range supportedVersions {
		if version == supported {
			return nil
		}
	}
	
	return fmt.Errorf("unsupported API version: %s (supported: %s)", version, strings.Join(supportedVersions, ", "))
}
