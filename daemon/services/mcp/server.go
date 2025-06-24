package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/domalab/uma/daemon/logger"
	"github.com/domalab/uma/daemon/services/api/utils"
	"github.com/domalab/uma/daemon/services/config"
	"github.com/gorilla/websocket"
)

// Server represents the MCP (Model Context Protocol) server
type Server struct {
	config      config.MCPConfig
	api         utils.APIInterface
	registry    *ToolRegistry
	upgrader    websocket.Upgrader
	connections map[string]*Connection
	mutex       sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
	server      *http.Server
}

// Connection represents an active MCP WebSocket connection
type Connection struct {
	id     string
	conn   *websocket.Conn
	ctx    context.Context
	cancel context.CancelFunc
	mutex  sync.RWMutex
	server *Server
}

// NewServer creates a new MCP server instance
func NewServer(config config.MCPConfig, api utils.APIInterface) *Server {
	ctx, cancel := context.WithCancel(context.Background())

	server := &Server{
		config:      config,
		api:         api,
		registry:    NewToolRegistry(api),
		connections: make(map[string]*Connection),
		ctx:         ctx,
		cancel:      cancel,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins for now
			},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
	}

	return server
}

// Start starts the MCP server
func (s *Server) Start() error {
	if !s.config.Enabled {
		logger.Info("MCP server is disabled")
		return nil
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/mcp", s.handleWebSocket)

	s.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", s.config.Port),
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	logger.Blue("Starting MCP server on port %d", s.config.Port)

	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Yellow("MCP server error: %v", err)
		}
	}()

	return nil
}

// Stop stops the MCP server
func (s *Server) Stop() error {
	if s.server == nil {
		return nil
	}

	logger.Blue("Stopping MCP server...")

	// Cancel context to signal shutdown
	s.cancel()

	// Close all connections
	s.mutex.Lock()
	for _, conn := range s.connections {
		conn.Close()
	}
	s.mutex.Unlock()

	// Shutdown HTTP server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil {
		logger.Yellow("Error shutting down MCP server: %v", err)
		return err
	}

	logger.Blue("MCP server stopped")
	return nil
}

// handleWebSocket handles WebSocket connections for MCP
func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Check connection limit
	s.mutex.RLock()
	connectionCount := len(s.connections)
	s.mutex.RUnlock()

	if connectionCount >= s.config.MaxConnections {
		http.Error(w, "Maximum connections reached", http.StatusServiceUnavailable)
		return
	}

	// Upgrade connection
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Yellow("Failed to upgrade WebSocket connection: %v", err)
		return
	}

	// Create connection instance
	connectionID := generateConnectionID()
	ctx, cancel := context.WithCancel(s.ctx)

	mcpConn := &Connection{
		id:     connectionID,
		conn:   conn,
		ctx:    ctx,
		cancel: cancel,
		server: s,
	}

	// Register connection
	s.mutex.Lock()
	s.connections[connectionID] = mcpConn
	s.mutex.Unlock()

	logger.Green("New MCP connection established: %s", connectionID)

	// Handle connection
	mcpConn.Handle()

	// Cleanup on disconnect
	s.mutex.Lock()
	delete(s.connections, connectionID)
	s.mutex.Unlock()

	logger.Yellow("MCP connection closed: %s", connectionID)
}

// Handle handles the MCP connection lifecycle
func (c *Connection) Handle() {
	defer c.Close()

	// Set up ping/pong handling
	c.conn.SetPongHandler(func(string) error {
		if err := c.conn.SetReadDeadline(time.Now().Add(60 * time.Second)); err != nil {
			logger.Warn("Failed to set read deadline: %v", err)
		}
		return nil
	})

	// Start ping ticker
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	go func() {
		for {
			select {
			case <-ticker.C:
				c.mutex.Lock()
				if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					c.mutex.Unlock()
					return
				}
				c.mutex.Unlock()
			case <-c.ctx.Done():
				return
			}
		}
	}()

	// Message handling loop
	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			messageType, data, err := c.conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					logger.Yellow("WebSocket error: %v", err)
				}
				return
			}

			if messageType == websocket.TextMessage {
				if err := c.processMessage(data); err != nil {
					logger.Yellow("Error processing MCP message: %v", err)
				}
			}
		}
	}
}

// Close closes the MCP connection
func (c *Connection) Close() {
	c.cancel()
	c.conn.Close()
}

// processMessage processes incoming MCP messages
func (c *Connection) processMessage(data []byte) error {
	var request JSONRPCRequest
	if err := json.Unmarshal(data, &request); err != nil {
		return c.sendError(nil, -32700, "Parse error", nil)
	}

	// Handle different MCP methods
	switch request.Method {
	case "initialize":
		return c.handleInitialize(&request)
	case "tools/list":
		return c.handleToolsList(&request)
	case "tools/call":
		return c.handleToolsCall(&request)
	default:
		return c.sendError(request.ID, -32601, "Method not found", nil)
	}
}

// sendResponse sends a JSON-RPC response
func (c *Connection) sendResponse(id interface{}, result interface{}) error {
	response := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}

	data, err := json.Marshal(response)
	if err != nil {
		return err
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.conn.WriteMessage(websocket.TextMessage, data)
}

// sendError sends a JSON-RPC error response
func (c *Connection) sendError(id interface{}, code int, message string, data interface{}) error {
	response := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &JSONRPCError{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}

	responseData, err := json.Marshal(response)
	if err != nil {
		return err
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.conn.WriteMessage(websocket.TextMessage, responseData)
}

// generateConnectionID generates a unique connection ID
func generateConnectionID() string {
	return fmt.Sprintf("mcp-%d", time.Now().UnixNano())
}

// GetRegistry returns the tool registry
func (s *Server) GetRegistry() *ToolRegistry {
	return s.registry
}
