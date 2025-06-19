package responses

import "time"

// Authentication-related response types

// LoginResponse represents a successful login response
type LoginResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"` // "Bearer"
	ExpiresIn    int       `json:"expires_in"` // Seconds until expiration
	User         UserInfo  `json:"user"`
	IssuedAt     time.Time `json:"issued_at"`
}

// UserInfo represents user information
type UserInfo struct {
	ID          string     `json:"id"`
	Username    string     `json:"username"`
	Email       string     `json:"email,omitempty"`
	FullName    string     `json:"full_name,omitempty"`
	Roles       []string   `json:"roles,omitempty"`
	Permissions []string   `json:"permissions,omitempty"`
	Enabled     bool       `json:"enabled"`
	LastLogin   *time.Time `json:"last_login,omitempty"`
	Created     time.Time  `json:"created"`
	Updated     time.Time  `json:"updated"`
}

// TokenResponse represents a token refresh response
type TokenResponse struct {
	AccessToken string    `json:"access_token"`
	TokenType   string    `json:"token_type"` // "Bearer"
	ExpiresIn   int       `json:"expires_in"` // Seconds until expiration
	IssuedAt    time.Time `json:"issued_at"`
}

// UserListResponse represents a list of users
type UserListResponse struct {
	Users       []UserInfo `json:"users"`
	Total       int        `json:"total"`
	Active      int        `json:"active"`
	Inactive    int        `json:"inactive"`
	LastUpdated time.Time  `json:"last_updated"`
}

// RoleInfo represents role information
type RoleInfo struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Permissions []string  `json:"permissions"`
	UserCount   int       `json:"user_count"`
	Created     time.Time `json:"created"`
	Updated     time.Time `json:"updated"`
}

// RoleListResponse represents a list of roles
type RoleListResponse struct {
	Roles       []RoleInfo `json:"roles"`
	Total       int        `json:"total"`
	LastUpdated time.Time  `json:"last_updated"`
}

// PermissionInfo represents permission information
type PermissionInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Resource    string `json:"resource"`
	Action      string `json:"action"`
}

// PermissionListResponse represents a list of permissions
type PermissionListResponse struct {
	Permissions []PermissionInfo `json:"permissions"`
	Total       int              `json:"total"`
	LastUpdated time.Time        `json:"last_updated"`
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
	LastFailedLogin     *time.Time `json:"last_failed_login,omitempty"`
	LastUpdated         time.Time  `json:"last_updated"`
}

// SessionInfo represents session information
type SessionInfo struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	Username   string    `json:"username"`
	IPAddress  string    `json:"ip_address"`
	UserAgent  string    `json:"user_agent"`
	Created    time.Time `json:"created"`
	LastAccess time.Time `json:"last_access"`
	ExpiresAt  time.Time `json:"expires_at"`
	Active     bool      `json:"active"`
}

// SessionListResponse represents a list of sessions
type SessionListResponse struct {
	Sessions    []SessionInfo `json:"sessions"`
	Total       int           `json:"total"`
	Active      int           `json:"active"`
	Expired     int           `json:"expired"`
	LastUpdated time.Time     `json:"last_updated"`
}
