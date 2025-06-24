package routes

import (
	"net/http"

	"github.com/domalab/uma/daemon/services/api/handlers"
	"github.com/domalab/uma/daemon/services/api/middleware"
)

// Router manages all API routes and middleware
type Router struct {
	mux *http.ServeMux

	// Handlers
	systemHandler  *handlers.SystemHandler
	storageHandler *handlers.StorageHandler
	dockerHandler  *handlers.DockerHandler
	vmHandler      *handlers.VMHandler
	healthHandler  *handlers.HealthHandler

	webSocketHandler    *handlers.WebSocketHandler
	notificationHandler *handlers.NotificationHandler
	asyncHandler        *handlers.AsyncHandler
	shareHandler        *handlers.ShareHandler
	scriptHandler       *handlers.ScriptHandler
	diagnosticsHandler  *handlers.DiagnosticsHandler
	mcpHandler          *handlers.MCPHandler

	// Removed httpServer field - no longer needed for OpenAPI
}

// NewRouter creates a new router with all handlers
func NewRouter(
	systemHandler *handlers.SystemHandler,
	storageHandler *handlers.StorageHandler,
	dockerHandler *handlers.DockerHandler,
	vmHandler *handlers.VMHandler,
	healthHandler *handlers.HealthHandler,
	webSocketHandler *handlers.WebSocketHandler,
	notificationHandler *handlers.NotificationHandler,
	asyncHandler *handlers.AsyncHandler,
	shareHandler *handlers.ShareHandler,
	scriptHandler *handlers.ScriptHandler,
	diagnosticsHandler *handlers.DiagnosticsHandler,
	mcpHandler *handlers.MCPHandler,
) *Router {
	return &Router{
		mux:                 http.NewServeMux(),
		systemHandler:       systemHandler,
		storageHandler:      storageHandler,
		dockerHandler:       dockerHandler,
		vmHandler:           vmHandler,
		healthHandler:       healthHandler,
		webSocketHandler:    webSocketHandler,
		notificationHandler: notificationHandler,
		asyncHandler:        asyncHandler,
		shareHandler:        shareHandler,
		scriptHandler:       scriptHandler,
		diagnosticsHandler:  diagnosticsHandler,
		mcpHandler:          mcpHandler,
	}
}

// RegisterRoutes registers all API routes
func (r *Router) RegisterRoutes() {
	// Register routes by domain
	r.registerHealthRoutes()
	// Removed registerDocsRoutes() - OpenAPI documentation system removed
	r.registerAsyncRoutes()
	r.registerMetricsRoutes()
	r.registerSystemRoutes()
	r.registerStorageRoutes()
	r.registerDockerRoutes()
	r.registerVMRoutes()
	r.registerWebSocketRoutes()
	r.registerNotificationRoutes()
	r.registerShareRoutes()
	r.registerScriptRoutes()
	r.registerDiagnosticsRoutes()
	r.registerMCPRoutes()
}

// GetHandler returns the configured handler with middleware chain
func (r *Router) GetHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// Check if this is a WebSocket endpoint - bypass middleware for WebSocket upgrades
		if isWebSocketEndpoint(req.URL.Path) {
			r.mux.ServeHTTP(w, req)
			return
		}

		// Build middleware chain for non-WebSocket endpoints
		handler := http.Handler(r.mux)
		handler = middleware.RequestID()(handler)
		handler = middleware.Versioning()(handler)
		handler = middleware.Compression()(handler)
		handler = middleware.Metrics()(handler)
		handler = middleware.Logging()(handler)
		handler = middleware.Sentry()(handler) // Add Sentry error tracking

		handler.ServeHTTP(w, req)
	})
}

// isWebSocketEndpoint checks if the path is a WebSocket endpoint
func isWebSocketEndpoint(path string) bool {
	wsEndpoints := []string{
		"/api/v1/ws",  // Real-time monitoring WebSocket
		"/api/v1/mcp", // MCP JSON-RPC 2.0 WebSocket
	}

	for _, endpoint := range wsEndpoints {
		if path == endpoint {
			return true
		}
	}
	return false
}

// registerHealthRoutes registers health and documentation endpoints
func (r *Router) registerHealthRoutes() {
	r.mux.HandleFunc("/api/v1/health", r.healthHandler.HandleHealth)
}

// Removed OpenAPIProvider interface and registerDocsRoutes function - OpenAPI system removed

// registerAsyncRoutes registers async operation endpoints
func (r *Router) registerAsyncRoutes() {
	r.mux.HandleFunc("/api/v1/operations", r.asyncHandler.HandleAsyncOperations)
	r.mux.HandleFunc("/api/v1/operations/", r.asyncHandler.HandleAsyncOperation)
	r.mux.HandleFunc("/api/v1/operations/stats", r.asyncHandler.HandleAsyncStats)
}

// Metrics endpoint
func (r *Router) registerMetricsRoutes() {
	r.mux.HandleFunc("/metrics", middleware.GetMetricsHandler().ServeHTTP)
}

// registerNotificationRoutes registers notification endpoints
func (r *Router) registerNotificationRoutes() {
	r.mux.HandleFunc("/api/v1/notifications", r.notificationHandler.HandleNotifications)
	r.mux.HandleFunc("/api/v1/notifications/", r.notificationHandler.HandleNotification)
	r.mux.HandleFunc("/api/v1/notifications/clear", r.notificationHandler.HandleNotificationsClear)
	r.mux.HandleFunc("/api/v1/notifications/stats", r.notificationHandler.HandleNotificationsStats)
	r.mux.HandleFunc("/api/v1/notifications/mark-all-read", r.notificationHandler.HandleNotificationsMarkAllRead)
}

// registerShareRoutes registers share management endpoints
func (r *Router) registerShareRoutes() {
	r.mux.HandleFunc("/api/v1/shares", r.shareHandler.HandleShares)
	r.mux.HandleFunc("/api/v1/shares/", r.shareHandler.HandleShare)
}

// registerScriptRoutes registers script management endpoints
func (r *Router) registerScriptRoutes() {
	r.mux.HandleFunc("/api/v1/scripts", r.scriptHandler.HandleScriptsList)
	r.mux.HandleFunc("/api/v1/scripts/", r.scriptHandler.HandleScript)
}

// registerDiagnosticsRoutes registers diagnostics endpoints
func (r *Router) registerDiagnosticsRoutes() {
	r.mux.HandleFunc("/api/v1/diagnostics/health", r.diagnosticsHandler.HandleDiagnosticsHealth)
	r.mux.HandleFunc("/api/v1/diagnostics/info", r.diagnosticsHandler.HandleDiagnosticsInfo)
	r.mux.HandleFunc("/api/v1/diagnostics/repair", r.diagnosticsHandler.HandleDiagnosticsRepair)
}

// registerMCPRoutes registers MCP (Model Context Protocol) endpoints
func (r *Router) registerMCPRoutes() {
	// MCP server status and management
	r.mux.HandleFunc("GET /api/v1/mcp/status", r.mcpHandler.GetMCPStatus)
	r.mux.HandleFunc("GET /api/v1/mcp/config", r.mcpHandler.GetMCPConfig)
	r.mux.HandleFunc("PUT /api/v1/mcp/config", r.mcpHandler.UpdateMCPConfig)

	// MCP tools management
	r.mux.HandleFunc("GET /api/v1/mcp/tools", r.mcpHandler.GetMCPTools)
	r.mux.HandleFunc("GET /api/v1/mcp/tools/categories", r.mcpHandler.GetMCPToolsByCategory)
	r.mux.HandleFunc("POST /api/v1/mcp/tools/refresh", r.mcpHandler.RefreshMCPTools)

	// Note: MCP WebSocket endpoint (/api/v1/mcp) is registered in registerWebSocketRoutes()
}
