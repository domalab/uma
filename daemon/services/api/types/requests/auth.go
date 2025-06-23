package requests

// Authentication-related request types

// LoginRequest represents a user login request
type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// PasswordChangeRequest represents a password change request
type PasswordChangeRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=8"`
}

// UserCreateRequest represents a user creation request
type UserCreateRequest struct {
	Username    string   `json:"username" validate:"required"`
	Password    string   `json:"password" validate:"required,min=8"`
	Email       string   `json:"email" validate:"required,email"`
	FullName    string   `json:"full_name,omitempty"`
	Roles       []string `json:"roles,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
	Enabled     bool     `json:"enabled"`
}

// UserUpdateRequest represents a user update request
type UserUpdateRequest struct {
	Email       *string  `json:"email,omitempty" validate:"omitempty,email"`
	FullName    *string  `json:"full_name,omitempty"`
	Roles       []string `json:"roles,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
	Enabled     *bool    `json:"enabled,omitempty"`
}

// TokenRefreshRequest represents a token refresh request
type TokenRefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// LogoutRequest represents a logout request
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token,omitempty"`
}
