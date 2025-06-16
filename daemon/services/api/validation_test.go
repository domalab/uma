package api

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateBulkRequest(t *testing.T) {
	// Create a mock HTTP server for testing
	httpServer := &HTTPServer{}

	tests := []struct {
		name        string
		request     BulkOperationRequest
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid request with single container",
			request: BulkOperationRequest{
				ContainerIDs: []string{"plex"},
			},
			expectError: false,
		},
		{
			name: "Valid request with multiple containers",
			request: BulkOperationRequest{
				ContainerIDs: []string{"plex", "nginx", "sonarr"},
			},
			expectError: false,
		},
		{
			name: "Empty container IDs array",
			request: BulkOperationRequest{
				ContainerIDs: []string{},
			},
			expectError: true,
			errorMsg:    "at least 1 container ID is required",
		},
		{
			name: "Nil container IDs",
			request: BulkOperationRequest{
				ContainerIDs: nil,
			},
			expectError: true,
			errorMsg:    "container_ids field is required",
		},
		{
			name: "Empty string in container IDs",
			request: BulkOperationRequest{
				ContainerIDs: []string{"plex", "", "nginx"},
			},
			expectError: true,
			errorMsg:    "container ID cannot be empty",
		},
		{
			name: "Duplicate container IDs",
			request: BulkOperationRequest{
				ContainerIDs: []string{"plex", "nginx", "plex"},
			},
			expectError: true,
			errorMsg:    "duplicate container ID: plex",
		},
		{
			name: "Container ID with leading/trailing whitespace",
			request: BulkOperationRequest{
				ContainerIDs: []string{" plex ", "nginx"},
			},
			expectError: true,
			errorMsg:    "container ID contains leading/trailing whitespace",
		},
		{
			name: "Container ID with spaces",
			request: BulkOperationRequest{
				ContainerIDs: []string{"plex server", "nginx"},
			},
			expectError: true,
			errorMsg:    "container ID contains spaces",
		},
		{
			name: "Too many containers (over 50)",
			request: BulkOperationRequest{
				ContainerIDs: generateContainerIDs(51),
			},
			expectError: true,
			errorMsg:    "maximum 50 containers allowed per bulk operation",
		},
		{
			name: "Maximum allowed containers (50)",
			request: BulkOperationRequest{
				ContainerIDs: generateContainerIDs(50),
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := httpServer.validateBulkRequest(&tt.request)

			if tt.expectError {
				require.Error(t, err, "Expected validation error but got none")
				assert.Contains(t, err.Error(), tt.errorMsg, "Error message should contain expected text")
			} else {
				assert.NoError(t, err, "Expected no validation error but got: %v", err)
			}
		})
	}
}

func TestValidatePaginationParams(t *testing.T) {
	httpServer := &HTTPServer{}

	tests := []struct {
		name        string
		page        int
		limit       int
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Valid pagination parameters",
			page:        1,
			limit:       10,
			expectError: false,
		},
		{
			name:        "Valid pagination with maximum limit",
			page:        5,
			limit:       1000,
			expectError: false,
		},
		{
			name:        "Invalid page (zero)",
			page:        0,
			limit:       10,
			expectError: true,
			errorMsg:    "page must be >= 1",
		},
		{
			name:        "Invalid page (negative)",
			page:        -1,
			limit:       10,
			expectError: true,
			errorMsg:    "page must be >= 1",
		},
		{
			name:        "Invalid limit (zero)",
			page:        1,
			limit:       0,
			expectError: true,
			errorMsg:    "limit must be >= 1",
		},
		{
			name:        "Invalid limit (negative)",
			page:        1,
			limit:       -5,
			expectError: true,
			errorMsg:    "limit must be >= 1",
		},
		{
			name:        "Invalid limit (too large)",
			page:        1,
			limit:       1001,
			expectError: true,
			errorMsg:    "limit must be <= 1000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := httpServer.validatePaginationParams(tt.page, tt.limit)

			if tt.expectError {
				require.Error(t, err, "Expected validation error but got none")
				assert.Contains(t, err.Error(), tt.errorMsg, "Error message should contain expected text")
			} else {
				assert.NoError(t, err, "Expected no validation error but got: %v", err)
			}
		})
	}
}

func TestValidateRequestID(t *testing.T) {
	httpServer := &HTTPServer{}

	tests := []struct {
		name        string
		requestID   string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Valid request ID",
			requestID:   "test-request-123",
			expectError: false,
		},
		{
			name:        "Valid request ID with UUID format",
			requestID:   "550e8400-e29b-41d4-a716-446655440000",
			expectError: false,
		},
		{
			name:        "Empty request ID (valid)",
			requestID:   "",
			expectError: false,
		},
		{
			name:        "Request ID too long",
			requestID:   string(make([]byte, 256)), // 256 characters
			expectError: true,
			errorMsg:    "request ID too long",
		},
		{
			name:        "Request ID with invalid characters",
			requestID:   "test\x00request",
			expectError: true,
			errorMsg:    "request ID contains invalid characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := httpServer.validateRequestID(tt.requestID)

			if tt.expectError {
				require.Error(t, err, "Expected validation error but got none")
				assert.Contains(t, err.Error(), tt.errorMsg, "Error message should contain expected text")
			} else {
				assert.NoError(t, err, "Expected no validation error but got: %v", err)
			}
		})
	}
}

func TestValidateAPIVersion(t *testing.T) {
	httpServer := &HTTPServer{}

	tests := []struct {
		name        string
		version     string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Valid API version v1",
			version:     "v1",
			expectError: false,
		},
		{
			name:        "Invalid API version v2",
			version:     "v2",
			expectError: true,
			errorMsg:    "unsupported API version: v2",
		},
		{
			name:        "Invalid API version empty",
			version:     "",
			expectError: true,
			errorMsg:    "unsupported API version:",
		},
		{
			name:        "Invalid API version format",
			version:     "1.0",
			expectError: true,
			errorMsg:    "unsupported API version: 1.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := httpServer.validateAPIVersion(tt.version)

			if tt.expectError {
				require.Error(t, err, "Expected validation error but got none")
				assert.Contains(t, err.Error(), tt.errorMsg, "Error message should contain expected text")
			} else {
				assert.NoError(t, err, "Expected no validation error but got: %v", err)
			}
		})
	}
}

// Helper function to generate container IDs for testing
func generateContainerIDs(count int) []string {
	ids := make([]string, count)
	for i := 0; i < count; i++ {
		ids[i] = fmt.Sprintf("container-%d", i+1)
	}
	return ids
}

// Benchmark tests for validation performance
func BenchmarkValidateBulkRequest(b *testing.B) {
	httpServer := &HTTPServer{}
	request := BulkOperationRequest{
		ContainerIDs: []string{"plex", "nginx", "sonarr", "radarr", "lidarr"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = httpServer.validateBulkRequest(&request)
	}
}

func BenchmarkValidateBulkRequestLarge(b *testing.B) {
	httpServer := &HTTPServer{}
	request := BulkOperationRequest{
		ContainerIDs: generateContainerIDs(50), // Maximum allowed
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = httpServer.validateBulkRequest(&request)
	}
}
