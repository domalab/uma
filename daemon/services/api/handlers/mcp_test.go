package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/domalab/uma/daemon/domain"
	"github.com/domalab/uma/daemon/services/api/utils"
	"github.com/domalab/uma/daemon/services/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockMCPServer provides a mock MCP server for testing
type MockMCPServer struct {
	enabled bool
	stats   map[string]interface{}
	tools   []mcp.Tool
}

func (m *MockMCPServer) GetServerStats() map[string]interface{} {
	if m.stats != nil {
		return m.stats
	}
	return map[string]interface{}{
		"enabled":            m.enabled,
		"port":               34800,
		"max_connections":    100,
		"active_connections": 2,
		"total_tools":        len(m.tools),
	}
}

func (m *MockMCPServer) GetRegistry() *mcp.SimpleToolRegistry {
	// Return a mock registry for testing
	return &mcp.SimpleToolRegistry{}
}

// MockAPIWithMCP extends MockAPIInterface to include MCP server
type MockAPIWithMCP struct {
	config *domain.Config
}

func (m *MockAPIWithMCP) GetInfo() interface{} {
	return map[string]interface{}{
		"version": "test-version",
		"status":  "healthy",
	}
}

func (m *MockAPIWithMCP) GetSystem() utils.SystemInterface              { return nil }
func (m *MockAPIWithMCP) GetStorage() utils.StorageInterface            { return nil }
func (m *MockAPIWithMCP) GetDocker() utils.DockerInterface              { return nil }
func (m *MockAPIWithMCP) GetVM() utils.VMInterface                      { return nil }
func (m *MockAPIWithMCP) GetAuth() utils.AuthInterface                  { return nil }
func (m *MockAPIWithMCP) GetNotifications() utils.NotificationInterface { return nil }
func (m *MockAPIWithMCP) GetUPSDetector() utils.UPSDetectorInterface    { return nil }

func (m *MockAPIWithMCP) GetMCPServer() interface{} {
	// Return nil to simulate disabled MCP server, or create a mock
	return nil
}

func (m *MockAPIWithMCP) GetConfig() *domain.Config {
	if m.config != nil {
		return m.config
	}
	return &domain.Config{
		MCP: domain.MCPConfig{
			Enabled:        false,
			MaxConnections: 100,
		},
	}
}

func (m *MockAPIWithMCP) GetConfigManager() interface{} {
	return &MockConfigManager{}
}

// TestGetMCPStatus tests MCP status endpoint
func TestGetMCPStatus(t *testing.T) {
	tests := []struct {
		name           string
		mcpServerNil   bool
		expectedStatus int
		checkResponse  func(t *testing.T, body map[string]interface{})
	}{
		{
			name:           "MCP server disabled",
			mcpServerNil:   true,
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				assert.True(t, body["success"].(bool))
				data := body["data"].(map[string]interface{})
				assert.False(t, data["enabled"].(bool))
				assert.Equal(t, "disabled", data["status"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAPI := &MockAPIWithMCP{}
			handler := NewMCPHandler(mockAPI)

			req := httptest.NewRequest("GET", "/api/v1/mcp/status", nil)
			req.Header.Set("X-Request-ID", "test-123")
			w := httptest.NewRecorder()

			handler.GetMCPStatus(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			tt.checkResponse(t, response)
		})
	}
}

// TestGetMCPTools tests MCP tools endpoint
func TestGetMCPTools(t *testing.T) {
	tests := []struct {
		name           string
		mcpServerNil   bool
		expectedStatus int
		checkResponse  func(t *testing.T, body map[string]interface{})
	}{
		{
			name:           "MCP server disabled",
			mcpServerNil:   true,
			expectedStatus: http.StatusServiceUnavailable,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				assert.False(t, body["success"].(bool))
				assert.Contains(t, body["error"], "not enabled")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAPI := &MockAPIWithMCP{}
			handler := NewMCPHandler(mockAPI)

			req := httptest.NewRequest("GET", "/api/v1/mcp/tools", nil)
			req.Header.Set("X-Request-ID", "test-123")
			w := httptest.NewRecorder()

			handler.GetMCPTools(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			tt.checkResponse(t, response)
		})
	}
}

// TestGetMCPConfig tests MCP configuration endpoint
func TestGetMCPConfig(t *testing.T) {
	mockAPI := &MockAPIWithMCP{}
	handler := NewMCPHandler(mockAPI)

	req := httptest.NewRequest("GET", "/api/v1/mcp/config", nil)
	req.Header.Set("X-Request-ID", "test-123")
	w := httptest.NewRecorder()

	handler.GetMCPConfig(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response["success"].(bool))
	data := response["data"].(map[string]interface{})
	assert.False(t, data["enabled"].(bool))
	assert.Equal(t, float64(34800), data["port"].(float64))
	assert.Equal(t, float64(100), data["max_connections"].(float64))
}

// TestUpdateMCPConfig tests MCP configuration update endpoint
func TestUpdateMCPConfig(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		expectedStatus int
		checkResponse  func(t *testing.T, body map[string]interface{})
	}{
		{
			name: "valid update - enable MCP",
			requestBody: map[string]interface{}{
				"enabled": true,
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				assert.True(t, body["success"].(bool))
				data := body["data"].(map[string]interface{})
				assert.Contains(t, data, "message")
			},
		},
		{
			name: "invalid port",
			requestBody: map[string]interface{}{
				"port": 1023,
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				assert.False(t, body["success"].(bool))
				assert.Contains(t, body["error"], "Port must be between 1024 and 65535")
			},
		},
		{
			name: "invalid max connections",
			requestBody: map[string]interface{}{
				"max_connections": 0,
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				assert.False(t, body["success"].(bool))
				assert.Contains(t, body["error"], "connections")
			},
		},
		{
			name:           "invalid JSON",
			requestBody:    nil, // Will send invalid JSON
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				assert.False(t, body["success"].(bool))
				assert.Contains(t, body["error"], "Invalid request body")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAPI := &MockAPIWithMCP{
				config: &domain.Config{
					MCP: domain.MCPConfig{
						Enabled:        false,
						MaxConnections: 100,
					},
				},
			}
			handler := NewMCPHandler(mockAPI)

			var reqBody *bytes.Buffer
			if tt.requestBody != nil {
				jsonData, _ := json.Marshal(tt.requestBody)
				reqBody = bytes.NewBuffer(jsonData)
			} else {
				reqBody = bytes.NewBufferString("invalid json")
			}

			req := httptest.NewRequest("PUT", "/api/v1/mcp/config", reqBody)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Request-ID", "test-123")
			w := httptest.NewRecorder()

			handler.UpdateMCPConfig(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			tt.checkResponse(t, response)
		})
	}
}

// TestRefreshMCPTools tests MCP tools refresh endpoint
func TestRefreshMCPTools(t *testing.T) {
	mockAPI := &MockAPIWithMCP{}
	handler := NewMCPHandler(mockAPI)

	req := httptest.NewRequest("POST", "/api/v1/mcp/tools/refresh", nil)
	req.Header.Set("X-Request-ID", "test-123")
	w := httptest.NewRecorder()

	handler.RefreshMCPTools(w, req)

	// Should return 503 because MCP server is not enabled
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response["success"].(bool))
	assert.Contains(t, response["error"], "not enabled")
}

// TestGetMCPToolsByCategory tests MCP tools by category endpoint
func TestGetMCPToolsByCategory(t *testing.T) {
	mockAPI := &MockAPIWithMCP{}
	handler := NewMCPHandler(mockAPI)

	req := httptest.NewRequest("GET", "/api/v1/mcp/tools/categories", nil)
	req.Header.Set("X-Request-ID", "test-123")
	w := httptest.NewRecorder()

	handler.GetMCPToolsByCategory(w, req)

	// Should return 503 because MCP server is not enabled
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response["success"].(bool))
	assert.Contains(t, response["error"], "not enabled")
}

// TestMCPHandlerCreation tests MCP handler creation
func TestMCPHandlerCreation(t *testing.T) {
	mockAPI := &MockAPIWithMCP{}
	handler := NewMCPHandler(mockAPI)

	assert.NotNil(t, handler)
	assert.Equal(t, mockAPI, handler.api)
}

// TestHTTPMethodValidation tests HTTP method validation for all endpoints
func TestHTTPMethodValidation(t *testing.T) {
	mockAPI := &MockAPIWithMCP{}
	handler := NewMCPHandler(mockAPI)

	tests := []struct {
		name           string
		method         string
		path           string
		handlerFunc    func(w http.ResponseWriter, r *http.Request)
		expectedStatus int
	}{
		{
			name:           "GetMCPStatus with POST method",
			method:         "POST",
			path:           "/api/v1/mcp/status",
			handlerFunc:    handler.GetMCPStatus,
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "GetMCPTools with PUT method",
			method:         "PUT",
			path:           "/api/v1/mcp/tools",
			handlerFunc:    handler.GetMCPTools,
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "GetMCPConfig with DELETE method",
			method:         "DELETE",
			path:           "/api/v1/mcp/config",
			handlerFunc:    handler.GetMCPConfig,
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "UpdateMCPConfig with GET method",
			method:         "GET",
			path:           "/api/v1/mcp/config",
			handlerFunc:    handler.UpdateMCPConfig,
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "RefreshMCPTools with GET method",
			method:         "GET",
			path:           "/api/v1/mcp/tools/refresh",
			handlerFunc:    handler.RefreshMCPTools,
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "GetMCPToolsByCategory with POST method",
			method:         "POST",
			path:           "/api/v1/mcp/tools/categories",
			handlerFunc:    handler.GetMCPToolsByCategory,
			expectedStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			tt.handlerFunc(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestUpdateMCPConfigEdgeCases tests edge cases for configuration updates
func TestUpdateMCPConfigEdgeCases(t *testing.T) {
	mockAPI := &MockAPIWithMCP{}
	handler := NewMCPHandler(mockAPI)

	tests := []struct {
		name           string
		requestBody    string
		expectedStatus int
		checkResponse  func(t *testing.T, body map[string]interface{})
	}{
		{
			name:           "empty request body",
			requestBody:    "{}",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				assert.True(t, body["success"].(bool))
				data := body["data"].(map[string]interface{})
				assert.Contains(t, data["message"], "updated successfully")
			},
		},

		{
			name:           "max connections boundary - minimum valid",
			requestBody:    `{"max_connections": 1}`,
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				assert.True(t, body["success"].(bool))
			},
		},
		{
			name:           "max connections boundary - zero",
			requestBody:    `{"max_connections": 0}`,
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				assert.False(t, body["success"].(bool))
				assert.Contains(t, body["error"], "Max connections must be greater than 0")
			},
		},
		{
			name:           "max connections boundary - negative",
			requestBody:    `{"max_connections": -1}`,
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				assert.False(t, body["success"].(bool))
				assert.Contains(t, body["error"], "Max connections must be greater than 0")
			},
		},
		{
			name:           "multiple valid fields",
			requestBody:    `{"enabled": true, "port": 35000, "max_connections": 200}`,
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				assert.True(t, body["success"].(bool))
			},
		},
		{
			name:           "malformed JSON",
			requestBody:    `{"enabled": true, "port":}`,
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				assert.False(t, body["success"].(bool))
				assert.Contains(t, body["error"], "Invalid request body")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("PUT", "/api/v1/mcp/config", strings.NewReader(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.UpdateMCPConfig(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			tt.checkResponse(t, response)
		})
	}
}

// Benchmark tests for MCP handler performance
func BenchmarkGetMCPStatus(b *testing.B) {
	mockAPI := &MockAPIWithMCP{}
	handler := NewMCPHandler(mockAPI)

	req := httptest.NewRequest("GET", "/api/v1/mcp/status", nil)
	req.Header.Set("X-Request-ID", "bench-test")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		handler.GetMCPStatus(w, req)
	}
}

func BenchmarkGetMCPConfig(b *testing.B) {
	mockAPI := &MockAPIWithMCP{}
	handler := NewMCPHandler(mockAPI)

	req := httptest.NewRequest("GET", "/api/v1/mcp/config", nil)
	req.Header.Set("X-Request-ID", "bench-test")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		handler.GetMCPConfig(w, req)
	}
}
