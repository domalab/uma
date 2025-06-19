package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/domalab/uma/daemon/services/api/types/requests"
	"github.com/domalab/uma/daemon/services/api/types/responses"
	"github.com/domalab/uma/daemon/services/api/utils"
)

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	api utils.APIInterface
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(api utils.APIInterface) *AuthHandler {
	return &AuthHandler{api: api}
}

// HandleAuthLogin handles POST /api/v1/auth/login
func (h *AuthHandler) HandleAuthLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var request requests.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid JSON request")
		return
	}

	// Validate the login request
	if err := utils.ValidateStruct(&request); err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Sprintf("Validation error: %v", err))
		return
	}

	// Attempt login
	loginResponse, err := h.api.GetAuth().Login(request.Username, request.Password)
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	utils.WriteJSON(w, http.StatusOK, loginResponse)
}

// HandleAuthUsers handles GET/POST /api/v1/auth/users
func (h *AuthHandler) HandleAuthUsers(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.handleGetUsers(w, r)
	case http.MethodPost:
		h.handleCreateUser(w, r)
	default:
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// HandleAuthStats handles GET /api/v1/auth/stats
func (h *AuthHandler) HandleAuthStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	stats, err := h.api.GetAuth().GetStats()
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get auth stats: %v", err))
		return
	}

	utils.WriteJSON(w, http.StatusOK, stats)
}

// HandleAuthUser handles individual user operations
func (h *AuthHandler) HandleAuthUser(w http.ResponseWriter, r *http.Request, userID string) {
	switch r.Method {
	case http.MethodGet:
		h.handleGetUser(w, r, userID)
	case http.MethodPut:
		h.handleUpdateUser(w, r, userID)
	case http.MethodDelete:
		h.handleDeleteUser(w, r, userID)
	default:
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// HandleAuthRefresh handles POST /api/v1/auth/refresh
func (h *AuthHandler) HandleAuthRefresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var request requests.TokenRefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid JSON request")
		return
	}

	// Validate the refresh request
	if err := utils.ValidateStruct(&request); err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Sprintf("Validation error: %v", err))
		return
	}

	// Implementation would refresh token
	// For now, return placeholder response
	response := responses.TokenResponse{
		AccessToken: "new_access_token",
		TokenType:   "Bearer",
		ExpiresIn:   3600,
		IssuedAt:    time.Now().UTC(),
	}

	utils.WriteJSON(w, http.StatusOK, response)
}

// HandleAuthLogout handles POST /api/v1/auth/logout
func (h *AuthHandler) HandleAuthLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Implementation would invalidate token
	// For now, return success
	response := map[string]interface{}{
		"message":   "Logged out successfully",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	utils.WriteJSON(w, http.StatusOK, response)
}

// Helper methods

// handleGetUsers handles GET requests for user list
func (h *AuthHandler) handleGetUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.api.GetAuth().GetUsers()
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get users: %v", err))
		return
	}

	utils.WriteJSON(w, http.StatusOK, users)
}

// handleCreateUser handles POST requests to create a new user
func (h *AuthHandler) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	var request requests.UserCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid JSON request")
		return
	}

	// Validate the create request
	if err := utils.ValidateStruct(&request); err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Sprintf("Validation error: %v", err))
		return
	}

	// Implementation would create user
	// For now, return success
	user := responses.UserInfo{
		ID:       "new_user_id",
		Username: request.Username,
		Email:    request.Email,
		FullName: request.FullName,
		Roles:    request.Roles,
		Enabled:  request.Enabled,
		Created:  time.Now().UTC(),
		Updated:  time.Now().UTC(),
	}

	utils.WriteJSON(w, http.StatusCreated, user)
}

// handleGetUser handles GET requests for individual user
func (h *AuthHandler) handleGetUser(w http.ResponseWriter, r *http.Request, userID string) {
	// Implementation would get user by ID
	// For now, return placeholder
	user := responses.UserInfo{
		ID:       userID,
		Username: "placeholder_user",
		Email:    "user@example.com",
		Enabled:  true,
		Created:  time.Now().UTC(),
		Updated:  time.Now().UTC(),
	}

	utils.WriteJSON(w, http.StatusOK, user)
}

// handleUpdateUser handles PUT requests to update user
func (h *AuthHandler) handleUpdateUser(w http.ResponseWriter, r *http.Request, userID string) {
	var request requests.UserUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid JSON request")
		return
	}

	// Validate the update request
	if err := utils.ValidateStruct(&request); err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Sprintf("Validation error: %v", err))
		return
	}

	// Implementation would update user
	// For now, return success
	user := responses.UserInfo{
		ID:       userID,
		Username: "updated_user",
		Email:    request.Email,
		FullName: request.FullName,
		Roles:    request.Roles,
		Enabled:  request.Enabled,
		Updated:  time.Now().UTC(),
	}

	utils.WriteJSON(w, http.StatusOK, user)
}

// handleDeleteUser handles DELETE requests to remove user
func (h *AuthHandler) handleDeleteUser(w http.ResponseWriter, r *http.Request, userID string) {
	// Implementation would delete user
	// For now, return success
	response := map[string]interface{}{
		"message":   "User deleted successfully",
		"user_id":   userID,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	utils.WriteJSON(w, http.StatusOK, response)
}

// HandleAuthChangePassword handles POST /api/v1/auth/change-password
func (h *AuthHandler) HandleAuthChangePassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var request requests.PasswordChangeRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid JSON request")
		return
	}

	// Validate the password change request
	if err := utils.ValidateStruct(&request); err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Sprintf("Validation error: %v", err))
		return
	}

	// Implementation would change password
	// For now, return success
	response := map[string]interface{}{
		"message":   "Password changed successfully",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	utils.WriteJSON(w, http.StatusOK, response)
}
