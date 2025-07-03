package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/domalab/uma/daemon/logger"
	restapi "github.com/domalab/uma/daemon/services/api/rest"
	"github.com/domalab/uma/daemon/services/api/services"
	"github.com/domalab/uma/daemon/services/collectors"
	"github.com/domalab/uma/daemon/services/command"
	"github.com/domalab/uma/daemon/services/config"
	"github.com/domalab/uma/daemon/services/streaming"
	"github.com/go-playground/validator/v10"
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

	// UMA v2 Components - Pure v2 implementation
	v2Collector  *collectors.SystemCollector
	v2Streamer   *streaming.WebSocketEngine
	v2RESTServer *restapi.RESTServer
}

// NewHTTPServer creates a new HTTP server instance - UMA v2 only
func NewHTTPServer(api *Api, port int) *HTTPServer {
	httpServer := &HTTPServer{
		api:             api,
		port:            port,
		commandExecutor: command.NewCommandExecutor(),
		cacheService:    services.NewCacheService(),
		configService:   config.NewViperConfigService(),
		validator:       validator.New(),
	}

	// Initialize v2 components only - no v1 compatibility
	httpServer.v2Collector = collectors.NewSystemCollector()
	httpServer.v2Streamer = streaming.NewWebSocketEngine(httpServer.v2Collector)
	httpServer.v2RESTServer = restapi.NewRESTServer(httpServer.v2Collector, httpServer.v2Streamer)

	return httpServer
}

// Start starts the HTTP server - UMA v2 only
func (h *HTTPServer) Start() error {
	// Start v2 collector
	if err := h.v2Collector.Start(); err != nil {
		logger.Red("Failed to start v2 collector: %v", err)
		return err
	}

	// UMA v2 API - Pure v2 implementation without v1 compatibility
	h.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", h.port),
		Handler:      h.v2RESTServer,
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
