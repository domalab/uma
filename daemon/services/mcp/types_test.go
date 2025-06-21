package mcp

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestJSONRPCRequest tests JSON-RPC request serialization/deserialization
func TestJSONRPCRequest(t *testing.T) {
	request := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      "test-123",
		Method:  "initialize",
		Params: map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]interface{}{},
		},
	}

	// Test serialization
	data, err := json.Marshal(request)
	require.NoError(t, err)
	assert.Contains(t, string(data), "2.0")
	assert.Contains(t, string(data), "initialize")

	// Test deserialization
	var decoded JSONRPCRequest
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, request.JSONRPC, decoded.JSONRPC)
	assert.Equal(t, request.ID, decoded.ID)
	assert.Equal(t, request.Method, decoded.Method)
	assert.NotNil(t, decoded.Params)
}

// TestJSONRPCResponse tests JSON-RPC response serialization/deserialization
func TestJSONRPCResponse(t *testing.T) {
	response := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      "test-123",
		Result: InitializeResult{
			ProtocolVersion: MCPProtocolVersion,
			Capabilities: ServerCapabilities{
				Tools: &ToolsCapability{
					ListChanged: false,
				},
			},
			ServerInfo: ServerInfo{
				Name:    "UMA MCP Server",
				Version: "1.0.0",
			},
		},
	}

	// Test serialization
	data, err := json.Marshal(response)
	require.NoError(t, err)
	assert.Contains(t, string(data), "2.0")
	assert.Contains(t, string(data), MCPProtocolVersion)

	// Test deserialization
	var decoded JSONRPCResponse
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, response.JSONRPC, decoded.JSONRPC)
	assert.Equal(t, response.ID, decoded.ID)
	assert.NotNil(t, decoded.Result)
}

// TestJSONRPCError tests JSON-RPC error serialization/deserialization
func TestJSONRPCError(t *testing.T) {
	errorResponse := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      "test-123",
		Error: &JSONRPCError{
			Code:    MethodNotFound,
			Message: "Method not found",
			Data:    map[string]interface{}{"method": "unknown_method"},
		},
	}

	// Test serialization
	data, err := json.Marshal(errorResponse)
	require.NoError(t, err)
	assert.Contains(t, string(data), "Method not found")
	assert.Contains(t, string(data), "error")

	// Test deserialization
	var decoded JSONRPCResponse
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, errorResponse.JSONRPC, decoded.JSONRPC)
	assert.Equal(t, errorResponse.ID, decoded.ID)
	assert.NotNil(t, decoded.Error)
	assert.Equal(t, MethodNotFound, decoded.Error.Code)
	assert.Equal(t, "Method not found", decoded.Error.Message)
}

// TestInitializeParams tests initialize parameters
func TestInitializeParams(t *testing.T) {
	params := InitializeParams{
		ProtocolVersion: MCPProtocolVersion,
		Capabilities: ClientCapabilities{
			Roots: &RootsCapability{
				ListChanged: true,
			},
		},
		ClientInfo: ClientInfo{
			Name:    "test-client",
			Version: "1.0.0",
		},
		Meta: map[string]interface{}{
			"custom": "value",
		},
	}

	// Test serialization
	data, err := json.Marshal(params)
	require.NoError(t, err)
	assert.Contains(t, string(data), MCPProtocolVersion)
	assert.Contains(t, string(data), "test-client")

	// Test deserialization
	var decoded InitializeParams
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, params.ProtocolVersion, decoded.ProtocolVersion)
	assert.Equal(t, params.ClientInfo.Name, decoded.ClientInfo.Name)
	assert.NotNil(t, decoded.Capabilities.Roots)
	assert.True(t, decoded.Capabilities.Roots.ListChanged)
}

// TestInitializeResult tests initialize result
func TestInitializeResult(t *testing.T) {
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

	// Test serialization
	data, err := json.Marshal(result)
	require.NoError(t, err)
	assert.Contains(t, string(data), MCPProtocolVersion)
	assert.Contains(t, string(data), "UMA MCP Server")

	// Test deserialization
	var decoded InitializeResult
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, result.ProtocolVersion, decoded.ProtocolVersion)
	assert.Equal(t, result.ServerInfo.Name, decoded.ServerInfo.Name)
	assert.NotNil(t, decoded.Capabilities.Tools)
	assert.NotNil(t, decoded.Capabilities.Logging)
}

// TestTool tests tool definition
func TestTool(t *testing.T) {
	tool := Tool{
		Name:        "get_system_info",
		Description: "Get comprehensive system information",
		InputSchema: ToolSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"include_hardware": map[string]interface{}{
					"type":        "boolean",
					"description": "Include hardware details",
					"default":     true,
				},
			},
			Required: []string{},
		},
	}

	// Test serialization
	data, err := json.Marshal(tool)
	require.NoError(t, err)
	assert.Contains(t, string(data), "get_system_info")
	assert.Contains(t, string(data), "comprehensive system information")

	// Test deserialization
	var decoded Tool
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, tool.Name, decoded.Name)
	assert.Equal(t, tool.Description, decoded.Description)
	assert.Equal(t, tool.InputSchema.Type, decoded.InputSchema.Type)
	assert.Contains(t, decoded.InputSchema.Properties, "include_hardware")
}

// TestToolsListResult tests tools list result
func TestToolsListResult(t *testing.T) {
	tools := []Tool{
		{
			Name:        "tool1",
			Description: "First tool",
			InputSchema: ToolSchema{Type: "object"},
		},
		{
			Name:        "tool2",
			Description: "Second tool",
			InputSchema: ToolSchema{Type: "object"},
		},
	}

	result := ToolsListResult{
		Tools: tools,
	}

	// Test serialization
	data, err := json.Marshal(result)
	require.NoError(t, err)
	assert.Contains(t, string(data), "tool1")
	assert.Contains(t, string(data), "tool2")

	// Test deserialization
	var decoded ToolsListResult
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Len(t, decoded.Tools, 2)
	assert.Equal(t, "tool1", decoded.Tools[0].Name)
	assert.Equal(t, "tool2", decoded.Tools[1].Name)
}

// TestToolCallParams tests tool call parameters
func TestToolCallParams(t *testing.T) {
	params := ToolCallParams{
		Name: "get_system_info",
		Arguments: map[string]interface{}{
			"include_hardware": true,
			"format":           "json",
		},
	}

	// Test serialization
	data, err := json.Marshal(params)
	require.NoError(t, err)
	assert.Contains(t, string(data), "get_system_info")
	assert.Contains(t, string(data), "include_hardware")

	// Test deserialization
	var decoded ToolCallParams
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, params.Name, decoded.Name)
	assert.Equal(t, true, decoded.Arguments["include_hardware"])
	assert.Equal(t, "json", decoded.Arguments["format"])
}

// TestToolCallResult tests tool call result
func TestToolCallResult(t *testing.T) {
	result := ToolCallResult{
		Content: []ToolContent{
			{
				Type: "text",
				Text: "System information retrieved successfully",
			},
			{
				Type: "data",
				Data: map[string]interface{}{
					"cpu_usage": 45.2,
					"memory":    "16GB",
				},
			},
		},
		IsError: false,
	}

	// Test serialization
	data, err := json.Marshal(result)
	require.NoError(t, err)
	assert.Contains(t, string(data), "System information")
	assert.Contains(t, string(data), "cpu_usage")

	// Test deserialization
	var decoded ToolCallResult
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Len(t, decoded.Content, 2)
	assert.Equal(t, "text", decoded.Content[0].Type)
	assert.Equal(t, "data", decoded.Content[1].Type)
	assert.False(t, decoded.IsError)
}

// TestErrorCodes tests error code constants
func TestErrorCodes(t *testing.T) {
	// Test standard JSON-RPC error codes
	assert.Equal(t, -32700, ParseError)
	assert.Equal(t, -32600, InvalidRequest)
	assert.Equal(t, -32601, MethodNotFound)
	assert.Equal(t, -32602, InvalidParams)
	assert.Equal(t, -32603, InternalError)

	// Test MCP-specific error codes
	assert.Equal(t, -32000, ToolNotFound)
	assert.Equal(t, -32001, ToolError)
	assert.Equal(t, -32002, InvalidTool)
	assert.Equal(t, -32003, ResourceError)
	assert.Equal(t, -32004, PromptError)
}

// TestMCPProtocolVersion tests protocol version constant
func TestMCPProtocolVersion(t *testing.T) {
	assert.Equal(t, "2024-11-05", MCPProtocolVersion)
	assert.NotEmpty(t, MCPProtocolVersion)
}

// TestComplexJSONRPCScenarios tests complex JSON-RPC scenarios
func TestComplexJSONRPCScenarios(t *testing.T) {
	// Test request with complex parameters
	complexRequest := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      123, // numeric ID
		Method:  "tools/call",
		Params: ToolCallParams{
			Name: "complex_tool",
			Arguments: map[string]interface{}{
				"nested": map[string]interface{}{
					"array": []interface{}{1, 2, 3},
					"bool":  true,
				},
				"string": "test",
				"number": 42.5,
			},
		},
	}

	data, err := json.Marshal(complexRequest)
	require.NoError(t, err)

	var decoded JSONRPCRequest
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, float64(123), decoded.ID) // JSON numbers become float64
	assert.Equal(t, "tools/call", decoded.Method)
}

// Benchmark tests for performance validation
func BenchmarkJSONRPCRequestMarshal(b *testing.B) {
	request := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      "test-123",
		Method:  "initialize",
		Params: map[string]interface{}{
			"protocolVersion": MCPProtocolVersion,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(request)
	}
}

func BenchmarkJSONRPCResponseMarshal(b *testing.B) {
	response := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      "test-123",
		Result: map[string]interface{}{
			"status": "success",
			"data":   []interface{}{1, 2, 3},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(response)
	}
}
