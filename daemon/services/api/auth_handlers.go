package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/domalab/uma/daemon/services/auth"
)

// LoginRequest represents a login request
type LoginRequest struct {
	APIKey string `json:"api_key" validate:"required"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	Token string     `json:"token"`
	User  *auth.User `json:"user"`
}

// CreateUserRequest represents a user creation request
type CreateUserRequest struct {
	Username string    `json:"username" validate:"required"`
	Role     auth.Role `json:"role" validate:"required"`
}

// UpdateUserRequest represents a user update request
type UpdateUserRequest struct {
	Role   *auth.Role `json:"role,omitempty"`
	Active *bool      `json:"active,omitempty"`
}

// handleAuthLogin handles POST /api/v1/auth/login
func (h *HTTPServer) handleAuthLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	if h.authService == nil || !h.authService.IsEnabled() {
		h.writeError(w, http.StatusNotImplemented, "Authentication is disabled")
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	// Validate API key and get user
	user, err := h.authService.ValidateAPIKey(req.APIKey)
	if err != nil {
		h.writeError(w, http.StatusUnauthorized, err.Error())
		return
	}

	// Generate JWT token
	token, err := h.authService.GenerateJWT(user)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	response := LoginResponse{
		Token: token,
		User:  user,
	}

	h.writeStandardResponse(w, http.StatusOK, response, nil)
}

// handleAuthToken handles POST /api/v1/auth/token (alias for login)
func (h *HTTPServer) handleAuthToken(w http.ResponseWriter, r *http.Request) {
	h.handleAuthLogin(w, r)
}

// handleAuthUsers handles GET /api/v1/auth/users
func (h *HTTPServer) handleAuthUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	if h.authService == nil || !h.authService.IsEnabled() {
		h.writeError(w, http.StatusNotImplemented, "Authentication is disabled")
		return
	}

	users, err := h.authService.GetUsers()
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.writeStandardResponse(w, http.StatusOK, users, nil)
}

// handleAuthCreateUser handles POST /api/v1/auth/users
func (h *HTTPServer) handleAuthCreateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	if h.authService == nil || !h.authService.IsEnabled() {
		h.writeError(w, http.StatusNotImplemented, "Authentication is disabled")
		return
	}

	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	// Validate request
	if err := h.validator.Struct(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Validation failed: "+err.Error())
		return
	}

	// Validate role
	if req.Role != auth.RoleAdmin && req.Role != auth.RoleOperator && req.Role != auth.RoleViewer {
		h.writeError(w, http.StatusBadRequest, "Invalid role. Must be admin, operator, or viewer")
		return
	}

	user, err := h.authService.CreateUser(req.Username, req.Role)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.writeStandardResponse(w, http.StatusCreated, user, nil)
}

// handleAuthGetUser handles GET /api/v1/auth/users/{id}
func (h *HTTPServer) handleAuthGetUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	if h.authService == nil || !h.authService.IsEnabled() {
		h.writeError(w, http.StatusNotImplemented, "Authentication is disabled")
		return
	}

	userID := chi.URLParam(r, "id")
	if userID == "" {
		h.writeError(w, http.StatusBadRequest, "User ID is required")
		return
	}

	user, err := h.authService.GetUser(userID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.writeError(w, http.StatusNotFound, err.Error())
		} else {
			h.writeError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	h.writeStandardResponse(w, http.StatusOK, user, nil)
}

// handleAuthUpdateUser handles PUT /api/v1/auth/users/{id}
func (h *HTTPServer) handleAuthUpdateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	if h.authService == nil || !h.authService.IsEnabled() {
		h.writeError(w, http.StatusNotImplemented, "Authentication is disabled")
		return
	}

	userID := chi.URLParam(r, "id")
	if userID == "" {
		h.writeError(w, http.StatusBadRequest, "User ID is required")
		return
	}

	var req UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	// Build updates map
	updates := make(map[string]interface{})
	if req.Role != nil {
		// Validate role
		if *req.Role != auth.RoleAdmin && *req.Role != auth.RoleOperator && *req.Role != auth.RoleViewer {
			h.writeError(w, http.StatusBadRequest, "Invalid role. Must be admin, operator, or viewer")
			return
		}
		updates["role"] = string(*req.Role)
	}
	if req.Active != nil {
		updates["active"] = *req.Active
	}

	if len(updates) == 0 {
		h.writeError(w, http.StatusBadRequest, "No updates provided")
		return
	}

	err := h.authService.UpdateUser(userID, updates)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.writeError(w, http.StatusNotFound, err.Error())
		} else {
			h.writeError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	h.writeStandardResponse(w, http.StatusOK, map[string]string{"message": "User updated successfully"}, nil)
}

// handleAuthDeleteUser handles DELETE /api/v1/auth/users/{id}
func (h *HTTPServer) handleAuthDeleteUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	if h.authService == nil || !h.authService.IsEnabled() {
		h.writeError(w, http.StatusNotImplemented, "Authentication is disabled")
		return
	}

	userID := chi.URLParam(r, "id")
	if userID == "" {
		h.writeError(w, http.StatusBadRequest, "User ID is required")
		return
	}

	err := h.authService.DeleteUser(userID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.writeError(w, http.StatusNotFound, err.Error())
		} else {
			h.writeError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	h.writeStandardResponse(w, http.StatusOK, map[string]string{"message": "User deleted successfully"}, nil)
}

// handleAuthRegenerateKey handles POST /api/v1/auth/users/{id}/regenerate-key
func (h *HTTPServer) handleAuthRegenerateKey(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	if h.authService == nil || !h.authService.IsEnabled() {
		h.writeError(w, http.StatusNotImplemented, "Authentication is disabled")
		return
	}

	userID := chi.URLParam(r, "id")
	if userID == "" {
		h.writeError(w, http.StatusBadRequest, "User ID is required")
		return
	}

	newAPIKey, err := h.authService.RegenerateAPIKey(userID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.writeError(w, http.StatusNotFound, err.Error())
		} else {
			h.writeError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	response := map[string]string{
		"message": "API key regenerated successfully",
		"api_key": newAPIKey,
	}

	h.writeStandardResponse(w, http.StatusOK, response, nil)
}

// handleAuthStats handles GET /api/v1/auth/stats
func (h *HTTPServer) handleAuthStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	if h.authService == nil || !h.authService.IsEnabled() {
		h.writeError(w, http.StatusNotImplemented, "Authentication is disabled")
		return
	}

	stats := h.authService.GetAuthStats()
	h.writeStandardResponse(w, http.StatusOK, stats, nil)
}
