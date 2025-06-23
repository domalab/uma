package responses

import "time"

// Authentication-related response types

// LoginResponse represents a successful login response
type LoginResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"`
	ExpiresIn    int       `json:"expires_in"`
	User         UserInfo  `json:"user"`
	IssuedAt     time.Time `json:"issued_at"`
}

// UserInfo represents user information
type UserInfo struct {
	ID       string    `json:"id"`
	Username string    `json:"username"`
	Email    string    `json:"email"`
	FullName string    `json:"full_name,omitempty"`
	Roles    []string  `json:"roles"`
	Enabled  bool      `json:"enabled"`
	Created  time.Time `json:"created"`
	Updated  time.Time `json:"updated"`
}

// UserListResponse represents a list of users response
type UserListResponse struct {
	Users       []UserInfo `json:"users"`
	Total       int        `json:"total"`
	Active      int        `json:"active"`
	Inactive    int        `json:"inactive"`
	LastUpdated time.Time  `json:"last_updated"`
}

// AuthStatsResponse represents authentication statistics
type AuthStatsResponse struct {
	TotalUsers          int        `json:"total_users"`
	ActiveUsers         int        `json:"active_users"`
	InactiveUsers       int        `json:"inactive_users"`
	TotalSessions       int        `json:"total_sessions"`
	ActiveSessions      int        `json:"active_sessions"`
	FailedLogins24h     int        `json:"failed_logins_24h"`
	SuccessfulLogins24h int        `json:"successful_logins_24h"`
	LastLogin           *time.Time `json:"last_login,omitempty"`
	LastUpdated         time.Time  `json:"last_updated"`
}

// TokenRefreshResponse represents a token refresh response
type TokenRefreshResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"`
	ExpiresIn    int       `json:"expires_in"`
	IssuedAt     time.Time `json:"issued_at"`
}

// LogoutResponse represents a logout response
type LogoutResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// UserCreateResponse represents a user creation response
type UserCreateResponse struct {
	User    UserInfo `json:"user"`
	Success bool     `json:"success"`
	Message string   `json:"message"`
}

// UserUpdateResponse represents a user update response
type UserUpdateResponse struct {
	User    UserInfo `json:"user"`
	Success bool     `json:"success"`
	Message string   `json:"message"`
}
