package mcp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/domalab/uma/daemon/services/config"
	"github.com/stretchr/testify/assert"
)

// TestNewServer tests MCP server creation
func TestNewServer(t *testing.T) {
	config := config.MCPConfig{
		Enabled:        true,
		Port:           34800,
		MaxConnections: 100,
	}

	mockAPI := &MockAPIInterface{}
	server := NewServer(config, mockAPI)

	assert.NotNil(t, server)
	assert.Equal(t, config.Enabled, server.config.Enabled)
	assert.Equal(t, config.Port, server.config.Port)
	assert.Equal(t, config.MaxConnections, server.config.MaxConnections)
	assert.NotNil(t, server.registry)
	assert.NotNil(t, server.connections)
	assert.NotNil(t, server.ctx)
}

// TestServerStartStop tests server lifecycle
func TestServerStartStop(t *testing.T) {
	config := config.MCPConfig{
		Enabled:        true,
		Port:           0, // Use random port for testing
		MaxConnections: 10,
	}

	mockAPI := &MockAPIInterface{}
	server := NewServer(config, mockAPI)

	// Test start with disabled server
	config.Enabled = false
	server.config = config
	err := server.Start()
	assert.NoError(t, err)

	// Test start with enabled server
	config.Enabled = true
	config.Port = 0 // Random port
	server.config = config
	err = server.Start()
	assert.NoError(t, err)

	// Test stop
	err = server.Stop()
	assert.NoError(t, err)

	// Test stop when server is nil
	server.server = nil
	err = server.Stop()
	assert.NoError(t, err)
}

// TestWebSocketUpgrade tests WebSocket connection upgrade
func TestWebSocketUpgrade(t *testing.T) {
	// Skip this test as it requires real WebSocket infrastructure
	// which is not suitable for unit testing environment
	t.Skip("WebSocket upgrade test requires real infrastructure - tested in integration tests")
}

// TestConnectionLimit tests maximum connection limit
func TestConnectionLimit(t *testing.T) {
	config := config.MCPConfig{
		Enabled:        true,
		Port:           34800,
		MaxConnections: 1, // Very low limit for testing
	}

	mockAPI := &MockAPIInterface{}
	server := NewServer(config, mockAPI)

	// Simulate max connections reached
	server.mutex.Lock()
	server.connections["test-conn-1"] = &Connection{id: "test-conn-1"}
	server.mutex.Unlock()

	// Create test request
	req := httptest.NewRequest("GET", "/mcp", nil)
	req.Header.Set("Connection", "upgrade")
	req.Header.Set("Upgrade", "websocket")
	req.Header.Set("Sec-WebSocket-Key", "test-key")
	req.Header.Set("Sec-WebSocket-Version", "13")

	w := httptest.NewRecorder()
	server.handleWebSocket(w, req)

	// Should return 503 Service Unavailable
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}

// TestGenerateConnectionID tests connection ID generation
func TestGenerateConnectionID(t *testing.T) {
	id1 := generateConnectionID()
	id2 := generateConnectionID()

	assert.NotEmpty(t, id1)
	assert.NotEmpty(t, id2)
	assert.NotEqual(t, id1, id2)
	assert.True(t, strings.HasPrefix(id1, "mcp-"))
	assert.True(t, strings.HasPrefix(id2, "mcp-"))
}

// TestGetServerStats tests server statistics
func TestGetServerStats(t *testing.T) {
	config := config.MCPConfig{
		Enabled:        true,
		Port:           34800,
		MaxConnections: 100,
	}

	mockAPI := &MockAPIInterface{}
	server := NewServer(config, mockAPI)

	// Add mock connection
	server.mutex.Lock()
	mockConn := &Connection{
		id:  "test-conn",
		ctx: context.Background(),
	}
	server.connections["test-conn"] = mockConn
	server.mutex.Unlock()

	stats := server.GetServerStats()

	assert.NotNil(t, stats)
	assert.Equal(t, true, stats["enabled"])
	assert.Equal(t, 34800, stats["port"])
	assert.Equal(t, 100, stats["max_connections"])
	assert.Equal(t, 1, stats["active_connections"])
	assert.Contains(t, stats, "connections")
	assert.Contains(t, stats, "total_tools")
}

// TestConnectionClose tests connection cleanup
func TestConnectionClose(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	conn := &Connection{
		id:     "test-conn",
		ctx:    ctx,
		cancel: cancel,
	}

	// Test close
	conn.Close()

	// Verify context is canceled
	select {
	case <-conn.ctx.Done():
		// Expected
	default:
		t.Error("Expected context to be canceled")
	}
}

// TestConnectionGetStats tests connection statistics
func TestConnectionGetStats(t *testing.T) {
	ctx := context.Background()

	// Mock WebSocket connection for testing
	// Note: In real tests, this would use a proper WebSocket connection
	conn := &Connection{
		id:  "test-conn-123",
		ctx: ctx,
	}

	stats := conn.GetConnectionStats()

	assert.NotNil(t, stats)
	assert.Equal(t, "test-conn-123", stats["id"])
	assert.Equal(t, true, stats["connected"])
}

// TestServerGetRegistry tests registry access
func TestServerGetRegistry(t *testing.T) {
	config := config.MCPConfig{
		Enabled:        true,
		Port:           34800,
		MaxConnections: 100,
	}

	mockAPI := &MockAPIInterface{}
	server := NewServer(config, mockAPI)

	registry := server.GetRegistry()
	assert.NotNil(t, registry)
	assert.IsType(t, &SimpleToolRegistry{}, registry)
}

// Benchmark tests for performance validation
func BenchmarkNewServer(b *testing.B) {
	config := config.MCPConfig{
		Enabled:        true,
		Port:           34800,
		MaxConnections: 100,
	}
	mockAPI := &MockAPIInterface{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewServer(config, mockAPI)
	}
}

func BenchmarkGenerateConnectionID(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = generateConnectionID()
	}
}
