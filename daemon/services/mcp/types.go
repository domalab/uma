package mcp

// JSON-RPC 2.0 message types for MCP

// JSONRPCRequest represents a JSON-RPC 2.0 request
type JSONRPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id,omitempty"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

// JSONRPCResponse represents a JSON-RPC 2.0 response
type JSONRPCResponse struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      interface{}   `json:"id,omitempty"`
	Result  interface{}   `json:"result,omitempty"`
	Error   *JSONRPCError `json:"error,omitempty"`
}

// JSONRPCError represents a JSON-RPC 2.0 error
type JSONRPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// MCP-specific types

// InitializeParams represents the parameters for the initialize method
type InitializeParams struct {
	ProtocolVersion string                 `json:"protocolVersion"`
	Capabilities    ClientCapabilities     `json:"capabilities"`
	ClientInfo      ClientInfo             `json:"clientInfo"`
	Meta            map[string]interface{} `json:"meta,omitempty"`
}

// ClientCapabilities represents client capabilities
type ClientCapabilities struct {
	Roots    *RootsCapability    `json:"roots,omitempty"`
	Sampling *SamplingCapability `json:"sampling,omitempty"`
}

// RootsCapability represents roots capability
type RootsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

// SamplingCapability represents sampling capability
type SamplingCapability struct{}

// ClientInfo represents client information
type ClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// InitializeResult represents the result of the initialize method
type InitializeResult struct {
	ProtocolVersion string             `json:"protocolVersion"`
	Capabilities    ServerCapabilities `json:"capabilities"`
	ServerInfo      ServerInfo         `json:"serverInfo"`
}

// ServerCapabilities represents server capabilities
type ServerCapabilities struct {
	Tools     *ToolsCapability     `json:"tools,omitempty"`
	Logging   *LoggingCapability   `json:"logging,omitempty"`
	Prompts   *PromptsCapability   `json:"prompts,omitempty"`
	Resources *ResourcesCapability `json:"resources,omitempty"`
}

// ToolsCapability represents tools capability
type ToolsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

// LoggingCapability represents logging capability
type LoggingCapability struct{}

// PromptsCapability represents prompts capability
type PromptsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

// ResourcesCapability represents resources capability
type ResourcesCapability struct {
	Subscribe   bool `json:"subscribe,omitempty"`
	ListChanged bool `json:"listChanged,omitempty"`
}

// ServerInfo represents server information
type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// Tool represents an MCP tool
type Tool struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	InputSchema ToolSchema `json:"inputSchema"`
}

// ToolSchema represents a tool's input schema
type ToolSchema struct {
	Type       string                 `json:"type"`
	Properties map[string]interface{} `json:"properties,omitempty"`
	Required   []string               `json:"required,omitempty"`
}

// ToolsListResult represents the result of tools/list
type ToolsListResult struct {
	Tools []Tool `json:"tools"`
}

// ToolCallParams represents parameters for tools/call
type ToolCallParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
}

// ToolCallResult represents the result of tools/call
type ToolCallResult struct {
	Content []ToolContent `json:"content"`
	IsError bool          `json:"isError,omitempty"`
}

// ToolContent represents tool execution content
type ToolContent struct {
	Type string      `json:"type"`
	Text string      `json:"text,omitempty"`
	Data interface{} `json:"data,omitempty"`
}

// Error codes for JSON-RPC
const (
	// Standard JSON-RPC error codes
	ParseError     = -32700
	InvalidRequest = -32600
	MethodNotFound = -32601
	InvalidParams  = -32602
	InternalError  = -32603

	// MCP-specific error codes
	ToolNotFound  = -32000
	ToolError     = -32001
	InvalidTool   = -32002
	ResourceError = -32003
	PromptError   = -32004
)

// MCP protocol version
const MCPProtocolVersion = "2024-11-05"
