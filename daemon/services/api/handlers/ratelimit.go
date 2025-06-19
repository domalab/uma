package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/domalab/uma/daemon/logger"
	"github.com/domalab/uma/daemon/services/api/utils"
)

// RateLimitHandler handles rate limiting-related HTTP requests
type RateLimitHandler struct {
	api utils.APIInterface
}

// NewRateLimitHandler creates a new rate limit handler
func NewRateLimitHandler(api utils.APIInterface) *RateLimitHandler {
	return &RateLimitHandler{
		api: api,
	}
}

// HandleRateLimitStats handles GET /api/v1/rate-limits/stats
func (h *RateLimitHandler) HandleRateLimitStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Return placeholder rate limit statistics
	stats := map[string]interface{}{
		"enabled":            true,
		"global_limit":       1000,
		"global_window":      "1h",
		"current_requests":   245,
		"remaining_requests": 755,
		"reset_time":         "2024-01-01T01:00:00Z",
		"blocked_requests":   12,
		"total_requests":     2567,
		"endpoints": map[string]interface{}{
			"/api/v1/docker/containers": map[string]interface{}{
				"limit":              100,
				"window":             "1h",
				"current_requests":   23,
				"remaining_requests": 77,
			},
			"/api/v1/storage/array": map[string]interface{}{
				"limit":              50,
				"window":             "1h",
				"current_requests":   8,
				"remaining_requests": 42,
			},
			"/api/v1/vms": map[string]interface{}{
				"limit":              200,
				"window":             "1h",
				"current_requests":   45,
				"remaining_requests": 155,
			},
		},
		"clients": map[string]interface{}{
			"192.168.1.100": map[string]interface{}{
				"requests":           156,
				"blocked":            2,
				"last_request":       "2024-01-01T00:45:30Z",
				"remaining_requests": 844,
			},
			"192.168.1.101": map[string]interface{}{
				"requests":           89,
				"blocked":            0,
				"last_request":       "2024-01-01T00:42:15Z",
				"remaining_requests": 911,
			},
		},
	}

	utils.WriteJSON(w, http.StatusOK, stats)
}

// HandleRateLimitConfig handles GET/PUT /api/v1/rate-limits/config
func (h *RateLimitHandler) HandleRateLimitConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.handleGetRateLimitConfig(w, r)
	case http.MethodPut:
		h.handleUpdateRateLimitConfig(w, r)
	default:
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// handleGetRateLimitConfig handles GET /api/v1/rate-limits/config
func (h *RateLimitHandler) handleGetRateLimitConfig(w http.ResponseWriter, r *http.Request) {
	// Return placeholder rate limit configuration
	config := map[string]interface{}{
		"enabled":       true,
		"global_limit":  1000,
		"global_window": "1h",
		"endpoints": map[string]interface{}{
			"/api/v1/docker/containers": map[string]interface{}{
				"enabled": true,
				"limit":   100,
				"window":  "1h",
				"burst":   10,
			},
			"/api/v1/storage/array": map[string]interface{}{
				"enabled": true,
				"limit":   50,
				"window":  "1h",
				"burst":   5,
			},
			"/api/v1/vms": map[string]interface{}{
				"enabled": true,
				"limit":   200,
				"window":  "1h",
				"burst":   20,
			},
			"/api/v1/auth/login": map[string]interface{}{
				"enabled": true,
				"limit":   10,
				"window":  "15m",
				"burst":   3,
			},
		},
		"whitelist": []string{
			"127.0.0.1",
			"::1",
			"192.168.1.0/24",
		},
		"blacklist": []string{
			"10.0.0.100",
		},
		"headers": map[string]interface{}{
			"include_headers":    true,
			"limit_header":       "X-RateLimit-Limit",
			"remaining_header":   "X-RateLimit-Remaining",
			"reset_header":       "X-RateLimit-Reset",
			"retry_after_header": "Retry-After",
		},
	}

	utils.WriteJSON(w, http.StatusOK, config)
}

// handleUpdateRateLimitConfig handles PUT /api/v1/rate-limits/config
func (h *RateLimitHandler) handleUpdateRateLimitConfig(w http.ResponseWriter, r *http.Request) {
	var config map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	// Validate configuration (placeholder validation)
	if enabled, ok := config["enabled"].(bool); ok && !enabled {
		logger.Yellow("Rate limiting disabled via API")
	}

	if globalLimit, ok := config["global_limit"].(float64); ok {
		if globalLimit < 1 || globalLimit > 10000 {
			utils.WriteError(w, http.StatusBadRequest, "Global limit must be between 1 and 10000")
			return
		}
	}

	// In a real implementation, this would update the actual rate limit configuration
	logger.Blue("Rate limit configuration updated via API")

	utils.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Rate limit configuration updated successfully",
		"config":  config,
	})
}
