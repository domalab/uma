package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/domalab/uma/daemon/services/api/types/requests"
)

// TestStorageHandler_HandleStorageArray tests the storage array endpoint
func TestStorageHandler_HandleStorageArray(t *testing.T) {
	mockAPI := &MockAPIInterface{}
	handler := NewStorageHandler(mockAPI)

	tests := []struct {
		name           string
		method         string
		expectedStatus int
	}{
		{
			name:           "GET request should return array info",
			method:         "GET",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "POST request should return method not allowed",
			method:         "POST",
			expectedStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/api/v1/storage/array", nil)
			w := httptest.NewRecorder()

			handler.HandleStorageArray(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.method == "GET" && w.Code == http.StatusOK {
				var response map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}
			}
		})
	}
}

// TestStorageHandler_HandleStorageDisks tests the storage disks endpoint
func TestStorageHandler_HandleStorageDisks(t *testing.T) {
	mockAPI := &MockAPIInterface{}
	handler := NewStorageHandler(mockAPI)

	tests := []struct {
		name           string
		method         string
		expectedStatus int
	}{
		{
			name:           "GET request should return disk info",
			method:         "GET",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "PUT request should return method not allowed",
			method:         "PUT",
			expectedStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/api/v1/storage/disks", nil)
			w := httptest.NewRecorder()

			handler.HandleStorageDisks(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

// TestStorageHandler_HandleStorageZFS tests the ZFS storage endpoint
func TestStorageHandler_HandleStorageZFS(t *testing.T) {
	mockAPI := &MockAPIInterface{}
	handler := NewStorageHandler(mockAPI)

	tests := []struct {
		name           string
		method         string
		expectedStatus int
	}{
		{
			name:           "GET request should return ZFS info",
			method:         "GET",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "DELETE request should return method not allowed",
			method:         "DELETE",
			expectedStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/api/v1/storage/zfs", nil)
			w := httptest.NewRecorder()

			handler.HandleStorageZFS(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

// TestStorageHandler_HandleStorageCache tests the cache storage endpoint
func TestStorageHandler_HandleStorageCache(t *testing.T) {
	mockAPI := &MockAPIInterface{}
	handler := NewStorageHandler(mockAPI)

	req := httptest.NewRequest("GET", "/api/v1/storage/cache", nil)
	w := httptest.NewRecorder()

	handler.HandleStorageCache(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}
}

// TestStorageHandler_HandleStorageBoot tests the boot storage endpoint
func TestStorageHandler_HandleStorageBoot(t *testing.T) {
	mockAPI := &MockAPIInterface{}
	handler := NewStorageHandler(mockAPI)

	req := httptest.NewRequest("GET", "/api/v1/storage/boot", nil)
	w := httptest.NewRecorder()

	handler.HandleStorageBoot(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	// Check for expected fields - boot storage should have basic usage info
	if response["device"] == nil {
		t.Error("Expected 'device' field in boot storage response")
	}
	if response["used"] == nil {
		t.Error("Expected 'used' field in boot storage response")
	}
	if response["available"] == nil {
		t.Error("Expected 'available' field in boot storage response")
	}
}

// TestStorageHandler_HandleStorageGeneral tests the general storage endpoint
func TestStorageHandler_HandleStorageGeneral(t *testing.T) {
	mockAPI := &MockAPIInterface{}
	handler := NewStorageHandler(mockAPI)

	req := httptest.NewRequest("GET", "/api/v1/storage/general", nil)
	w := httptest.NewRecorder()

	handler.HandleStorageGeneral(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	// Check for expected fields
	expectedFields := []string{"docker_vdisk", "log_usage", "boot_usage", "last_updated"}
	for _, field := range expectedFields {
		if response[field] == nil {
			t.Errorf("Expected '%s' field in general storage response", field)
		}
	}
}

// TestStorageHandler_HandleArrayStart tests the array start endpoint
func TestStorageHandler_HandleArrayStart(t *testing.T) {
	mockAPI := &MockAPIInterface{}
	handler := NewStorageHandler(mockAPI)

	tests := []struct {
		name           string
		method         string
		body           interface{}
		expectedStatus int
	}{
		{
			name:   "Valid array start request",
			method: "POST",
			body: requests.ArrayStartRequest{
				MaintenanceMode: false,
				CheckFilesystem: true,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid JSON body",
			method:         "POST",
			body:           "invalid json",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "GET method not allowed",
			method:         "GET",
			body:           nil,
			expectedStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body []byte
			if tt.body != nil {
				if str, ok := tt.body.(string); ok {
					body = []byte(str)
				} else {
					body, _ = json.Marshal(tt.body)
				}
			}

			req := httptest.NewRequest(tt.method, "/api/v1/storage/array/start", bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			handler.HandleArrayStart(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}

				if response["success"] == nil {
					t.Error("Expected 'success' field in array start response")
				}
			}
		})
	}
}

// TestStorageHandler_HandleArrayStop tests the array stop endpoint
func TestStorageHandler_HandleArrayStop(t *testing.T) {
	mockAPI := &MockAPIInterface{}
	handler := NewStorageHandler(mockAPI)

	tests := []struct {
		name           string
		method         string
		body           interface{}
		expectedStatus int
	}{
		{
			name:   "Valid array stop request",
			method: "POST",
			body: requests.ArrayStopRequest{
				Force:          true,
				UnmountShares:  true,
				StopContainers: false,
				StopVMs:        false,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid JSON body",
			method:         "POST",
			body:           "invalid json",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "PUT method not allowed",
			method:         "PUT",
			body:           nil,
			expectedStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body []byte
			if tt.body != nil {
				if str, ok := tt.body.(string); ok {
					body = []byte(str)
				} else {
					body, _ = json.Marshal(tt.body)
				}
			}

			req := httptest.NewRequest(tt.method, "/api/v1/storage/array/stop", bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			handler.HandleArrayStop(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

// TestStorageHandler_HandleParityCheck tests the parity check endpoint
func TestStorageHandler_HandleParityCheck(t *testing.T) {
	mockAPI := &MockAPIInterface{}
	handler := NewStorageHandler(mockAPI)

	tests := []struct {
		name           string
		method         string
		body           interface{}
		expectedStatus int
	}{
		{
			name:           "GET request should return parity status",
			method:         "GET",
			body:           nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:   "POST request should start parity check",
			method: "POST",
			body: requests.ParityCheckRequest{
				Type:     "read-check",
				Priority: "normal",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid JSON body",
			method:         "POST",
			body:           "invalid json",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "PUT method not allowed",
			method:         "PUT",
			body:           nil,
			expectedStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body []byte
			if tt.body != nil {
				if str, ok := tt.body.(string); ok {
					body = []byte(str)
				} else {
					body, _ = json.Marshal(tt.body)
				}
			}

			req := httptest.NewRequest(tt.method, "/api/v1/system/parity/check", bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			handler.HandleParityCheck(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}

				if tt.method == "GET" {
					// Check for parity status fields
					if response["active"] == nil {
						t.Error("Expected 'active' field in parity status response")
					}
				} else if tt.method == "POST" {
					// Check for operation response fields
					if response["success"] == nil {
						t.Error("Expected 'success' field in parity check start response")
					}
				}
			}
		})
	}
}
