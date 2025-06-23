package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/domalab/uma/daemon/dto"
	"github.com/domalab/uma/daemon/logger"
	"github.com/domalab/uma/daemon/services/api/adapters"
	"github.com/domalab/uma/daemon/services/api/handlers"
	"github.com/domalab/uma/daemon/services/api/routes"
	"github.com/domalab/uma/daemon/services/api/services"
	"github.com/domalab/uma/daemon/services/api/utils"
	"github.com/domalab/uma/daemon/services/command"
	"github.com/domalab/uma/daemon/services/config"
	"github.com/getsentry/sentry-go"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// Data structures moved to types/ packages

// HTTPServer handles REST API requests
type HTTPServer struct {
	api             *Api
	server          *http.Server
	port            int
	commandExecutor *command.CommandExecutor
	cacheService    *services.CacheService
	configService   *config.ViperConfigService
	validator       *validator.Validate

	// Handler instances
	systemHandler  *handlers.SystemHandler
	storageHandler *handlers.StorageHandler
	dockerHandler  *handlers.DockerHandler
	vmHandler      *handlers.VMHandler
	healthHandler  *handlers.HealthHandler

	webSocketHandler    *handlers.WebSocketHandler
	notificationHandler *handlers.NotificationHandler
	asyncHandler        *handlers.AsyncHandler

	// Router for modular route management
	router *routes.Router

	// API adapter
	apiAdapter utils.APIInterface

	// Services
	shareService  *services.ShareService
	scriptService *services.ScriptService

	// New handlers
	shareHandler       *handlers.ShareHandler
	scriptHandler      *handlers.ScriptHandler
	diagnosticsHandler *handlers.DiagnosticsHandler
	mcpHandler         *handlers.MCPHandler
}

// NewHTTPServer creates a new HTTP server instance
func NewHTTPServer(api *Api, port int) *HTTPServer {
	httpServer := &HTTPServer{
		api:             api,
		port:            port,
		commandExecutor: command.NewCommandExecutor(),
		cacheService:    services.NewCacheService(),
		configService:   config.NewViperConfigService(),
		validator:       validator.New(),
	}

	// Initialize API adapter
	httpServer.apiAdapter = adapters.NewAPIAdapter(api)

	// Initialize services
	httpServer.shareService = services.NewShareService()
	httpServer.scriptService = services.NewScriptService(httpServer.apiAdapter)

	// Initialize new handlers
	httpServer.shareHandler = handlers.NewShareHandler(httpServer.apiAdapter)
	httpServer.scriptHandler = handlers.NewScriptHandler(httpServer.apiAdapter)
	httpServer.diagnosticsHandler = handlers.NewDiagnosticsHandler(httpServer.apiAdapter)
	httpServer.mcpHandler = handlers.NewMCPHandler(httpServer.apiAdapter)

	// Initialize handlers
	httpServer.systemHandler = handlers.NewSystemHandler(httpServer.apiAdapter)
	httpServer.storageHandler = handlers.NewStorageHandler(httpServer.apiAdapter)
	httpServer.dockerHandler = handlers.NewDockerHandler(httpServer.apiAdapter)
	httpServer.vmHandler = handlers.NewVMHandler(httpServer.apiAdapter)
	httpServer.healthHandler = handlers.NewHealthHandler(httpServer.apiAdapter, api.ctx.Config.Version)
	// OpenAPI documentation is now handled directly by HTTPServer (no separate docs handler needed)
	httpServer.webSocketHandler = handlers.NewWebSocketHandler(httpServer.apiAdapter, api.ctx.Hub)
	httpServer.notificationHandler = handlers.NewNotificationHandler(httpServer.apiAdapter)
	httpServer.asyncHandler = handlers.NewAsyncHandler(httpServer.apiAdapter)

	// Legacy handlers removed - functionality moved to modular handlers

	// Initialize router with all handlers
	httpServer.router = routes.NewRouter(
		httpServer.systemHandler,
		httpServer.storageHandler,
		httpServer.dockerHandler,
		httpServer.vmHandler,
		httpServer.healthHandler,
		httpServer.webSocketHandler,
		httpServer.notificationHandler,
		httpServer.asyncHandler,
		httpServer.shareHandler,
		httpServer.scriptHandler,
		httpServer.diagnosticsHandler,
		httpServer.mcpHandler,
		// Removed httpServer parameter - OpenAPI system removed
	)

	return httpServer
}

// HTTPServerInterface implementation methods for legacy handlers

// WriteJSON writes JSON response
func (h *HTTPServer) WriteJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		logger.Yellow("Error encoding JSON response: %v", err)
	}
}

// WriteError writes error response
func (h *HTTPServer) WriteError(w http.ResponseWriter, status int, message string) {
	// Capture error in Sentry for production monitoring
	if status >= 500 {
		sentry.CaptureMessage(fmt.Sprintf("HTTP %d Error: %s", status, message))
	}

	errorResponse := map[string]interface{}{
		"error":   message,
		"message": http.StatusText(status),
	}

	h.WriteJSON(w, status, errorResponse)
}

// Utility methods for backward compatibility

// WriteStandardResponse writes standardized response
func (h *HTTPServer) WriteStandardResponse(w http.ResponseWriter, status int, data interface{}, pagination *dto.PaginationInfo) {
	response := map[string]interface{}{
		"data": data,
	}
	if pagination != nil {
		response["pagination"] = pagination
	}
	h.WriteJSON(w, status, response)
}

// Removed unused function: writeStandardResponse

// ParsePaginationParams parses pagination parameters
func (h *HTTPServer) ParsePaginationParams(r *http.Request) *dto.PaginationParams {
	params := &dto.PaginationParams{}

	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			params.Page = page
		}
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			params.Limit = limit
		}
	}

	return params
}

// GetRequestIDFromContext gets request ID from context
func (h *HTTPServer) GetRequestIDFromContext(r *http.Request) string {
	if requestID := r.Header.Get("X-Request-ID"); requestID != "" {
		return requestID
	}
	return h.generateRequestID()
}

// Removed unused function: getRequestIDFromContext

// generateRequestID generates a new request ID using UUID
func (h *HTTPServer) generateRequestID() string {
	return uuid.New().String()
}

// GetSystemHandler returns the system handler
func (h *HTTPServer) GetSystemHandler() *handlers.SystemHandler {
	return h.systemHandler
}

// GetHealthHandler returns the health handler
func (h *HTTPServer) GetHealthHandler() *handlers.HealthHandler {
	return h.healthHandler
}

// Sentry middleware moved to middleware/sentry.go

// Start starts the HTTP server
func (h *HTTPServer) Start() error {
	// Register all routes using the modular router
	h.router.RegisterRoutes()

	// Get the configured handler with middleware chain
	handler := h.router.GetHandler()

	h.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", h.port),
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	logger.Blue("Starting HTTP API server on port %d", h.port)

	go func() {
		if err := h.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Yellow("HTTP server error: %v", err)
		}
	}()

	// WebSocket handlers are now managed by the modular router

	return nil
}

// Stop gracefully stops the HTTP server
func (h *HTTPServer) Stop() error {
	if h.server == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	logger.Blue("Shutting down HTTP API server...")
	return h.server.Shutdown(ctx)
}

// Legacy handler methods moved to respective handler files

// Removed unused function: getAPIVersionFromContext
