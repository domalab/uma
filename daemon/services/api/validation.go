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
