package handlers

import (
	"net/http"
	"time"

	"github.com/domalab/uma/daemon/services/api/services"
	"github.com/domalab/uma/daemon/services/api/types/responses"
	"github.com/domalab/uma/daemon/services/api/utils"
)

// HealthHandler handles health check HTTP requests
type HealthHandler struct {
	api                 utils.APIInterface
	version             string
	startTime           time.Time
	readinessChecker    *utils.ProductionReadinessChecker
	monitoringCollector *utils.MonitoringCollector
	healthService       *services.HealthService
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(api utils.APIInterface, version string) *HealthHandler {
	return &HealthHandler{
		api:                 api,
		version:             version,
		startTime:           time.Now(),
		readinessChecker:    utils.NewProductionReadinessChecker(),
		monitoringCollector: utils.NewMonitoringCollector(),
		healthService:       services.NewHealthService(api, version),
	}
}

// HandleHealth handles GET /api/v1/health
func (h *HealthHandler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Calculate uptime in seconds
	uptime := time.Since(h.startTime)
	uptimeSeconds := int(uptime.Seconds())

	// Perform health checks
	checks := h.performHealthChecks()

	// Determine overall status
	status := h.determineOverallStatus(checks)

	utils.WriteHealthResponse(w, status, h.version, uptimeSeconds, checks)
}

// GetHealthStatus returns comprehensive health status using the health service
func (h *HealthHandler) GetHealthStatus() *services.HealthStatus {
	return h.healthService.GetHealthStatus()
}

// HandleHealthLive handles GET /api/v1/health/live (liveness probe)
func (h *HealthHandler) HandleHealthLive(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Simple liveness check - just return OK if the service is running
	response := map[string]interface{}{
		"status":    "alive",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	utils.WriteJSON(w, http.StatusOK, response)
}

// HandleHealthReady handles GET /api/v1/health/ready (readiness probe)
func (h *HealthHandler) HandleHealthReady(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Perform readiness checks
	checks := h.performReadinessChecks()
	ready := h.isSystemReady(checks)

	status := http.StatusOK
	if !ready {
		status = http.StatusServiceUnavailable
	}

	response := map[string]interface{}{
		"status":    map[string]bool{"ready": ready},
		"checks":    checks,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	utils.WriteJSON(w, status, response)
}

// HandleHealthProduction handles GET /api/v1/health/production (comprehensive production health)
func (h *HealthHandler) HandleHealthProduction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Use the production readiness checker
	utils.WriteProductionHealthResponse(w, h.readinessChecker, h.version)
}

// HandleHealthMetrics handles GET /api/v1/health/metrics (system metrics)
func (h *HealthHandler) HandleHealthMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Get system metrics
	metrics := h.monitoringCollector.GetSystemMetrics()

	response := map[string]interface{}{
		"metrics":   metrics,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"version":   h.version,
	}

	utils.WriteJSON(w, http.StatusOK, response)
}

// performHealthChecks performs comprehensive health checks
func (h *HealthHandler) performHealthChecks() map[string]responses.HealthCheck {
	checks := make(map[string]responses.HealthCheck)

	// System health check
	checks["system"] = h.checkSystemHealth()

	// Storage health check
	checks["storage"] = h.checkStorageHealth()

	// Docker health check
	checks["docker"] = h.checkDockerHealth()

	// Auth service health check
	checks["auth"] = h.checkAuthHealth()

	// UPS health check
	checks["ups"] = h.checkUPSHealth()

	// Virtual Machines health check
	checks["vms"] = h.checkVMHealth()

	return checks
}

// performReadinessChecks performs readiness-specific checks
func (h *HealthHandler) performReadinessChecks() map[string]responses.HealthCheck {
	checks := make(map[string]responses.HealthCheck)

	// Check if system APIs are responsive
	checks["system_api"] = h.checkSystemAPIReadiness()

	// Check if storage APIs are responsive
	checks["storage_api"] = h.checkStorageAPIReadiness()

	// Check if Docker APIs are responsive
	checks["docker_api"] = h.checkDockerAPIReadiness()

	return checks
}

// Individual health check methods

func (h *HealthHandler) checkSystemHealth() responses.HealthCheck {
	start := time.Now()

	// Try to get system info
	_, err := h.api.GetSystem().GetCPUInfo()
	duration := time.Since(start)

	if err != nil {
		return responses.HealthCheck{
			Status:    "fail",
			Message:   "System API not responding",
			Timestamp: time.Now().UTC(),
			Duration:  duration.String(),
		}
	}

	return responses.HealthCheck{
		Status:    "pass",
		Message:   "System API healthy",
		Timestamp: time.Now().UTC(),
		Duration:  duration.String(),
	}
}

func (h *HealthHandler) checkStorageHealth() responses.HealthCheck {
	start := time.Now()

	// Try to get storage info
	_, err := h.api.GetStorage().GetArrayInfo()
	duration := time.Since(start)

	if err != nil {
		return responses.HealthCheck{
			Status:    "fail",
			Message:   "Storage API not responding",
			Timestamp: time.Now().UTC(),
			Duration:  duration.String(),
		}
	}

	return responses.HealthCheck{
		Status:    "pass",
		Message:   "Storage API healthy",
		Timestamp: time.Now().UTC(),
		Duration:  duration.String(),
	}
}

func (h *HealthHandler) checkDockerHealth() responses.HealthCheck {
	start := time.Now()

	// Try to get Docker info
	_, err := h.api.GetDocker().GetSystemInfo()
	duration := time.Since(start)

	if err != nil {
		return responses.HealthCheck{
			Status:    "warn",
			Message:   "Docker API not responding",
			Timestamp: time.Now().UTC(),
			Duration:  duration.String(),
		}
	}

	return responses.HealthCheck{
		Status:    "pass",
		Message:   "Docker API healthy",
		Timestamp: time.Now().UTC(),
		Duration:  duration.String(),
	}
}

func (h *HealthHandler) checkAuthHealth() responses.HealthCheck {
	return responses.HealthCheck{
		Status:    "pass",
		Message:   "Authentication not implemented - UMA operates without authentication",
		Timestamp: time.Now().UTC(),
		Duration:  "0ms",
	}
}

func (h *HealthHandler) checkUPSHealth() responses.HealthCheck {
	start := time.Now()

	// Check if UPS detector is available
	upsDetector := h.api.GetUPSDetector()
	if !upsDetector.IsAvailable() {
		duration := time.Since(start)
		return responses.HealthCheck{
			Status:    "fail",
			Message:   "UPS service unavailable",
			Timestamp: time.Now().UTC(),
			Duration:  duration.String(),
		}
	}

	// Try to get UPS status
	_ = upsDetector.GetStatus()
	duration := time.Since(start)

	return responses.HealthCheck{
		Status:    "pass",
		Message:   "UPS API healthy",
		Timestamp: time.Now().UTC(),
		Duration:  duration.String(),
	}
}

func (h *HealthHandler) checkVMHealth() responses.HealthCheck {
	start := time.Now()

	// Try to get VM list
	_, err := h.api.GetVM().GetVMs()
	duration := time.Since(start)

	if err != nil {
		return responses.HealthCheck{
			Status:    "fail",
			Message:   "VM service unavailable",
			Timestamp: time.Now().UTC(),
			Duration:  duration.String(),
		}
	}

	return responses.HealthCheck{
		Status:    "pass",
		Message:   "Virtual Machines API healthy",
		Timestamp: time.Now().UTC(),
		Duration:  duration.String(),
	}
}

// Readiness check methods

func (h *HealthHandler) checkSystemAPIReadiness() responses.HealthCheck {
	return h.checkSystemHealth() // Same as health check for now
}

func (h *HealthHandler) checkStorageAPIReadiness() responses.HealthCheck {
	return h.checkStorageHealth() // Same as health check for now
}

func (h *HealthHandler) checkDockerAPIReadiness() responses.HealthCheck {
	return h.checkDockerHealth() // Same as health check for now
}

// Helper methods

func (h *HealthHandler) determineOverallStatus(checks map[string]responses.HealthCheck) string {
	hasFailures := false
	hasWarnings := false

	for _, check := range checks {
		switch check.Status {
		case "fail":
			hasFailures = true
		case "warn":
			hasWarnings = true
		}
	}

	if hasFailures {
		return "unhealthy"
	}
	if hasWarnings {
		return "degraded"
	}
	return "healthy"
}

func (h *HealthHandler) isSystemReady(checks map[string]responses.HealthCheck) bool {
	for _, check := range checks {
		if check.Status == "fail" {
			return false
		}
	}
	return true
}

// Removed unused function: formatDuration
