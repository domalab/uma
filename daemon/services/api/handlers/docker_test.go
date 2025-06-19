package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestDockerHandler_HandleDockerContainers tests the Docker containers endpoint
func TestDockerHandler_HandleDockerContainers(t *testing.T) {
	handler := NewDockerHandler(&MockAPIInterface{})

	req := httptest.NewRequest("GET", "/api/v1/docker/containers", nil)
	w := httptest.NewRecorder()

	handler.HandleDockerContainers(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response []interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Should return an array (even if empty)
	if response == nil {
		t.Error("Expected containers array in response")
	}
}

// TestDockerHandler_HandleDockerContainer tests individual container operations
func TestDockerHandler_HandleDockerContainer(t *testing.T) {
	handler := NewDockerHandler(&MockAPIInterface{})

	tests := []struct {
		name           string
		method         string
		containerID    string
		action         string
		expectedStatus int
	}{
		{
			name:           "Get container info",
			method:         "GET",
			containerID:    "test-container",
			action:         "",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Start container",
			method:         "POST",
			containerID:    "test-container",
			action:         "start",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Stop container",
			method:         "POST",
			containerID:    "test-container",
			action:         "stop",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Restart container",
			method:         "POST",
			containerID:    "test-container",
			action:         "restart",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid action",
			method:         "POST",
			containerID:    "test-container",
			action:         "invalid",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request
			if tt.action != "" {
				req = httptest.NewRequest(tt.method, "/api/v1/docker/containers/"+tt.containerID+"/"+tt.action, nil)
			} else {
				req = httptest.NewRequest(tt.method, "/api/v1/docker/containers/"+tt.containerID+"/info", nil)
			}
			w := httptest.NewRecorder()

			handler.HandleDockerContainer(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

// TestDockerHandler_HandleDockerBulkStart tests bulk container start operations
func TestDockerHandler_HandleDockerBulkStart(t *testing.T) {
	handler := NewDockerHandler(&MockAPIInterface{})

	tests := []struct {
		name           string
		method         string
		body           string
		expectedStatus int
	}{
		{
			name:           "Bulk start containers",
			method:         "POST",
			body:           `{"container_ids":["container1","container2"]}`,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid JSON",
			method:         "POST",
			body:           `invalid json`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Empty container list",
			method:         "POST",
			body:           `{"container_ids":[]}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Method not allowed",
			method:         "GET",
			body:           "",
			expectedStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/api/v1/docker/containers/bulk/start", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.HandleDockerBulkStart(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

// TestDockerHandler_HandleDockerImages tests the Docker images endpoint
func TestDockerHandler_HandleDockerImages(t *testing.T) {
	handler := NewDockerHandler(&MockAPIInterface{})

	req := httptest.NewRequest("GET", "/api/v1/docker/images", nil)
	w := httptest.NewRecorder()

	handler.HandleDockerImages(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response []interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Should return an array (even if empty)
	if response == nil {
		t.Error("Expected images array in response")
	}
}

// TestDockerHandler_HandleDockerNetworks tests the Docker networks endpoint
func TestDockerHandler_HandleDockerNetworks(t *testing.T) {
	handler := NewDockerHandler(&MockAPIInterface{})

	req := httptest.NewRequest("GET", "/api/v1/docker/networks", nil)
	w := httptest.NewRecorder()

	handler.HandleDockerNetworks(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response []interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Should return an array (even if empty)
	if response == nil {
		t.Error("Expected networks array in response")
	}
}

// TestDockerHandler_HandleDockerInfo tests the Docker system info endpoint
func TestDockerHandler_HandleDockerInfo(t *testing.T) {
	handler := NewDockerHandler(&MockAPIInterface{})

	req := httptest.NewRequest("GET", "/api/v1/docker/info", nil)
	w := httptest.NewRecorder()

	handler.HandleDockerInfo(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Should return an object with Docker info
	if response == nil {
		t.Error("Expected Docker info object in response")
	}
}

// TestDockerHandler_MethodNotAllowed tests unsupported HTTP methods
func TestDockerHandler_MethodNotAllowed(t *testing.T) {
	handler := NewDockerHandler(&MockAPIInterface{})

	tests := []struct {
		name   string
		method string
		path   string
	}{
		{
			name:   "POST on containers list",
			method: "POST",
			path:   "/api/v1/docker/containers",
		},
		{
			name:   "PUT on container",
			method: "PUT",
			path:   "/api/v1/docker/containers/test/start",
		},
		{
			name:   "DELETE on images",
			method: "DELETE",
			path:   "/api/v1/docker/images",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			// Route to appropriate handler based on path
			if tt.path == "/api/v1/docker/containers" {
				handler.HandleDockerContainers(w, req)
			} else if tt.path == "/api/v1/docker/images" {
				handler.HandleDockerImages(w, req)
			} else {
				handler.HandleDockerContainer(w, req)
			}

			if w.Code != http.StatusMethodNotAllowed {
				t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
			}
		})
	}
}
