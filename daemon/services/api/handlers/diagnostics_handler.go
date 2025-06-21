package handlers

import (
	"net/http"
	"time"

	"github.com/domalab/uma/daemon/services/api/utils"
)

// DiagnosticsHandler handles diagnostics-related HTTP requests
type DiagnosticsHandler struct {
	apiAdapter utils.APIInterface
}

// NewDiagnosticsHandler creates a new diagnostics handler instance
func NewDiagnosticsHandler(apiAdapter utils.APIInterface) *DiagnosticsHandler {
	return &DiagnosticsHandler{
		apiAdapter: apiAdapter,
	}
}

// HandleDiagnosticsHealth handles GET /api/v1/diagnostics/health
func (h *DiagnosticsHandler) HandleDiagnosticsHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Get diagnostics manager from API
	if apiInstance, ok := h.apiAdapter.(interface{ GetDiagnosticsManager() interface{} }); ok {
		if diagnosticsManager := apiInstance.GetDiagnosticsManager(); diagnosticsManager != nil {
			// Use reflection to call GetHealthChecks method
			if dm, ok := diagnosticsManager.(interface{ GetHealthChecks() interface{} }); ok {
				health := dm.GetHealthChecks()
				// Transform health data to ensure schema compliance
				transformedHealth := h.transformHealthData(health)
				utils.WriteJSON(w, http.StatusOK, transformedHealth)
				return
			}
		}
	}

	// Fallback: diagnostics service not available
	health := map[string]interface{}{
		"status":     "unknown", // Use "unknown" instead of "unavailable" to match enum
		"checks":     []interface{}{},
		"message":    "Diagnostics service not available",
		"last_check": time.Now().UTC().Format(time.RFC3339), // Use current time instead of nil
	}

	utils.WriteJSON(w, http.StatusOK, health)
}

// HandleDiagnosticsInfo handles GET /api/v1/diagnostics/info
func (h *DiagnosticsHandler) HandleDiagnosticsInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Get diagnostics manager from API
	if apiInstance, ok := h.apiAdapter.(interface{ GetDiagnosticsManager() interface{} }); ok {
		if diagnosticsManager := apiInstance.GetDiagnosticsManager(); diagnosticsManager != nil {
			// Use reflection to call GetDiagnosticsInfo method
			if dm, ok := diagnosticsManager.(interface{ GetDiagnosticsInfo() interface{} }); ok {
				info := dm.GetDiagnosticsInfo()
				utils.WriteJSON(w, http.StatusOK, info)
				return
			}
		}
	}

	// Fallback: diagnostics service not available
	info := map[string]interface{}{
		"version":     "1.0.0",
		"system":      "uma",
		"diagnostics": "unavailable",
		"message":     "Diagnostics service not available",
	}

	utils.WriteJSON(w, http.StatusOK, info)
}

// HandleDiagnosticsRepair handles GET/POST /api/v1/diagnostics/repair
func (h *DiagnosticsHandler) HandleDiagnosticsRepair(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// Get available repairs from diagnostics manager
		if apiInstance, ok := h.apiAdapter.(interface{ GetDiagnosticsManager() interface{} }); ok {
			if diagnosticsManager := apiInstance.GetDiagnosticsManager(); diagnosticsManager != nil {
				// Use reflection to call GetAvailableRepairs method
				if dm, ok := diagnosticsManager.(interface{ GetAvailableRepairs() interface{} }); ok {
					repairs := dm.GetAvailableRepairs()
					utils.WriteJSON(w, http.StatusOK, repairs)
					return
				}
			}
		}

		// Fallback: repair service not available
		repairs := map[string]interface{}{
			"available_repairs": []interface{}{},
			"message":           "Repair service not available",
		}
		utils.WriteJSON(w, http.StatusOK, repairs)

	case http.MethodPost:
		repairName := r.URL.Query().Get("action")
		if repairName == "" {
			utils.WriteError(w, http.StatusBadRequest, "Repair action required")
			return
		}

		// Execute repair through diagnostics manager
		if apiInstance, ok := h.apiAdapter.(interface{ GetDiagnosticsManager() interface{} }); ok {
			if diagnosticsManager := apiInstance.GetDiagnosticsManager(); diagnosticsManager != nil {
				// Use reflection to call ExecuteRepair method
				if dm, ok := diagnosticsManager.(interface{ ExecuteRepair(string) interface{} }); ok {
					result := dm.ExecuteRepair(repairName)
					utils.WriteJSON(w, http.StatusOK, result)
					return
				}
			}
		}

		// Fallback: repair service not available
		utils.WriteError(w, http.StatusNotImplemented, "Repair service not available")

	default:
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// transformHealthData transforms health data to ensure schema compliance
func (h *DiagnosticsHandler) transformHealthData(health interface{}) map[string]interface{} {
	transformed := map[string]interface{}{
		"status":     "unknown",
		"checks":     []interface{}{},
		"last_check": time.Now().UTC().Format(time.RFC3339),
	}

	// Handle different possible return types
	if healthMap, ok := health.(map[string]interface{}); ok {
		// Copy existing fields
		for key, value := range healthMap {
			transformed[key] = value
		}

		// Fix status enum violations
		if status, exists := healthMap["status"]; exists {
			if statusStr, ok := status.(string); ok {
				transformed["status"] = h.normalizeHealthStatus(statusStr)
			}
		}

		// Fix last_check null values
		if lastCheck, exists := healthMap["last_check"]; exists {
			if lastCheck == nil {
				transformed["last_check"] = time.Now().UTC().Format(time.RFC3339)
			} else if lastCheckStr, ok := lastCheck.(string); ok && lastCheckStr != "" {
				transformed["last_check"] = lastCheckStr
			} else {
				transformed["last_check"] = time.Now().UTC().Format(time.RFC3339)
			}
		}

		// Ensure checks is an array
		if checks, exists := healthMap["checks"]; exists {
			if checksArray, ok := checks.([]interface{}); ok {
				transformedChecks := make([]interface{}, 0, len(checksArray))
				for _, check := range checksArray {
					if checkMap, ok := check.(map[string]interface{}); ok {
						transformedCheck := h.transformHealthCheck(checkMap)
						transformedChecks = append(transformedChecks, transformedCheck)
					}
				}
				transformed["checks"] = transformedChecks
			}
		}
	}

	return transformed
}

// normalizeHealthStatus converts various status values to schema-compliant enum values
func (h *DiagnosticsHandler) normalizeHealthStatus(status string) string {
	switch status {
	case "healthy", "ok", "pass", "passed", "good":
		return "healthy"
	case "warning", "warn", "caution":
		return "warning"
	case "critical", "error", "fail", "failed", "bad":
		return "critical"
	case "unavailable", "unknown", "pending", "":
		return "unknown"
	default:
		return "unknown"
	}
}

// transformHealthCheck transforms individual health check to match schema
func (h *DiagnosticsHandler) transformHealthCheck(check map[string]interface{}) map[string]interface{} {
	transformed := make(map[string]interface{})

	// Copy all existing fields
	for key, value := range check {
		transformed[key] = value
	}

	// Ensure required fields exist
	if _, exists := transformed["name"]; !exists {
		transformed["name"] = "unknown_check"
	}
	if _, exists := transformed["message"]; !exists {
		transformed["message"] = "No message available"
	}

	// Fix status enum violations for individual checks
	if status, exists := transformed["status"]; exists {
		if statusStr, ok := status.(string); ok {
			// Individual checks use different enum values
			transformed["status"] = h.normalizeCheckStatus(statusStr)
		}
	} else {
		transformed["status"] = "unknown"
	}

	// Add last_updated if missing
	if _, exists := transformed["last_updated"]; !exists {
		transformed["last_updated"] = time.Now().UTC().Format(time.RFC3339)
	}

	return transformed
}

// normalizeCheckStatus converts check status to schema-compliant enum values
func (h *DiagnosticsHandler) normalizeCheckStatus(status string) string {
	switch status {
	case "passed", "pass", "ok", "healthy", "good":
		return "passed"
	case "warning", "warn", "caution":
		return "warning"
	case "critical", "error", "fail", "failed", "bad":
		return "critical"
	case "unknown", "pending", "unavailable", "":
		return "unknown"
	default:
		return "unknown"
	}
}
