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
	systemHandler       *handlers.SystemHandler
	storageHandler      *handlers.StorageHandler
	dockerHandler       *handlers.DockerHandler
	vmHandler           *handlers.VMHandler
	authHandler         *handlers.AuthHandler
	healthHandler       *handlers.HealthHandler
	docsHandler         *handlers.DocsHandler
	websocketHandler    *handlers.WebSocketHandler
	notificationHandler *handlers.NotificationHandler
	asyncHandler        *handlers.AsyncHandler
	rateLimitHandler    *handlers.RateLimitHandler
	shareHandler        *handlers.ShareHandler
	scriptHandler       *handlers.ScriptHandler
	diagnosticsHandler  *handlers.DiagnosticsHandler

	// Legacy handlers for gradual migration
	httpServer interface{} // Will be *HTTPServer during transition
}

// NewRouter creates a new router with all handlers
func NewRouter(
	systemHandler *handlers.SystemHandler,
	storageHandler *handlers.StorageHandler,
	dockerHandler *handlers.DockerHandler,
	vmHandler *handlers.VMHandler,
	authHandler *handlers.AuthHandler,
	healthHandler *handlers.HealthHandler,
	docsHandler *handlers.DocsHandler,
	websocketHandler *handlers.WebSocketHandler,
	notificationHandler *handlers.NotificationHandler,
	asyncHandler *handlers.AsyncHandler,
	rateLimitHandler *handlers.RateLimitHandler,
	shareHandler *handlers.ShareHandler,
	scriptHandler *handlers.ScriptHandler,
	diagnosticsHandler *handlers.DiagnosticsHandler,
	httpServer interface{}, // Legacy server for transition
) *Router {
	return &Router{
		mux:                 http.NewServeMux(),
		systemHandler:       systemHandler,
		storageHandler:      storageHandler,
		dockerHandler:       dockerHandler,
		vmHandler:           vmHandler,
		authHandler:         authHandler,
		healthHandler:       healthHandler,
		docsHandler:         docsHandler,
		websocketHandler:    websocketHandler,
		notificationHandler: notificationHandler,
		asyncHandler:        asyncHandler,
		rateLimitHandler:    rateLimitHandler,
		shareHandler:        shareHandler,
		scriptHandler:       scriptHandler,
		diagnosticsHandler:  diagnosticsHandler,
		httpServer:          httpServer,
	}
}

// RegisterRoutes registers all API routes
func (r *Router) RegisterRoutes() {
	// Register routes by domain
	r.registerHealthRoutes()
	r.registerDocsRoutes()
	r.registerAsyncRoutes()
	r.registerRateLimitRoutes()
	r.registerSystemRoutes()
	r.registerStorageRoutes()
	r.registerDockerRoutes()
	r.registerVMRoutes()
	r.registerAuthRoutes()
	r.registerWebSocketRoutes()
	r.registerNotificationRoutes()
	r.registerShareRoutes()
	r.registerScriptRoutes()
	r.registerDiagnosticsRoutes()
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
		handler := middleware.CORS()(r.mux)
		handler = middleware.RequestID()(handler)
		handler = middleware.Versioning()(handler)
		handler = middleware.Compression()(handler)
		handler = middleware.Metrics()(handler)
		handler = middleware.Logging()(handler)
		handler = middleware.Sentry()(handler) // Add Sentry error tracking
		// Authentication middleware ready for production (temporarily disabled for testing)
		// handler = middleware.Auth(authService)(handler)

		handler.ServeHTTP(w, req)
	})
}

// isWebSocketEndpoint checks if the path is a WebSocket endpoint
func isWebSocketEndpoint(path string) bool {
	wsEndpoints := []string{
		"/api/v1/ws/system/stats",
		"/api/v1/ws/docker/events",
		"/api/v1/ws/storage/status",
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

// registerDocsRoutes registers documentation endpoints
func (r *Router) registerDocsRoutes() {
	r.mux.HandleFunc("/api/v1/docs", r.docsHandler.SwaggerUIHandler)
	r.mux.HandleFunc("/api/v1/openapi.json", r.docsHandler.OpenAPIHandler)
}

// registerAsyncRoutes registers async operation endpoints
func (r *Router) registerAsyncRoutes() {
	r.mux.HandleFunc("/api/v1/operations", r.asyncHandler.HandleAsyncOperations)
	r.mux.HandleFunc("/api/v1/operations/", r.asyncHandler.HandleAsyncOperation)
	r.mux.HandleFunc("/api/v1/operations/stats", r.asyncHandler.HandleAsyncStats)
}

// registerRateLimitRoutes registers rate limiting endpoints
func (r *Router) registerRateLimitRoutes() {
	r.mux.HandleFunc("/api/v1/rate-limits/stats", r.rateLimitHandler.HandleRateLimitStats)
	r.mux.HandleFunc("/api/v1/rate-limits/config", r.rateLimitHandler.HandleRateLimitConfig)

	// Metrics endpoint
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
