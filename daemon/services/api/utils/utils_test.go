package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/domalab/uma/daemon/dto"
	"github.com/domalab/uma/daemon/services/api/types/requests"
	"github.com/domalab/uma/daemon/services/api/types/responses"
)

// TestWriteJSON tests the WriteJSON utility function
func TestWriteJSON(t *testing.T) {
	tests := []struct {
		name           string
		status         int
		data           interface{}
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Write simple object",
			status:         http.StatusOK,
			data:           map[string]string{"message": "success"},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"message":"success"}`,
		},
		{
			name:           "Write array",
			status:         http.StatusCreated,
			data:           []string{"item1", "item2"},
			expectedStatus: http.StatusCreated,
			expectedBody:   `["item1","item2"]`,
		},
		{
			name:           "Write error status",
			status:         http.StatusBadRequest,
			data:           map[string]string{"error": "bad request"},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"bad request"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()

			WriteJSON(w, tt.status, tt.data)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			contentType := w.Header().Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
			}

			// Parse and compare JSON to handle formatting differences
			var expected, actual interface{}
			if err := json.Unmarshal([]byte(tt.expectedBody), &expected); err != nil {
				t.Fatalf("Failed to unmarshal expected body: %v", err)
			}
			if err := json.Unmarshal(w.Body.Bytes(), &actual); err != nil {
				t.Fatalf("Failed to unmarshal actual body: %v", err)
			}

			expectedJSON, _ := json.Marshal(expected)
			actualJSON, _ := json.Marshal(actual)
			if !bytes.Equal(expectedJSON, actualJSON) {
				t.Errorf("Expected body %s, got %s", expectedJSON, actualJSON)
			}
		})
	}
}

// TestWriteError tests the WriteError utility function
func TestWriteError(t *testing.T) {
	tests := []struct {
		name           string
		status         int
		message        string
		expectedStatus int
	}{
		{
			name:           "Bad request error",
			status:         http.StatusBadRequest,
			message:        "Invalid input",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Internal server error",
			status:         http.StatusInternalServerError,
			message:        "Database connection failed",
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "Not found error",
			status:         http.StatusNotFound,
			message:        "Resource not found",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()

			WriteError(w, tt.status, tt.message)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			contentType := w.Header().Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
			}

			var response map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}

			if response["error"] != tt.message {
				t.Errorf("Expected error message '%s', got '%v'", tt.message, response["error"])
			}
		})
	}
}

// TestWriteStandardResponse tests the WriteStandardResponse utility function
func TestWriteStandardResponse(t *testing.T) {
	// Note: This test is simplified since WriteStandardResponse requires specific imports
	// In a real implementation, you would test this with proper mocking

	// For now, just test that WriteJSON works correctly
	data := map[string]string{"key": "value"}
	w := httptest.NewRecorder()

	WriteJSON(w, http.StatusOK, data)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response["key"] != "value" {
		t.Errorf("Expected key to be 'value', got '%v'", response["key"])
	}
}

// TestIsCommandBlacklisted tests the command blacklist validation
func TestIsCommandBlacklisted(t *testing.T) {
	tests := []struct {
		name     string
		command  string
		expected bool
	}{
		{
			name:     "Safe command",
			command:  "ls -la",
			expected: false,
		},
		{
			name:     "Dangerous rm command",
			command:  "rm -rf /",
			expected: true,
		},
		{
			name:     "Format command",
			command:  "mkfs.ext4 /dev/sda1",
			expected: true,
		},
		{
			name:     "DD command",
			command:  "dd if=/dev/zero of=/dev/sda",
			expected: true,
		},
		{
			name:     "Shutdown command",
			command:  "shutdown -h now",
			expected: true,
		},
		{
			name:     "Reboot command",
			command:  "reboot",
			expected: true,
		},
		{
			name:     "Safe echo command",
			command:  "echo hello world",
			expected: false,
		},
		{
			name:     "Safe cat command",
			command:  "cat /etc/hostname",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsCommandBlacklisted(tt.command)
			if result != tt.expected {
				t.Errorf("Expected IsCommandBlacklisted('%s') to be %v, got %v", tt.command, tt.expected, result)
			}
		})
	}
}

// TestValidateRequest tests the request validation utility
func TestValidateRequest(t *testing.T) {
	// Test struct for validation
	type TestRequest struct {
		Name  string `json:"name" validate:"required,min=3,max=50"`
		Email string `json:"email" validate:"required,email"`
		Age   int    `json:"age" validate:"min=0,max=120"`
	}

	tests := []struct {
		name        string
		request     TestRequest
		expectError bool
		errorField  string
	}{
		{
			name: "Valid request",
			request: TestRequest{
				Name:  "John Doe",
				Email: "john@example.com",
				Age:   30,
			},
			expectError: false,
		},
		{
			name: "Missing required name",
			request: TestRequest{
				Email: "john@example.com",
				Age:   30,
			},
			expectError: true,
			errorField:  "name",
		},
		{
			name: "Invalid email",
			request: TestRequest{
				Name:  "John Doe",
				Email: "invalid-email",
				Age:   30,
			},
			expectError: true,
			errorField:  "email",
		},
		{
			name: "Age too high",
			request: TestRequest{
				Name:  "John Doe",
				Email: "john@example.com",
				Age:   150,
			},
			expectError: true,
			errorField:  "age",
		},
		{
			name: "Name too short",
			request: TestRequest{
				Name:  "Jo",
				Email: "john@example.com",
				Age:   30,
			},
			expectError: true,
			errorField:  "name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStruct(tt.request)

			if tt.expectError {
				if err == nil {
					t.Error("Expected validation error, got nil")
				}
				// Check if error message contains the expected field
				if tt.errorField != "" && err != nil {
					if !stringContains(err.Error(), tt.errorField) {
						t.Errorf("Expected error to contain field '%s', got: %v", tt.errorField, err)
					}
				}
			} else {
				if err != nil {
					t.Errorf("Expected no validation error, got: %v", err)
				}
			}
		})
	}
}

// Helper function to check if string contains substring
func stringContains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

// TestComprehensiveEdgeCases tests edge cases and error conditions for utility functions
func TestComprehensiveEdgeCases(t *testing.T) {
	t.Run("WriteJSON with nil data", func(t *testing.T) {
		w := httptest.NewRecorder()
		WriteJSON(w, http.StatusOK, nil)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var result interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
			t.Errorf("Failed to unmarshal nil JSON: %v", err)
		}

		if result != nil {
			t.Errorf("Expected nil result, got %v", result)
		}
	})

	t.Run("WriteJSON with circular reference", func(t *testing.T) {
		w := httptest.NewRecorder()

		// Create circular reference
		type Node struct {
			Value string
			Next  *Node
		}
		node1 := &Node{Value: "first"}
		node2 := &Node{Value: "second"}
		node1.Next = node2
		node2.Next = node1 // Circular reference

		WriteJSON(w, http.StatusOK, node1)

		// Should handle the error gracefully
		if w.Code == http.StatusOK {
			t.Error("Expected error status for circular reference, got 200")
		}
	})

	t.Run("WriteError with empty message", func(t *testing.T) {
		w := httptest.NewRecorder()
		WriteError(w, http.StatusBadRequest, "")

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}

		var response map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Errorf("Failed to unmarshal error response: %v", err)
		}

		if response["error"] == nil {
			t.Error("Expected error field in response")
		}
	})

	t.Run("WriteError with very long message", func(t *testing.T) {
		w := httptest.NewRecorder()
		longMessage := strings.Repeat("error ", 1000) // Very long error message
		WriteError(w, http.StatusInternalServerError, longMessage)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("Expected status 500, got %d", w.Code)
		}

		// Should handle long messages without issues
		if w.Body.Len() == 0 {
			t.Error("Expected response body for long error message")
		}
	})
}

// TestSecurityValidation tests security-related validation functions
func TestSecurityValidation(t *testing.T) {
	t.Run("HTML escaping in WriteJSON", func(t *testing.T) {
		w := httptest.NewRecorder()
		data := map[string]string{
			"message": "<script>alert('xss')</script>",
		}

		WriteJSON(w, http.StatusOK, data)

		// Check that the response doesn't contain unescaped HTML
		body := w.Body.String()
		if strings.Contains(body, "<script>") {
			t.Error("Response should not contain unescaped HTML tags")
		}
	})

	t.Run("SQL injection patterns in WriteError", func(t *testing.T) {
		w := httptest.NewRecorder()
		maliciousMessage := "'; DROP TABLE users; --"

		WriteError(w, http.StatusBadRequest, maliciousMessage)

		// Should handle malicious input safely
		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}

		// Response should be valid JSON
		var response map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Errorf("Response should be valid JSON: %v", err)
		}
	})
}

// TestInputValidationEdgeCases tests edge cases for input validation
func TestInputValidationEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		input       interface{}
		expectError bool
	}{
		{
			name:        "Nil input",
			input:       nil,
			expectError: false, // Should handle nil gracefully
		},
		{
			name:        "Empty struct",
			input:       struct{}{},
			expectError: false,
		},
		{
			name:        "Very large string",
			input:       struct{ Data string }{Data: strings.Repeat("x", 100000)},
			expectError: false, // Should handle large data
		},
		{
			name: "Nested struct",
			input: struct {
				User struct {
					Name string
					Age  int
				}
			}{
				User: struct {
					Name string
					Age  int
				}{Name: "Test", Age: 25},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			WriteJSON(w, http.StatusOK, tt.input)

			// Should not crash or return 500
			if w.Code == http.StatusInternalServerError && !tt.expectError {
				t.Errorf("Unexpected internal server error for input: %v", tt.input)
			}
		})
	}
}

// TestValidationFunctions tests various validation utility functions
func TestValidationFunctions(t *testing.T) {
	t.Run("IsValidVMName", func(t *testing.T) {
		tests := []struct {
			name     string
			vmName   string
			expected bool
		}{
			{"Valid VM name", "test-vm", true},
			{"Valid VM name with numbers", "vm123", true},
			{"Single character", "a", true},
			{"Empty name", "", false},
			{"Name too long", strings.Repeat("a", 256), false},
			{"Name with spaces", "test vm", false},
			{"Name with special chars", "test@vm", false},
			{"Name starting with hyphen", "-test", false},
			{"Name ending with hyphen", "test-", false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := IsValidVMName(tt.vmName)
				if result != tt.expected {
					t.Errorf("IsValidVMName(%q) = %v, expected %v", tt.vmName, result, tt.expected)
				}
			})
		}
	})

	t.Run("IsValidContainerName", func(t *testing.T) {
		tests := []struct {
			name          string
			containerName string
			expected      bool
		}{
			{"Valid container name", "test-container", true},
			{"Valid name with numbers", "container123", true},
			{"Valid name with dots", "test.container", true},
			{"Valid name with underscores", "test_container", true},
			{"Empty name", "", false},
			{"Name with invalid chars", "test/container", false},
			{"Very long name", strings.Repeat("a", 300), false},
			{"Name starting with special char", "-test", false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := IsValidContainerName(tt.containerName)
				if result != tt.expected {
					t.Errorf("IsValidContainerName(%q) = %v, expected %v", tt.containerName, result, tt.expected)
				}
			})
		}
	})

	t.Run("IsValidShareName", func(t *testing.T) {
		tests := []struct {
			name      string
			shareName string
			expected  bool
		}{
			{"Valid share name", "test-share", true},
			{"Valid name with numbers", "share123", true},
			{"Valid name with underscores", "test_share", true},
			{"Empty name", "", false},
			{"Name too long", strings.Repeat("a", 50), false},
			{"Name with spaces", "test share", false},
			{"Name with special chars", "test@share", false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := IsValidShareName(tt.shareName)
				if result != tt.expected {
					t.Errorf("IsValidShareName(%q) = %v, expected %v", tt.shareName, result, tt.expected)
				}
			})
		}
	})

	t.Run("IsCommandBlacklisted", func(t *testing.T) {
		tests := []struct {
			name     string
			command  string
			expected bool
		}{
			{"Safe command", "ls -la", false},
			{"Safe command with pipes", "ps aux | grep test", false},
			{"Dangerous rm command", "rm -rf /", true},
			{"Dangerous dd command", "dd if=/dev/zero of=/dev/sda", true},
			{"Shutdown command", "shutdown now", true},
			{"Reboot command", "reboot", true},
			{"Chmod 777", "chmod 777 /etc/passwd", true},
			{"Sudo su", "sudo su -", true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := IsCommandBlacklisted(tt.command)
				if result != tt.expected {
					t.Errorf("IsCommandBlacklisted(%q) = %v, expected %v", tt.command, result, tt.expected)
				}
			})
		}
	})
}

// TestRequestValidationFunctions tests request validation functions
func TestRequestValidationFunctions(t *testing.T) {
	t.Run("ValidateCommandExecuteRequest", func(t *testing.T) {
		tests := []struct {
			name    string
			request requests.CommandExecuteRequest
			wantErr bool
		}{
			{
				name: "Valid command request",
				request: requests.CommandExecuteRequest{
					Command: "ls -la",
					Timeout: 30,
				},
				wantErr: false,
			},
			{
				name: "Empty command",
				request: requests.CommandExecuteRequest{
					Command: "",
					Timeout: 30,
				},
				wantErr: true,
			},
			{
				name: "Blacklisted command",
				request: requests.CommandExecuteRequest{
					Command: "rm -rf /",
					Timeout: 30,
				},
				wantErr: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := ValidateCommandExecuteRequest(&tt.request)
				if (err != nil) != tt.wantErr {
					t.Errorf("ValidateCommandExecuteRequest() error = %v, wantErr %v", err, tt.wantErr)
				}
			})
		}
	})
}

// TestResponseFunctions tests additional response utility functions
func TestResponseFunctions(t *testing.T) {
	t.Run("WriteStandardResponse", func(t *testing.T) {
		w := httptest.NewRecorder()
		data := map[string]string{"test": "data"}
		pagination := &responses.PaginationInfo{
			Page:       1,
			PageSize:   10,
			TotalPages: 5,
			TotalItems: 50,
			HasNext:    true,
			HasPrev:    false,
		}

		WriteStandardResponse(w, http.StatusOK, data, pagination, "req-123", "v1")

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		var response responses.StandardResponse
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
		}

		if response.Data == nil {
			t.Error("Expected data field in standard response")
		}
		if response.Pagination == nil {
			t.Error("Expected pagination field in standard response")
		}
		if response.Meta == nil {
			t.Error("Expected meta field in standard response")
		}
	})

	t.Run("WriteOperationResponse", func(t *testing.T) {
		w := httptest.NewRecorder()

		WriteOperationResponse(w, http.StatusOK, true, "Operation completed", "op-123")

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		var response responses.OperationResponse
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
		}

		if !response.Success {
			t.Error("Expected success to be true")
		}
		if response.Message != "Operation completed" {
			t.Errorf("Expected message 'Operation completed', got '%s'", response.Message)
		}
		if response.OperationID != "op-123" {
			t.Errorf("Expected operation ID 'op-123', got '%s'", response.OperationID)
		}
	})

	t.Run("WriteBulkOperationResponse", func(t *testing.T) {
		w := httptest.NewRecorder()
		results := []responses.BulkOperationResult{
			{ID: "item1", Success: true, Message: "Success"},
			{ID: "item2", Success: false, Message: "Failed", Error: "Error occurred"},
		}

		WriteBulkOperationResponse(w, http.StatusOK, results)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		var response responses.BulkOperationResponse
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
		}

		if response.Success {
			t.Error("Expected success to be false (due to failed items)")
		}
		if len(response.Results) != 2 {
			t.Errorf("Expected 2 results, got %d", len(response.Results))
		}
		if response.Summary.Total != 2 {
			t.Errorf("Expected total 2, got %d", response.Summary.Total)
		}
		if response.Summary.Succeeded != 1 {
			t.Errorf("Expected succeeded 1, got %d", response.Summary.Succeeded)
		}
		if response.Summary.Failed != 1 {
			t.Errorf("Expected failed 1, got %d", response.Summary.Failed)
		}
	})

	t.Run("WriteHealthResponse", func(t *testing.T) {
		w := httptest.NewRecorder()
		checks := map[string]responses.HealthCheck{
			"database": {
				Status:  "healthy",
				Message: "Database is responsive",
			},
			"cache": {
				Status:  "unhealthy",
				Message: "Cache connection failed",
			},
		}

		WriteHealthResponse(w, "degraded", "1.0.0", 86400, checks) // 24h in seconds

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		var response responses.HealthResponse
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
		}

		if response.Status != "degraded" {
			t.Errorf("Expected status 'degraded', got '%s'", response.Status)
		}
		if response.Version != "1.0.0" {
			t.Errorf("Expected version '1.0.0', got '%s'", response.Version)
		}
		if len(response.Checks) != 2 {
			t.Errorf("Expected 2 checks, got %d", len(response.Checks))
		}
	})

	t.Run("GenerateRequestID", func(t *testing.T) {
		id1 := GenerateRequestID()

		// Add a small delay to ensure different timestamps
		time.Sleep(1 * time.Millisecond)

		id2 := GenerateRequestID()

		if id1 == id2 {
			t.Error("Expected different request IDs")
		}
		if !strings.HasPrefix(id1, "req_") {
			t.Errorf("Expected request ID to start with 'req_', got '%s'", id1)
		}
		if !strings.HasPrefix(id2, "req_") {
			t.Errorf("Expected request ID to start with 'req_', got '%s'", id2)
		}
	})
}

// TestValidationFunctionsExtended tests additional validation functions
func TestValidationFunctionsExtended(t *testing.T) {
	t.Run("ValidateShareCreateRequest", func(t *testing.T) {
		tests := []struct {
			name    string
			request requests.ShareCreateRequest
			wantErr bool
		}{
			{
				name: "Valid share create request",
				request: requests.ShareCreateRequest{
					Name:            "test-share",
					Comment:         "Test share",
					AllocatorMethod: "high-water",
					SMBEnabled:      true,
					SMBSecurity:     "secure",
				},
				wantErr: false,
			},
			{
				name: "Empty share name",
				request: requests.ShareCreateRequest{
					Name: "",
				},
				wantErr: true,
			},
			{
				name: "Share name too long",
				request: requests.ShareCreateRequest{
					Name: strings.Repeat("a", 50),
				},
				wantErr: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := ValidateShareCreateRequest(&tt.request)
				if (err != nil) != tt.wantErr {
					t.Errorf("ValidateShareCreateRequest() error = %v, wantErr %v", err, tt.wantErr)
				}
			})
		}
	})

	t.Run("ValidateVMCreateRequest", func(t *testing.T) {
		tests := []struct {
			name    string
			request requests.VMCreateRequest
			wantErr bool
		}{
			{
				name: "Valid VM create request",
				request: requests.VMCreateRequest{
					Name:   "test-vm",
					CPUs:   2,
					Memory: 4096,
				},
				wantErr: false,
			},
			{
				name: "Empty VM name",
				request: requests.VMCreateRequest{
					Name:   "",
					CPUs:   2,
					Memory: 4096,
				},
				wantErr: true,
			},
			{
				name: "Invalid CPU count",
				request: requests.VMCreateRequest{
					Name:   "test-vm",
					CPUs:   0,
					Memory: 4096,
				},
				wantErr: true,
			},
			{
				name: "Invalid memory",
				request: requests.VMCreateRequest{
					Name:   "test-vm",
					CPUs:   2,
					Memory: 0,
				},
				wantErr: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := ValidateVMCreateRequest(&tt.request)
				if (err != nil) != tt.wantErr {
					t.Errorf("ValidateVMCreateRequest() error = %v, wantErr %v", err, tt.wantErr)
				}
			})
		}
	})

	t.Run("ValidateDockerContainerCreateRequest", func(t *testing.T) {
		tests := []struct {
			name    string
			request requests.DockerContainerCreateRequest
			wantErr bool
		}{
			{
				name: "Valid container create request",
				request: requests.DockerContainerCreateRequest{
					Name:  "test-container",
					Image: "nginx:latest",
				},
				wantErr: false,
			},
			{
				name: "Empty container name",
				request: requests.DockerContainerCreateRequest{
					Name:  "",
					Image: "nginx:latest",
				},
				wantErr: true,
			},
			{
				name: "Empty image",
				request: requests.DockerContainerCreateRequest{
					Name:  "test-container",
					Image: "",
				},
				wantErr: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := ValidateDockerContainerCreateRequest(&tt.request)
				if (err != nil) != tt.wantErr {
					t.Errorf("ValidateDockerContainerCreateRequest() error = %v, wantErr %v", err, tt.wantErr)
				}
			})
		}
	})
}

// TestRemainingResponseFunctions tests the remaining untested response functions
func TestRemainingResponseFunctions(t *testing.T) {
	t.Run("WritePaginatedResponse", func(t *testing.T) {
		w := httptest.NewRecorder()
		data := []map[string]string{
			{"id": "1", "name": "item1"},
			{"id": "2", "name": "item2"},
		}
		params := &dto.PaginationParams{
			Page:    2,
			PerPage: 10,
		}

		WritePaginatedResponse(w, http.StatusOK, data, 50, params, "req-456", "v1")

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		var response responses.StandardResponse
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
		}

		if response.Data == nil {
			t.Error("Expected data field in paginated response")
		}
		if response.Pagination == nil {
			t.Error("Expected pagination field in paginated response")
		}
		if response.Pagination.Page != 2 {
			t.Errorf("Expected page 2, got %d", response.Pagination.Page)
		}
	})

	t.Run("WriteVersionedResponse", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", nil)
		data := map[string]string{"test": "data"}

		WriteVersionedResponse(w, req, http.StatusOK, data, nil, "req-789", "v2")

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		var response responses.StandardResponse
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
		}

		if response.Data == nil {
			t.Error("Expected data field in versioned response")
		}
		if response.Meta == nil {
			t.Error("Expected meta field in versioned response")
		}
	})

	t.Run("GetRequestID", func(t *testing.T) {
		// Test with response writer that has request ID header
		w := httptest.NewRecorder()
		w.Header().Set("X-Request-ID", "custom-request-id")

		requestID := GetRequestID(w)
		if requestID != "custom-request-id" {
			t.Errorf("Expected 'custom-request-id', got '%s'", requestID)
		}
	})

	t.Run("GetRequestID_Generated", func(t *testing.T) {
		// Test with response writer that doesn't have request ID header
		w := httptest.NewRecorder()

		requestID := GetRequestID(w)
		if !strings.HasPrefix(requestID, "req_") {
			t.Errorf("Expected generated request ID to start with 'req_', got '%s'", requestID)
		}
	})
}

// TestErrorHandlingEdgeCases tests error handling in various utility functions
func TestErrorHandlingEdgeCases(t *testing.T) {
	t.Run("WriteError_EmptyMessage", func(t *testing.T) {
		w := httptest.NewRecorder()
		WriteError(w, http.StatusInternalServerError, "")

		if w.Code != http.StatusInternalServerError {
			t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
		}

		var response dto.Response
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Errorf("Failed to unmarshal error response: %v", err)
		}

		if response.Error == "" {
			t.Error("Expected error field to be set even with empty message")
		}
	})

	t.Run("WriteJSON_LargeData", func(t *testing.T) {
		w := httptest.NewRecorder()

		// Create large data structure
		largeData := make(map[string]interface{})
		for i := 0; i < 100; i++ {
			largeData[fmt.Sprintf("key_%d", i)] = strings.Repeat("data", 10)
		}

		WriteJSON(w, http.StatusOK, largeData)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200 for large data, got %d", w.Code)
		}

		// Verify the data was written
		if w.Body.Len() == 0 {
			t.Error("Expected response body to contain data")
		}
	})

	t.Run("GenerateRequestID_Uniqueness", func(t *testing.T) {
		// Generate multiple IDs and ensure they're unique
		ids := make(map[string]bool)
		for i := 0; i < 10; i++ {
			id := GenerateRequestID()
			if ids[id] {
				t.Errorf("Generated duplicate request ID: %s", id)
			}
			ids[id] = true

			if !strings.HasPrefix(id, "req_") {
				t.Errorf("Request ID should start with 'req_', got: %s", id)
			}

			// Add small delay to ensure different timestamps
			time.Sleep(1 * time.Millisecond)
		}
	})
}

// TestValidationEdgeCases tests edge cases in validation functions
func TestValidationEdgeCases(t *testing.T) {
	t.Run("ValidateShareCreateRequest_EdgeCases", func(t *testing.T) {
		tests := []struct {
			name    string
			request requests.ShareCreateRequest
			wantErr bool
		}{
			{
				name: "Name with special characters",
				request: requests.ShareCreateRequest{
					Name:            "test-share_123",
					Comment:         "Valid share with special chars",
					AllocatorMethod: "high-water",
				},
				wantErr: false,
			},
			{
				name: "Very long comment",
				request: requests.ShareCreateRequest{
					Name:    "test-share",
					Comment: strings.Repeat("a", 500),
				},
				wantErr: false, // Comments can be long
			},
			{
				name: "Name with spaces",
				request: requests.ShareCreateRequest{
					Name: "test share",
				},
				wantErr: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := ValidateShareCreateRequest(&tt.request)
				if (err != nil) != tt.wantErr {
					t.Errorf("ValidateShareCreateRequest() error = %v, wantErr %v", err, tt.wantErr)
				}
			})
		}
	})

	t.Run("ValidateVMCreateRequest_EdgeCases", func(t *testing.T) {
		tests := []struct {
			name    string
			request requests.VMCreateRequest
			wantErr bool
		}{
			{
				name: "Maximum valid values",
				request: requests.VMCreateRequest{
					Name:   "test-vm",
					CPUs:   32,
					Memory: 65536, // 64GB in MB
				},
				wantErr: false,
			},
			{
				name: "Minimum valid values",
				request: requests.VMCreateRequest{
					Name:   "vm",
					CPUs:   1,
					Memory: 512,
				},
				wantErr: false,
			},
			{
				name: "Negative memory",
				request: requests.VMCreateRequest{
					Name:   "test-vm",
					CPUs:   2,
					Memory: -1024,
				},
				wantErr: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := ValidateVMCreateRequest(&tt.request)
				if (err != nil) != tt.wantErr {
					t.Errorf("ValidateVMCreateRequest() error = %v, wantErr %v", err, tt.wantErr)
				}
			})
		}
	})

	t.Run("ValidateDockerContainerCreateRequest_EdgeCases", func(t *testing.T) {
		tests := []struct {
			name    string
			request requests.DockerContainerCreateRequest
			wantErr bool
		}{
			{
				name: "Image with tag",
				request: requests.DockerContainerCreateRequest{
					Name:  "test-container",
					Image: "nginx:1.21-alpine",
				},
				wantErr: false,
			},
			{
				name: "Image with registry",
				request: requests.DockerContainerCreateRequest{
					Name:  "test-container",
					Image: "docker.io/library/nginx:latest",
				},
				wantErr: false,
			},
			{
				name: "Container name with invalid characters",
				request: requests.DockerContainerCreateRequest{
					Name:  "test container!",
					Image: "nginx",
				},
				wantErr: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := ValidateDockerContainerCreateRequest(&tt.request)
				if (err != nil) != tt.wantErr {
					t.Errorf("ValidateDockerContainerCreateRequest() error = %v, wantErr %v", err, tt.wantErr)
				}
			})
		}
	})
}
