package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestSystemHandler_HandleSystemInfo tests the system info endpoint
func TestSystemHandler_HandleSystemInfo(t *testing.T) {
	handler := NewSystemHandler(&MockAPIInterface{})

	tests := []struct {
		name           string
		method         string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "GET request should return system info",
			method:         "GET",
			expectedStatus: http.StatusOK,
			expectedBody:   `{"status":"healthy","version":"test-version"}`,
		},
		{
			name:           "POST request should return method not allowed",
			method:         "POST",
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   `{"error":"Method not allowed"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/api/v1/system/info", nil)
			w := httptest.NewRecorder()

			handler.HandleSystemInfo(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			var response map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}

			// For successful requests, check specific fields
			if tt.expectedStatus == http.StatusOK {
				if response["version"] != "test-version" {
					t.Errorf("Expected version 'test-version', got %v", response["version"])
				}
				if response["status"] != "healthy" {
					t.Errorf("Expected status 'healthy', got %v", response["status"])
				}
			}

			// For error requests, check error message
			if tt.expectedStatus != http.StatusOK {
				if response["error"] != "Method not allowed" {
					t.Errorf("Expected error 'Method not allowed', got %v", response["error"])
				}
			}
		})
	}
}

// TestSystemHandler_HandleSystemCPU tests the CPU endpoint
func TestSystemHandler_HandleSystemCPU(t *testing.T) {
	handler := NewSystemHandler(&MockAPIInterface{})

	req := httptest.NewRequest("GET", "/api/v1/system/cpu", nil)
	w := httptest.NewRecorder()

	handler.HandleSystemCPU(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Check that required fields are present
	requiredFields := []string{"usage", "temperature", "cores", "last_updated"}
	for _, field := range requiredFields {
		if _, exists := response[field]; !exists {
			t.Errorf("Expected field '%s' to be present in response", field)
		}
	}
}

// TestSystemHandler_HandleSystemMemory tests the memory endpoint
func TestSystemHandler_HandleSystemMemory(t *testing.T) {
	handler := NewSystemHandler(&MockAPIInterface{})

	req := httptest.NewRequest("GET", "/api/v1/system/memory", nil)
	w := httptest.NewRecorder()

	handler.HandleSystemMemory(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Check that required fields are present
	requiredFields := []string{"total", "used", "available", "last_updated"}
	for _, field := range requiredFields {
		if _, exists := response[field]; !exists {
			t.Errorf("Expected field '%s' to be present in response", field)
		}
	}
}

// TestSystemHandler_HandleSystemExecute tests the command execution endpoint
func TestSystemHandler_HandleSystemExecute(t *testing.T) {
	handler := NewSystemHandler(&MockAPIInterface{})

	tests := []struct {
		name           string
		method         string
		body           string
		expectedStatus int
	}{
		{
			name:           "Valid command execution",
			method:         "POST",
			body:           `{"command":"echo test","timeout":30}`,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid JSON body",
			method:         "POST",
			body:           `invalid json`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Empty command",
			method:         "POST",
			body:           `{"command":"","timeout":30}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "GET method not allowed",
			method:         "GET",
			body:           "",
			expectedStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/api/v1/system/execute", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.HandleSystemExecute(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

// TestSystemHandler_HandleSystemScripts tests the scripts endpoint
func TestSystemHandler_HandleSystemScripts(t *testing.T) {
	handler := NewSystemHandler(&MockAPIInterface{})

	// Test GET request (list scripts)
	req := httptest.NewRequest("GET", "/api/v1/system/scripts", nil)
	w := httptest.NewRecorder()

	handler.HandleSystemScripts(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Check that scripts field is present
	if _, exists := response["scripts"]; !exists {
		t.Error("Expected 'scripts' field to be present in response")
	}
}

// TestSystemHandler_HandleSystemReboot tests the reboot endpoint
func TestSystemHandler_HandleSystemReboot(t *testing.T) {
	handler := NewSystemHandler(&MockAPIInterface{})

	tests := []struct {
		name           string
		method         string
		body           string
		expectedStatus int
	}{
		{
			name:           "Valid reboot request",
			method:         "POST",
			body:           `{"delay_seconds":0,"message":"Test reboot","force":false}`,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid delay (too high)",
			method:         "POST",
			body:           `{"delay_seconds":500}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "GET method not allowed",
			method:         "GET",
			body:           "",
			expectedStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/api/v1/system/reboot", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.HandleSystemReboot(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

// TestSystemHandler_HandleSystemLogs tests the logs endpoint
func TestSystemHandler_HandleSystemLogs(t *testing.T) {
	handler := NewSystemHandler(&MockAPIInterface{})

	req := httptest.NewRequest("GET", "/api/v1/system/logs?type=system&lines=100", nil)
	w := httptest.NewRecorder()

	handler.HandleSystemLogs(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Check that required fields are present
	requiredFields := []string{"type", "lines", "logs"}
	for _, field := range requiredFields {
		if _, exists := response[field]; !exists {
			t.Errorf("Expected field '%s' to be present in response", field)
		}
	}
}
