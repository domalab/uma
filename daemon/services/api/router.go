package api

import (
	"net/http"
	"time"

	"github.com/domalab/uma/daemon/services/auth"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// setupChiRouter creates and configures the Chi router
func (h *HTTPServer) setupChiRouter() *chi.Mux {
	r := chi.NewRouter()

	// Add Chi middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// Add custom middleware in the correct order
	r.Use(h.corsMiddleware)
	r.Use(h.requestIDMiddleware)
	r.Use(h.versioningMiddleware)
	r.Use(h.compressionMiddleware)
	r.Use(h.metricsMiddleware)
	r.Use(h.loggingMiddleware)

	// Setup route groups
	h.setupAPIRoutes(r)
	h.setupAuthRoutes(r)
	h.setupDocumentationRoutes(r)
	h.setupMetricsRoutes(r)
	h.setupWebSocketRoutes(r)

	return r
}

// setupAPIRoutes configures the main API routes
func (h *HTTPServer) setupAPIRoutes(r *chi.Mux) {
	r.Route("/api/v1", func(r chi.Router) {
		// Health endpoint (no auth required)
		r.Get("/health", h.handleHealth)

		// System endpoints
		r.Route("/system", func(r chi.Router) {
			r.Use(h.authMiddleware) // Apply auth to system routes
			r.Get("/info", h.handleSystemInfo)
			r.Get("/logs", h.handleSystemLogs)
			r.Get("/origin", h.handleSystemOrigin)
			r.Get("/resources", h.handleSystemResources)
			r.Get("/cpu", h.handleSystemCPU)
			r.Get("/memory", h.handleSystemMemory)
			r.Get("/temperature", h.handleSystemTemperature)
			r.Get("/network", h.handleSystemNetwork)
			r.Get("/ups", h.handleSystemUPS)
			r.Get("/gpu", h.handleSystemGPU)
			r.Get("/filesystems", h.handleSystemFilesystems)
		})

		// Storage endpoints
		r.Route("/storage", func(r chi.Router) {
			r.Use(h.authMiddleware)
			r.Get("/array", h.handleStorageArray)
			r.Get("/cache", h.handleStorageCache)
			r.Get("/boot", h.handleStorageBoot)
			r.Get("/general", h.handleStorageGeneral)
			r.Get("/disks", h.handleStorageDisks)
		})

		// Array control endpoints
		r.Route("/array", func(r chi.Router) {
			r.Use(h.authMiddleware)
			r.Use(h.requirePermission("array.manage"))
			r.Post("/start", h.handleArrayStart)
			r.Post("/stop", h.handleArrayStop)
			r.Post("/parity-check", h.handleArrayParityCheck)

			r.Route("/disk", func(r chi.Router) {
				r.Post("/add", h.handleArrayDiskAdd)
				r.Delete("/remove", h.handleArrayDiskRemove)
			})
		})

		// System power management endpoints
		r.Route("/system/power", func(r chi.Router) {
			r.Use(h.authMiddleware)
			r.Use(h.requirePermission("system.power"))
			r.Post("/shutdown", h.handleSystemShutdown)
			r.Post("/reboot", h.handleSystemReboot)
			r.Post("/sleep", h.handleSystemSleep)
			r.Post("/wake", h.handleSystemWake)
		})

		// Docker endpoints
		r.Route("/docker", func(r chi.Router) {
			r.Use(h.authMiddleware)

			// Container listing (read permission)
			r.With(h.requirePermission("read.docker")).Get("/containers", h.handleDockerContainers)
			r.With(h.requirePermission("read.docker")).Get("/networks", h.handleDockerNetworks)
			r.With(h.requirePermission("read.docker")).Get("/images", h.handleDockerImages)
			r.With(h.requirePermission("read.docker")).Get("/info", h.handleDockerInfo)

			// Container operations (use existing bulk operations)
			r.Route("/containers/bulk", func(r chi.Router) {
				r.Use(h.requirePermission("docker.manage"))
				r.Post("/start", h.handleDockerBulkStart)
				r.Post("/stop", h.handleDockerBulkStop)
				r.Post("/restart", h.handleDockerBulkRestart)
			})
		})

		// VM endpoints
		r.Route("/vms", func(r chi.Router) {
			r.Use(h.authMiddleware)
			r.With(h.requirePermission("read.vms")).Get("/", h.handleVMList)
		})

		// GPU endpoints
		r.Route("/gpu", func(r chi.Router) {
			r.Use(h.authMiddleware)
			r.With(h.requirePermission("read.system")).Get("/", h.handleGPU)
		})

		// Configuration endpoints
		r.Route("/config", func(r chi.Router) {
			r.Use(h.authMiddleware)
			r.With(h.requirePermission("read.config")).Get("/", h.handleConfig)
		})

		// Notifications endpoints
		r.Route("/notifications", func(r chi.Router) {
			r.Use(h.authMiddleware)
			r.With(h.requirePermission("read.notifications")).Get("/", h.handleGetNotifications)
			r.With(h.requirePermission("notifications.manage")).Post("/", h.handleCreateNotification)
			r.With(h.requirePermission("notifications.manage")).Post("/clear", h.handleNotificationsClear)
			r.With(h.requirePermission("read.notifications")).Get("/stats", h.handleNotificationsStats)
			r.With(h.requirePermission("notifications.manage")).Post("/mark-all-read", h.handleNotificationsMarkAllRead)
		})
	})
}

// setupAuthRoutes configures authentication-related routes
func (h *HTTPServer) setupAuthRoutes(r *chi.Mux) {
	r.Route("/api/v1/auth", func(r chi.Router) {
		// Public auth endpoints
		r.Post("/login", h.handleAuthLogin)
		r.Post("/token", h.handleAuthToken)

		// Protected auth endpoints
		r.Group(func(r chi.Router) {
			r.Use(h.authMiddleware)
			r.Use(h.requirePermission("user.manage"))

			r.Get("/users", h.handleAuthUsers)
			r.Post("/users", h.handleAuthCreateUser)
			r.Get("/users/{id}", h.handleAuthGetUser)
			r.Put("/users/{id}", h.handleAuthUpdateUser)
			r.Delete("/users/{id}", h.handleAuthDeleteUser)
			r.Post("/users/{id}/regenerate-key", h.handleAuthRegenerateKey)
		})

		// Auth stats (admin only)
		r.With(h.authMiddleware, h.requirePermission("user.manage")).Get("/stats", h.handleAuthStats)
	})
}

// setupDocumentationRoutes configures documentation routes
func (h *HTTPServer) setupDocumentationRoutes(r *chi.Mux) {
	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/docs", h.handleSwaggerUI)
		r.Get("/openapi.json", h.handleOpenAPISpec)
	})
}

// setupMetricsRoutes configures metrics routes
func (h *HTTPServer) setupMetricsRoutes(r *chi.Mux) {
	r.Get("/metrics", h.handleMetrics)
}

// setupWebSocketRoutes configures WebSocket routes
func (h *HTTPServer) setupWebSocketRoutes(r *chi.Mux) {
	r.Route("/api/v1/ws", func(r chi.Router) {
		// WebSocket endpoints with optional auth
		r.Get("/system/stats", h.handleSystemStatsWebSocket)
		r.Get("/docker/events", h.handleDockerEventsWebSocket)
		r.Get("/storage/status", h.handleStorageStatusWebSocket)
	})
}

// authMiddleware wraps the auth service middleware for Chi
func (h *HTTPServer) authMiddleware(next http.Handler) http.Handler {
	if h.authService != nil {
		return h.authService.AuthMiddleware(next)
	}
	return next
}

// requirePermission creates a middleware that checks for specific permissions
func (h *HTTPServer) requirePermission(permission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip permission check if auth is disabled
			if h.authService == nil || !h.authService.IsEnabled() {
				next.ServeHTTP(w, r)
				return
			}

			user := auth.GetUserFromContext(r)
			if user == nil {
				h.writeError(w, http.StatusUnauthorized, "Authentication required")
				return
			}

			if !h.authService.HasPermission(user, permission) {
				h.writeError(w, http.StatusForbidden, "Insufficient permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// Note: The existing middleware functions in http_server.go are used directly
// Chi router will call them through the Use() method
