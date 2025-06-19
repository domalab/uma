package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/domalab/uma/daemon/services/api/types/requests"
)

// TestVMHandler_HandleVMList tests the VM list endpoint
func TestVMHandler_HandleVMList(t *testing.T) {
	mockAPI := &MockAPIInterface{}
	handler := NewVMHandler(mockAPI)

	tests := []struct {
		name           string
		method         string
		expectedStatus int
	}{
		{
			name:           "GET request should return VM list",
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
			req := httptest.NewRequest(tt.method, "/api/v1/vms", nil)
			w := httptest.NewRecorder()

			handler.HandleVMList(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.method == "GET" && w.Code == http.StatusOK {
				var response []interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}
			}
		})
	}
}

// TestVMHandler_HandleVM tests individual VM operations
func TestVMHandler_HandleVM(t *testing.T) {
	mockAPI := &MockAPIInterface{}
	handler := NewVMHandler(mockAPI)

	tests := []struct {
		name           string
		method         string
		vmName         string
		action         string
		expectedStatus int
	}{
		{
			name:           "GET VM info",
			method:         "GET",
			vmName:         "test-vm",
			action:         "",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "GET VM stats",
			method:         "GET",
			vmName:         "test-vm",
			action:         "stats",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "GET VM console",
			method:         "GET",
			vmName:         "test-vm",
			action:         "console",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "POST start VM",
			method:         "POST",
			vmName:         "test-vm",
			action:         "start",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "POST stop VM",
			method:         "POST",
			vmName:         "test-vm",
			action:         "stop",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "POST restart VM",
			method:         "POST",
			vmName:         "test-vm",
			action:         "restart",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "POST pause VM",
			method:         "POST",
			vmName:         "test-vm",
			action:         "pause",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "POST resume VM",
			method:         "POST",
			vmName:         "test-vm",
			action:         "resume",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "POST reset VM",
			method:         "POST",
			vmName:         "test-vm",
			action:         "reset",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "PUT update VM",
			method:         "PUT",
			vmName:         "test-vm",
			action:         "",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "DELETE VM",
			method:         "DELETE",
			vmName:         "test-vm",
			action:         "",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid action",
			method:         "POST",
			vmName:         "test-vm",
			action:         "invalid",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid method",
			method:         "PATCH",
			vmName:         "test-vm",
			action:         "",
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "Missing VM name",
			method:         "GET",
			vmName:         "",
			action:         "",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var url string
			if tt.vmName == "" {
				url = "/api/v1/vms/"
			} else if tt.action == "" {
				url = "/api/v1/vms/" + tt.vmName
			} else {
				url = "/api/v1/vms/" + tt.vmName + "/" + tt.action
			}

			var body []byte
			if tt.method == "PUT" {
				updateRequest := requests.VMUpdateRequest{
					CPUs:   4,
					Memory: 8192,
				}
				body, _ = json.Marshal(updateRequest)
			}

			req := httptest.NewRequest(tt.method, url, bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			handler.HandleVM(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}

				// Check for expected fields based on operation
				if tt.method == "POST" || tt.method == "PUT" || tt.method == "DELETE" {
					if response["success"] == nil {
						t.Error("Expected 'success' field in operation response")
					}
					if response["message"] == nil {
						t.Error("Expected 'message' field in operation response")
					}
				}
			}
		})
	}
}

// TestVMHandler_HandleVMCreate tests VM creation
func TestVMHandler_HandleVMCreate(t *testing.T) {
	mockAPI := &MockAPIInterface{}
	handler := NewVMHandler(mockAPI)

	tests := []struct {
		name           string
		method         string
		body           interface{}
		expectedStatus int
	}{
		{
			name:   "Valid VM create request",
			method: "POST",
			body: requests.VMCreateRequest{
				Name:        "new-vm",
				Description: "Test VM",
				CPUs:        2,
				Memory:      4096,
				Autostart:   false,
			},
			expectedStatus: http.StatusCreated,
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

			req := httptest.NewRequest(tt.method, "/api/v1/vms", bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			handler.HandleVMCreate(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusCreated {
				var response map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}

				if response["success"] == nil {
					t.Error("Expected 'success' field in create response")
				}
				if response["vm_name"] == nil {
					t.Error("Expected 'vm_name' field in create response")
				}
			}
		})
	}
}

// TestVMHandler_HandleVMSnapshot tests VM snapshot operations
func TestVMHandler_HandleVMSnapshot(t *testing.T) {
	mockAPI := &MockAPIInterface{}
	handler := NewVMHandler(mockAPI)

	tests := []struct {
		name           string
		method         string
		body           interface{}
		expectedStatus int
	}{
		{
			name:           "GET snapshots list",
			method:         "GET",
			body:           nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:   "POST create snapshot",
			method: "POST",
			body: requests.VMSnapshotRequest{
				Name:        "test-snapshot",
				Description: "Test snapshot",
			},
			expectedStatus: http.StatusCreated,
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

			req := httptest.NewRequest(tt.method, "/api/v1/vms/test-vm/snapshots", bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			handler.HandleVMSnapshot(w, req, "test-vm")

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusOK || tt.expectedStatus == http.StatusCreated {
				var response interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}

				if tt.method == "GET" {
					// Should return array of snapshots
					if _, ok := response.([]interface{}); !ok {
						t.Error("Expected array response for GET snapshots")
					}
				} else if tt.method == "POST" {
					// Should return operation response
					if respMap, ok := response.(map[string]interface{}); ok {
						if respMap["success"] == nil {
							t.Error("Expected 'success' field in snapshot create response")
						}
					}
				}
			}
		})
	}
}
