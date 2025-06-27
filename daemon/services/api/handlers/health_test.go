package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/domalab/uma/daemon/services/api/utils"
)

// TestHealthHandler_HandleHealth tests the health check endpoint
func TestHealthHandler_HandleHealth(t *testing.T) {
	handler := NewHealthHandler(&MockAPIInterface{}, "test-version")

	tests := []struct {
		name           string
		method         string
		expectedStatus int
	}{
		{
			name:           "GET health check",
			method:         "GET",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "POST method not allowed",
			method:         "POST",
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "PUT method not allowed",
			method:         "PUT",
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "DELETE method not allowed",
			method:         "DELETE",
			expectedStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/api/v1/health", nil)
			w := httptest.NewRecorder()

			handler.HandleHealth(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.method == "GET" {
				// Verify response structure for successful health check
				var response map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Fatalf("Failed to unmarshal health response: %v", err)
				}

				// Check for required fields in health response
				if response["status"] == nil {
					t.Error("Expected 'status' field in health response")
				}

				if response["timestamp"] == nil {
					t.Error("Expected 'timestamp' field in health response")
				}

				if response["version"] == nil {
					t.Error("Expected 'version' field in health response")
				}

				if response["checks"] == nil {
					t.Error("Expected 'checks' field in health response")
				}

				// Verify checks structure
				checks, ok := response["checks"].(map[string]interface{})
				if !ok {
					t.Error("Expected 'checks' to be an object")
				} else {
					// Check for common health checks
					expectedChecks := []string{"system", "storage", "docker", "auth"}
					for _, check := range expectedChecks {
						if checks[check] == nil {
							t.Errorf("Expected check '%s' in health response", check)
						}
					}
				}
			}
		})
	}
}

// TestHealthHandler_HandleHealthLive tests the liveness check endpoint
func TestHealthHandler_HandleHealthLive(t *testing.T) {
	handler := NewHealthHandler(&MockAPIInterface{}, "test-version")

	tests := []struct {
		name           string
		method         string
		expectedStatus int
	}{
		{
			name:           "GET liveness check",
			method:         "GET",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "POST method not allowed",
			method:         "POST",
			expectedStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/api/v1/health/live", nil)
			w := httptest.NewRecorder()

			handler.HandleHealthLive(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.method == "GET" {
				// Verify response structure for liveness check
				var response map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Fatalf("Failed to unmarshal liveness response: %v", err)
				}

				// Check for required fields in liveness response
				if response["status"] == nil {
					t.Error("Expected 'status' field in liveness response")
				}

				if response["timestamp"] == nil {
					t.Error("Expected 'timestamp' field in liveness response")
				}

				// Verify status field is string
				if status, ok := response["status"].(string); !ok {
					t.Error("Expected 'status' field to be string")
				} else if status != "alive" {
					t.Errorf("Expected status 'alive', got '%s'", status)
				}
			}
		})
	}
}

// TestHealthHandler_HandleHealthReady tests the readiness check endpoint
func TestHealthHandler_HandleHealthReady(t *testing.T) {
	handler := NewHealthHandler(&MockAPIInterface{}, "test-version")

	tests := []struct {
		name           string
		method         string
		expectedStatus int
	}{
		{
			name:           "GET readiness check",
			method:         "GET",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "POST method not allowed",
			method:         "POST",
			expectedStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/api/v1/health/ready", nil)
			w := httptest.NewRecorder()

			handler.HandleHealthReady(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.method == "GET" {
				// Verify response structure for readiness check
				var response map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Fatalf("Failed to unmarshal readiness response: %v", err)
				}

				// Check for required fields in readiness response
				if response["status"] == nil {
					t.Error("Expected 'status' field in readiness response")
				}

				if response["timestamp"] == nil {
					t.Error("Expected 'timestamp' field in readiness response")
				}

				if response["checks"] == nil {
					t.Error("Expected 'checks' field in readiness response")
				}

				// Verify status field contains ready boolean
				if statusObj, ok := response["status"].(map[string]interface{}); !ok {
					t.Error("Expected 'status' field to be an object")
				} else if ready, exists := statusObj["ready"]; !exists {
					t.Error("Expected 'ready' field in status object")
				} else if _, ok := ready.(bool); !ok {
					t.Error("Expected 'ready' field to be boolean")
				}
			}
		})
	}
}

// TestHealthHandler_HandleHealthLive_Duplicate tests the liveness check endpoint (duplicate removed)
// This test is now covered by the earlier TestHealthHandler_HandleHealthLive function
/*
func TestHealthHandler_HandleHealthLive_Duplicate(t *testing.T) {
	handler := NewHealthHandler(&MockAPIInterface{}, "test-version")

	tests := []struct {
		name           string
		method         string
		expectedStatus int
	}{
		{
			name:           "GET liveness check",
			method:         "GET",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "POST method not allowed",
			method:         "POST",
			expectedStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/api/v1/health/live", nil)
			w := httptest.NewRecorder()

			handler.HandleHealthLive(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.method == "GET" {
				// Verify response structure for liveness check
				var response map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Fatalf("Failed to unmarshal liveness response: %v", err)
				}

				// Check for required fields in liveness response
				if response["alive"] == nil {
					t.Error("Expected 'alive' field in liveness response")
				}

				if response["timestamp"] == nil {
					t.Error("Expected 'timestamp' field in liveness response")
				}

				// Verify alive field is boolean
				if _, ok := response["alive"].(bool); !ok {
					t.Error("Expected 'alive' field to be boolean")
				}
			}
		})
	}
}
*/

// TestHealthHandler_ContentType tests that health endpoints return proper content type
func TestHealthHandler_ContentType(t *testing.T) {
	handler := NewHealthHandler(&MockAPIInterface{}, "test-version")

	endpoints := []string{
		"/api/v1/health",
		"/api/v1/health/ready",
		"/api/v1/health/live",
	}

	for _, endpoint := range endpoints {
		t.Run("Content-Type for "+endpoint, func(t *testing.T) {
			req := httptest.NewRequest("GET", endpoint, nil)
			w := httptest.NewRecorder()

			switch endpoint {
			case "/api/v1/health":
				handler.HandleHealth(w, req)
			case "/api/v1/health/ready":
				handler.HandleHealthReady(w, req)
			case "/api/v1/health/live":
				handler.HandleHealthLive(w, req)
			}

			contentType := w.Header().Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
			}
		})
	}
}

// TestHealthHandler_ErrorConditions tests error conditions and edge cases
func TestHealthHandler_ErrorConditions(t *testing.T) {
	t.Run("HealthCheck_WithFailedDependencies", func(t *testing.T) {
		// Create a mock API that simulates failed dependencies
		mockAPI := &MockAPIWithFailures{}
		handler := NewHealthHandler(mockAPI, "test-version")

		req := httptest.NewRequest("GET", "/api/v1/health", nil)
		w := httptest.NewRecorder()

		handler.HandleHealth(w, req)

		// Should return 503 when dependencies fail, or 200 with degraded status
		if w.Code != http.StatusOK && w.Code != http.StatusServiceUnavailable {
			t.Errorf("Expected status %d or %d, got %d", http.StatusOK, http.StatusServiceUnavailable, w.Code)
		}

		var response map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		// Check that status indicates issues
		if status, ok := response["status"].(string); ok {
			if status == "healthy" {
				t.Error("Expected degraded status when dependencies fail")
			}
		}
	})

	t.Run("HealthCheck_ConcurrentRequests", func(t *testing.T) {
		mockAPI := &MockAPIInterface{}
		handler := NewHealthHandler(mockAPI, "test-version")

		const numRequests = 10
		results := make(chan int, numRequests)

		// Make concurrent health check requests
		for i := 0; i < numRequests; i++ {
			go func() {
				req := httptest.NewRequest("GET", "/api/v1/health", nil)
				w := httptest.NewRecorder()
				handler.HandleHealth(w, req)
				results <- w.Code
			}()
		}

		// Collect results
		for i := 0; i < numRequests; i++ {
			statusCode := <-results
			if statusCode != http.StatusOK {
				t.Errorf("Concurrent request %d: expected status %d, got %d", i, http.StatusOK, statusCode)
			}
		}
	})
}

// MockAPIWithFailures simulates an API with failing dependencies
type MockAPIWithFailures struct {
	MockAPIInterface
}

func (m *MockAPIWithFailures) GetSystem() utils.SystemInterface {
	return &MockSystemWithFailures{}
}

func (m *MockAPIWithFailures) GetDocker() utils.DockerInterface {
	return &MockDockerWithFailures{}
}

// MockSystemWithFailures simulates system interface failures
type MockSystemWithFailures struct{}

func (m *MockSystemWithFailures) GetCPUInfo() (interface{}, error) {
	return nil, fmt.Errorf("failed to read CPU info")
}

func (m *MockSystemWithFailures) GetMemoryInfo() (interface{}, error) {
	return nil, fmt.Errorf("failed to read memory info")
}

func (m *MockSystemWithFailures) GetLoadInfo() (interface{}, error) {
	return nil, fmt.Errorf("failed to read load info")
}

func (m *MockSystemWithFailures) GetUptimeInfo() (interface{}, error) {
	return nil, fmt.Errorf("failed to read uptime info")
}

func (m *MockSystemWithFailures) GetNetworkInfo() (interface{}, error) {
	return nil, fmt.Errorf("failed to read network info")
}

func (m *MockSystemWithFailures) GetEnhancedTemperatureData() (interface{}, error) {
	return nil, fmt.Errorf("failed to read temperature data")
}

func (m *MockSystemWithFailures) GetGPUInfo() (interface{}, error) {
	return nil, fmt.Errorf("failed to read GPU data")
}

func (m *MockSystemWithFailures) GetSystemLogs() (interface{}, error) {
	return nil, fmt.Errorf("failed to read system logs")
}

func (m *MockSystemWithFailures) GetRealArrayInfo() (interface{}, error) {
	return nil, fmt.Errorf("failed to read array info")
}

func (m *MockSystemWithFailures) GetRealDisks() (interface{}, error) {
	return nil, fmt.Errorf("failed to read disk info")
}

// MockDockerWithFailures simulates Docker interface failures
type MockDockerWithFailures struct{}

func (m *MockDockerWithFailures) GetContainers() (interface{}, error) {
	return nil, fmt.Errorf("docker daemon not available")
}

func (m *MockDockerWithFailures) GetContainersWithStats() (interface{}, error) {
	return nil, fmt.Errorf("docker daemon not available")
}

func (m *MockDockerWithFailures) GetContainer(id string) (interface{}, error) {
	return nil, fmt.Errorf("container not found")
}

func (m *MockDockerWithFailures) StartContainer(id string) error {
	return fmt.Errorf("failed to start container")
}

func (m *MockDockerWithFailures) StopContainer(id string, timeout int) error {
	return fmt.Errorf("failed to stop container")
}

func (m *MockDockerWithFailures) RestartContainer(id string, timeout int) error {
	return fmt.Errorf("failed to restart container")
}

func (m *MockDockerWithFailures) GetImages() (interface{}, error) {
	return nil, fmt.Errorf("failed to get images")
}

func (m *MockDockerWithFailures) GetNetworks() (interface{}, error) {
	return nil, fmt.Errorf("failed to get networks")
}

func (m *MockDockerWithFailures) GetContainerStats(id string) (interface{}, error) {
	return nil, fmt.Errorf("failed to get container stats")
}

func (m *MockDockerWithFailures) GetSystemInfo() (interface{}, error) {
	return nil, fmt.Errorf("docker system info unavailable")
}
