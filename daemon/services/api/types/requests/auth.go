package requests

// Authentication-related request types

// LoginRequest represents a user login request
type LoginRequest struct {
	Username string `json:"username,omitempty" validate:"required_without=APIKey"`
	Password string `json:"password,omitempty" validate:"required_without=APIKey"`
	APIKey   string `json:"api_key,omitempty" validate:"required_without_all=Username Password"`
}

// UserCreateRequest represents a request to create a new user
type UserCreateRequest struct {
	Username    string   `json:"username" validate:"required,min=3,max=32"`
	Password    string   `json:"password" validate:"required,min=8"`
	Email       string   `json:"email,omitempty" validate:"omitempty,email"`
	FullName    string   `json:"full_name,omitempty"`
	Roles       []string `json:"roles,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
	Enabled     bool     `json:"enabled"`
}

// UserUpdateRequest represents a request to update user information
type UserUpdateRequest struct {
	Password    string   `json:"password,omitempty" validate:"omitempty,min=8"`
	Email       string   `json:"email,omitempty" validate:"omitempty,email"`
	FullName    string   `json:"full_name,omitempty"`
	Roles       []string `json:"roles,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
	Enabled     bool     `json:"enabled,omitempty"`
}

// PasswordChangeRequest represents a request to change user password
type PasswordChangeRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=8"`
}

// TokenRefreshRequest represents a request to refresh an authentication token
type TokenRefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// RoleCreateRequest represents a request to create a new role
type RoleCreateRequest struct {
	Name        string   `json:"name" validate:"required,min=3,max=32"`
	Description string   `json:"description,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
}

// RoleUpdateRequest represents a request to update role information
type RoleUpdateRequest struct {
	Description string   `json:"description,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
}
