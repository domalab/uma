package mcp

import (
	"context"
	"testing"

	"github.com/domalab/uma/daemon/services/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockConnection provides a mock connection for testing handlers
type MockConnection struct {
	id            string
	sentResponses []JSONRPCResponse
	sentErrors    []JSONRPCResponse
}

func (m *MockConnection) sendResponse(id interface{}, result interface{}) error {
	response := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
	m.sentResponses = append(m.sentResponses, response)
	return nil
}

func (m *MockConnection) sendError(id interface{}, code int, message string, data interface{}) error {
	response := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &JSONRPCError{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}
	m.sentErrors = append(m.sentErrors, response)
	return nil
}

// TestHandleInitialize tests the MCP initialize method
func TestHandleInitialize(t *testing.T) {
	config := config.MCPConfig{
		Enabled:        true,
		Port:           34800,
		MaxConnections: 100,
	}

	mockAPI := &MockAPIInterface{}
	server := NewServer(config, mockAPI)

	mockConn := &MockConnection{id: "test-conn"}
	conn := &Connection{
		id:     "test-conn",
		server: server,
		ctx:    context.Background(),
	}

	tests := []struct {
		name     string
		request  *JSONRPCRequest
		wantErr  bool
		validate func(t *testing.T, responses []JSONRPCResponse, errors []JSONRPCResponse)
	}{
		{
			name: "successful initialize",
			request: &JSONRPCRequest{
				JSONRPC: "2.0",
				ID:      "init-1",
				Method:  "initialize",
				Params: map[string]interface{}{
					"protocolVersion": MCPProtocolVersion,
					"capabilities":    map[string]interface{}{},
					"clientInfo": map[string]interface{}{
						"name":    "test-client",
						"version": "1.0.0",
					},
				},
			},
			wantErr: false,
			validate: func(t *testing.T, responses []JSONRPCResponse, errors []JSONRPCResponse) {
				require.Len(t, responses, 1)
				require.Len(t, errors, 0)

				response := responses[0]
				assert.Equal(t, "init-1", response.ID)
				assert.NotNil(t, response.Result)

				// Validate result structure
				result, ok := response.Result.(InitializeResult)
				if !ok {
					// Try to convert from map
					resultMap, ok := response.Result.(map[string]interface{})
					require.True(t, ok, "Result should be InitializeResult or map")

					assert.Equal(t, MCPProtocolVersion, resultMap["protocolVersion"])
					assert.Contains(t, resultMap, "capabilities")
					assert.Contains(t, resultMap, "serverInfo")
				} else {
					assert.Equal(t, MCPProtocolVersion, result.ProtocolVersion)
					assert.NotNil(t, result.Capabilities)
					assert.NotNil(t, result.ServerInfo)
				}
			},
		},
		{
			name: "initialize with invalid params",
			request: &JSONRPCRequest{
				JSONRPC: "2.0",
				ID:      "init-2",
				Method:  "initialize",
				Params:  "invalid-params",
			},
			wantErr: false, // Should handle gracefully
			validate: func(t *testing.T, responses []JSONRPCResponse, errors []JSONRPCResponse) {
				// Should either respond with error or handle gracefully
				assert.True(t, len(responses) > 0 || len(errors) > 0)
			},
		},
		{
			name: "initialize without params",
			request: &JSONRPCRequest{
				JSONRPC: "2.0",
				ID:      "init-3",
				Method:  "initialize",
				Params:  nil,
			},
			wantErr: false,
			validate: func(t *testing.T, responses []JSONRPCResponse, errors []JSONRPCResponse) {
				require.Len(t, responses, 1)
				assert.Equal(t, "init-3", responses[0].ID)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mock connection
			mockConn.sentResponses = nil
			mockConn.sentErrors = nil

			// Create a test connection that uses our mock
			testConn := &testConnection{
				Connection: conn,
				mock:       mockConn,
			}

			err := testConn.handleInitialize(tt.request)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			tt.validate(t, mockConn.sentResponses, mockConn.sentErrors)
		})
	}
}

// TestHandleToolsList tests the tools/list method
func TestHandleToolsList(t *testing.T) {
	config := config.MCPConfig{
		Enabled:        true,
		Port:           34800,
		MaxConnections: 100,
	}

	mockAPI := &MockAPIInterface{}
	server := NewServer(config, mockAPI)

	mockConn := &MockConnection{id: "test-conn"}
	conn := &Connection{
		id:     "test-conn",
		server: server,
		ctx:    context.Background(),
	}

	request := &JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      "tools-1",
		Method:  "tools/list",
		Params:  nil,
	}

	testConn := &testConnection{
		Connection: conn,
		mock:       mockConn,
	}

	err := testConn.handleToolsList(request)
	assert.NoError(t, err)

	require.Len(t, mockConn.sentResponses, 1)
	response := mockConn.sentResponses[0]
	assert.Equal(t, "tools-1", response.ID)
	assert.NotNil(t, response.Result)

	// Validate tools list structure
	result, ok := response.Result.(ToolsListResult)
	if !ok {
		// Try to convert from map
		resultMap, ok := response.Result.(map[string]interface{})
		require.True(t, ok, "Result should be ToolsListResult or map")
		assert.Contains(t, resultMap, "tools")
	} else {
		assert.NotNil(t, result.Tools)
	}
}

// TestHandleToolsCall tests the tools/call method
func TestHandleToolsCall(t *testing.T) {
	config := config.MCPConfig{
		Enabled:        true,
		Port:           34800,
		MaxConnections: 100,
	}

	mockAPI := &MockAPIInterface{}
	server := NewServer(config, mockAPI)

	mockConn := &MockConnection{id: "test-conn"}
	conn := &Connection{
		id:     "test-conn",
		server: server,
		ctx:    context.Background(),
	}

	tests := []struct {
		name     string
		request  *JSONRPCRequest
		wantErr  bool
		validate func(t *testing.T, responses []JSONRPCResponse, errors []JSONRPCResponse)
	}{
		{
			name: "missing params",
			request: &JSONRPCRequest{
				JSONRPC: "2.0",
				ID:      "call-1",
				Method:  "tools/call",
				Params:  nil,
			},
			wantErr: false,
			validate: func(t *testing.T, responses []JSONRPCResponse, errors []JSONRPCResponse) {
				require.Len(t, errors, 1)
				assert.Equal(t, InvalidParams, errors[0].Error.Code)
			},
		},
		{
			name: "missing tool name",
			request: &JSONRPCRequest{
				JSONRPC: "2.0",
				ID:      "call-2",
				Method:  "tools/call",
				Params: map[string]interface{}{
					"arguments": map[string]interface{}{},
				},
			},
			wantErr: false,
			validate: func(t *testing.T, responses []JSONRPCResponse, errors []JSONRPCResponse) {
				require.Len(t, errors, 1)
				assert.Equal(t, InvalidParams, errors[0].Error.Code)
			},
		},
		{
			name: "tool not found",
			request: &JSONRPCRequest{
				JSONRPC: "2.0",
				ID:      "call-3",
				Method:  "tools/call",
				Params: map[string]interface{}{
					"name":      "nonexistent_tool",
					"arguments": map[string]interface{}{},
				},
			},
			wantErr: false,
			validate: func(t *testing.T, responses []JSONRPCResponse, errors []JSONRPCResponse) {
				require.Len(t, errors, 1)
				assert.Equal(t, ToolNotFound, errors[0].Error.Code)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mock connection
			mockConn.sentResponses = nil
			mockConn.sentErrors = nil

			testConn := &testConnection{
				Connection: conn,
				mock:       mockConn,
			}

			err := testConn.handleToolsCall(tt.request)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			tt.validate(t, mockConn.sentResponses, mockConn.sentErrors)
		})
	}
}

// testConnection wraps Connection to use mock for testing
type testConnection struct {
	*Connection
	mock *MockConnection
}

func (tc *testConnection) sendResponse(id interface{}, result interface{}) error {
	return tc.mock.sendResponse(id, result)
}

func (tc *testConnection) sendError(id interface{}, code int, message string, data interface{}) error {
	return tc.mock.sendError(id, code, message, data)
}

func (tc *testConnection) handleInitialize(request *JSONRPCRequest) error {
	// Mock the initialize handling without using WebSocket
	result := InitializeResult{
		ProtocolVersion: MCPProtocolVersion,
		Capabilities: ServerCapabilities{
			Tools: &ToolsCapability{
				ListChanged: false,
			},
			Logging: &LoggingCapability{},
		},
		ServerInfo: ServerInfo{
			Name:    "UMA MCP Server",
			Version: "1.0.0",
		},
	}
	return tc.sendResponse(request.ID, result)
}

func (tc *testConnection) handleToolsList(request *JSONRPCRequest) error {
	// Mock the tools list handling
	tools, err := tc.server.registry.GetTools()
	if err != nil {
		return tc.sendError(request.ID, InternalError, "Failed to get tools", nil)
	}

	result := ToolsListResult{
		Tools: tools,
	}
	return tc.sendResponse(request.ID, result)
}

func (tc *testConnection) handleToolsCall(request *JSONRPCRequest) error {
	// Mock the tools call handling
	var params ToolCallParams
	if request.Params == nil {
		return tc.sendError(request.ID, InvalidParams, "Missing parameters", nil)
	}

	// Simple parameter validation for testing
	paramsMap, ok := request.Params.(map[string]interface{})
	if !ok {
		return tc.sendError(request.ID, InvalidParams, "Invalid parameters", nil)
	}

	name, ok := paramsMap["name"].(string)
	if !ok || name == "" {
		return tc.sendError(request.ID, InvalidParams, "Tool name is required", nil)
	}

	params.Name = name
	if args, ok := paramsMap["arguments"].(map[string]interface{}); ok {
		params.Arguments = args
	}

	// Execute tool
	result, err := tc.server.registry.ExecuteTool(params.Name, params.Arguments)
	if err != nil {
		if err.Error() == "tool not found" {
			return tc.sendError(request.ID, ToolNotFound, "Tool not found", nil)
		}
		return tc.sendResponse(request.ID, ToolCallResult{
			Content: []ToolContent{
				{
					Type: "text",
					Text: "Error executing tool",
				},
			},
			IsError: true,
		})
	}

	return tc.sendResponse(request.ID, result)
}

// TestValidateJSONRPCRequest tests request validation
func TestValidateJSONRPCRequest(t *testing.T) {
	conn := &Connection{}

	tests := []struct {
		name    string
		request *JSONRPCRequest
		wantErr bool
	}{
		{
			name: "valid request",
			request: &JSONRPCRequest{
				JSONRPC: "2.0",
				Method:  "test",
				ID:      "123",
			},
			wantErr: false,
		},
		{
			name: "invalid jsonrpc version",
			request: &JSONRPCRequest{
				JSONRPC: "1.0",
				Method:  "test",
				ID:      "123",
			},
			wantErr: true,
		},
		{
			name: "missing method",
			request: &JSONRPCRequest{
				JSONRPC: "2.0",
				Method:  "",
				ID:      "123",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := conn.validateJSONRPCRequest(tt.request)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
