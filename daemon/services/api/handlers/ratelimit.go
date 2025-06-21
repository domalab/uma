package handlers

import (
	"encoding/json"
	"net/http"
	"time"

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

	// Note: Rate limiting is not currently implemented in UMA
	// This endpoint returns a disabled state to indicate no active rate limiting
	stats := map[string]interface{}{
		"enabled":            false,
		"global_limit":       0,
		"global_window":      "disabled",
		"current_requests":   0,
		"remaining_requests": 0,
		"reset_time":         nil,
		"blocked_requests":   0,
		"total_requests":     0,
		"endpoints":          map[string]interface{}{},
		"clients":            map[string]interface{}{},
		"message":            "Rate limiting is not currently implemented",
		"last_updated":       time.Now().UTC().Format(time.RFC3339),
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
	// Note: Rate limiting is not currently implemented in UMA
	// Return disabled configuration to indicate no active rate limiting
	config := map[string]interface{}{
		"enabled":       false,
		"global_limit":  0,
		"global_window": "disabled",
		"endpoints":     map[string]interface{}{},
		"whitelist":     []string{},
		"blacklist":     []string{},
		"headers": map[string]interface{}{
			"include_headers":    false,
			"limit_header":       "",
			"remaining_header":   "",
			"reset_header":       "",
			"retry_after_header": "",
		},
		"message":      "Rate limiting is not currently implemented",
		"last_updated": time.Now().UTC().Format(time.RFC3339),
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

	// Note: Rate limiting is not currently implemented in UMA
	// This endpoint returns an error to indicate the feature is not available
	logger.Yellow("Rate limit configuration update attempted, but rate limiting is not implemented")

	utils.WriteError(w, http.StatusNotImplemented, "Rate limiting is not currently implemented in UMA")
}
