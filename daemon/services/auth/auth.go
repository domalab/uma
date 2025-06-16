package auth

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/domalab/uma/daemon/domain"
	"github.com/domalab/uma/daemon/logger"
	"github.com/golang-jwt/jwt/v5"
)

// Role represents user roles for access control
type Role string

const (
	RoleAdmin    Role = "admin"
	RoleOperator Role = "operator"
	RoleViewer   Role = "viewer"
)

// User represents a user in the system
type User struct {
	ID       string    `json:"id"`
	Username string    `json:"username"`
	Role     Role      `json:"role"`
	APIKey   string    `json:"api_key,omitempty"`
	Created  time.Time `json:"created"`
	LastUsed time.Time `json:"last_used,omitempty"`
	Active   bool      `json:"active"`
}

// Claims represents JWT claims
type Claims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Role     Role   `json:"role"`
	jwt.RegisteredClaims
}

// AuthContext key for storing user in request context
type AuthContext string

const UserContextKey AuthContext = "user"

// AuthService handles authentication and authorization
type AuthService struct {
	config    domain.AuthConfig
	jwtSecret []byte
	users     map[string]*User
	apiKeys   map[string]*User
}

// NewAuthService creates a new authentication service
func NewAuthService(config domain.AuthConfig) *AuthService {
	service := &AuthService{
		config:  config,
		users:   make(map[string]*User),
		apiKeys: make(map[string]*User),
	}

	// Set JWT secret
	if config.JWTSecret != "" {
		service.jwtSecret = []byte(config.JWTSecret)
	} else {
		// Generate a random secret if none provided
		service.jwtSecret = make([]byte, 32)
		rand.Read(service.jwtSecret)
		logger.Info("Generated random JWT secret for authentication")
	}

	// Create default admin user if authentication is enabled and no legacy API key
	if config.Enabled && config.APIKey == "" {
		service.createDefaultAdmin()
	}

	return service
}

// IsEnabled returns whether authentication is enabled
func (a *AuthService) IsEnabled() bool {
	return a.config.Enabled
}

// createDefaultAdmin creates a default admin user
func (a *AuthService) createDefaultAdmin() {
	adminUser := &User{
		ID:       "admin-001",
		Username: "admin",
		Role:     RoleAdmin,
		Created:  time.Now(),
		Active:   true,
	}

	// Generate API key for admin
	apiKey, err := a.GenerateAPIKey()
	if err != nil {
		logger.Error("Failed to generate API key for admin user: %v", err)
		return
	}

	adminUser.APIKey = "uma_" + apiKey
	a.users[adminUser.ID] = adminUser
	a.apiKeys[adminUser.APIKey] = adminUser

	logger.Info("Created default admin user with API key: %s", adminUser.APIKey)
}

// GenerateAPIKey generates a new API key
func (a *AuthService) GenerateAPIKey() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate API key: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// ValidateAPIKey validates an API key and returns the associated user
func (a *AuthService) ValidateAPIKey(providedKey string) (*User, error) {
	if !a.config.Enabled {
		return nil, nil // Authentication disabled, allow access
	}

	if providedKey == "" {
		return nil, errors.New("API key is required")
	}

	// Check new user-based API keys first
	if user, exists := a.apiKeys[providedKey]; exists {
		if !user.Active {
			return nil, errors.New("user account is disabled")
		}
		// Update last used timestamp
		user.LastUsed = time.Now()
		return user, nil
	}

	// Fallback to legacy API key for backward compatibility
	if a.config.APIKey != "" {
		if subtle.ConstantTimeCompare([]byte(providedKey), []byte(a.config.APIKey)) == 1 {
			// Create a virtual admin user for legacy API key
			return &User{
				ID:       "legacy-admin",
				Username: "legacy-admin",
				Role:     RoleAdmin,
				Active:   true,
			}, nil
		}
	}

	return nil, errors.New("invalid API key")
}

// GenerateJWT generates a JWT token for a user
func (a *AuthService) GenerateJWT(user *User) (string, error) {
	if !a.config.Enabled {
		return "", errors.New("authentication is disabled")
	}

	claims := &Claims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "uma",
			Subject:   user.ID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(a.jwtSecret)
}

// ValidateJWT validates a JWT token and returns the claims
func (a *AuthService) ValidateJWT(tokenString string) (*Claims, error) {
	if !a.config.Enabled {
		return nil, nil // Authentication disabled, allow access
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return a.jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// HasPermission checks if a user has permission for a specific action
func (a *AuthService) HasPermission(user *User, action string) bool {
	if !a.config.Enabled || user == nil {
		return true // Authentication disabled or no user, allow access
	}

	switch user.Role {
	case RoleAdmin:
		return true // Admin has all permissions
	case RoleOperator:
		// Operators can read and perform operations but not manage users
		return !strings.HasPrefix(action, "user.")
	case RoleViewer:
		// Viewers can only read
		return strings.HasPrefix(action, "read.")
	default:
		return false
	}
}

// AuthMiddleware returns an HTTP middleware for authentication
func (a *AuthService) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !a.config.Enabled {
			next.ServeHTTP(w, r)
			return
		}

		var user *User
		var err error

		// Check for JWT token first
		auth := r.Header.Get("Authorization")
		if strings.HasPrefix(auth, "Bearer ") {
			token := strings.TrimPrefix(auth, "Bearer ")

			// Try JWT validation first
			claims, jwtErr := a.ValidateJWT(token)
			if jwtErr == nil && claims != nil {
				// JWT is valid, get user from claims
				user = &User{
					ID:       claims.UserID,
					Username: claims.Username,
					Role:     claims.Role,
					Active:   true,
				}
			} else {
				// Try as API key
				user, err = a.ValidateAPIKey(token)
			}
		} else {
			// Check for API key in X-API-Key header
			apiKey := r.Header.Get("X-API-Key")
			if apiKey != "" {
				user, err = a.ValidateAPIKey(apiKey)
			}
		}

		if user == nil {
			if err != nil {
				a.writeUnauthorized(w, err.Error())
			} else {
				a.writeUnauthorized(w, "Missing or invalid authentication")
			}
			return
		}

		// Add user to request context
		ctx := context.WithValue(r.Context(), UserContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// writeUnauthorized writes an unauthorized response
func (a *AuthService) writeUnauthorized(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)

	response := map[string]interface{}{
		"error":     "Unauthorized",
		"message":   message,
		"timestamp": time.Now().UTC(),
	}

	// Simple JSON encoding without external dependencies
	w.Write([]byte(fmt.Sprintf(`{"error":"%s","message":"%s","timestamp":"%s"}`,
		response["error"], response["message"], response["timestamp"])))
}

// RateLimiter provides basic rate limiting functionality
type RateLimiter struct {
	requests map[string][]time.Time
	limit    int
	window   time.Duration
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

// Allow checks if a request from the given IP is allowed
func (rl *RateLimiter) Allow(ip string) bool {
	now := time.Now()

	// Clean old requests
	if requests, exists := rl.requests[ip]; exists {
		var validRequests []time.Time
		for _, reqTime := range requests {
			if now.Sub(reqTime) < rl.window {
				validRequests = append(validRequests, reqTime)
			}
		}
		rl.requests[ip] = validRequests
	}

	// Check if limit exceeded
	if len(rl.requests[ip]) >= rl.limit {
		return false
	}

	// Add current request
	rl.requests[ip] = append(rl.requests[ip], now)
	return true
}

// RateLimitMiddleware returns an HTTP middleware for rate limiting
func (rl *RateLimiter) RateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get client IP
		ip := r.Header.Get("X-Forwarded-For")
		if ip == "" {
			ip = r.Header.Get("X-Real-IP")
		}
		if ip == "" {
			ip = strings.Split(r.RemoteAddr, ":")[0]
		}

		if !rl.Allow(ip) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error":"Rate limit exceeded","message":"Too many requests"}`))
			return
		}

		next.ServeHTTP(w, r)
	})
}

// CreateUser creates a new user
func (a *AuthService) CreateUser(username string, role Role) (*User, error) {
	if !a.config.Enabled {
		return nil, errors.New("authentication is disabled")
	}

	// Check if username already exists
	for _, user := range a.users {
		if user.Username == username {
			return nil, errors.New("username already exists")
		}
	}

	// Generate user ID
	userID := fmt.Sprintf("user-%d", time.Now().Unix())

	// Generate API key
	apiKey, err := a.GenerateAPIKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate API key: %v", err)
	}

	user := &User{
		ID:       userID,
		Username: username,
		Role:     role,
		APIKey:   "uma_" + apiKey,
		Created:  time.Now(),
		Active:   true,
	}

	a.users[userID] = user
	a.apiKeys[user.APIKey] = user

	logger.Info("Created user %s with role %s", username, role)
	return user, nil
}

// GetUsers returns all users (admin only)
func (a *AuthService) GetUsers() ([]*User, error) {
	if !a.config.Enabled {
		return nil, errors.New("authentication is disabled")
	}

	users := make([]*User, 0, len(a.users))
	for _, user := range a.users {
		// Don't include API key in response
		userCopy := *user
		userCopy.APIKey = ""
		users = append(users, &userCopy)
	}

	return users, nil
}

// GetUser returns a specific user by ID
func (a *AuthService) GetUser(userID string) (*User, error) {
	if !a.config.Enabled {
		return nil, errors.New("authentication is disabled")
	}

	user, exists := a.users[userID]
	if !exists {
		return nil, errors.New("user not found")
	}

	// Don't include API key in response
	userCopy := *user
	userCopy.APIKey = ""
	return &userCopy, nil
}

// UpdateUser updates user properties
func (a *AuthService) UpdateUser(userID string, updates map[string]interface{}) error {
	if !a.config.Enabled {
		return errors.New("authentication is disabled")
	}

	user, exists := a.users[userID]
	if !exists {
		return errors.New("user not found")
	}

	// Update allowed fields
	if role, ok := updates["role"].(string); ok {
		user.Role = Role(role)
	}
	if active, ok := updates["active"].(bool); ok {
		user.Active = active
	}

	logger.Info("Updated user %s", user.Username)
	return nil
}

// DeleteUser deletes a user
func (a *AuthService) DeleteUser(userID string) error {
	if !a.config.Enabled {
		return errors.New("authentication is disabled")
	}

	user, exists := a.users[userID]
	if !exists {
		return errors.New("user not found")
	}

	// Remove from both maps
	delete(a.users, userID)
	delete(a.apiKeys, user.APIKey)

	logger.Info("Deleted user %s", user.Username)
	return nil
}

// RegenerateAPIKey generates a new API key for a user
func (a *AuthService) RegenerateAPIKey(userID string) (string, error) {
	if !a.config.Enabled {
		return "", errors.New("authentication is disabled")
	}

	user, exists := a.users[userID]
	if !exists {
		return "", errors.New("user not found")
	}

	// Remove old API key
	delete(a.apiKeys, user.APIKey)

	// Generate new API key
	newAPIKey, err := a.GenerateAPIKey()
	if err != nil {
		return "", fmt.Errorf("failed to generate new API key: %v", err)
	}

	// Update user and maps
	user.APIKey = "uma_" + newAPIKey
	a.apiKeys[user.APIKey] = user

	logger.Info("Regenerated API key for user %s", user.Username)
	return user.APIKey, nil
}

// GetAuthStats returns authentication statistics
func (a *AuthService) GetAuthStats() map[string]interface{} {
	stats := map[string]interface{}{
		"enabled":     a.config.Enabled,
		"total_users": len(a.users),
	}

	if a.config.Enabled {
		activeUsers := 0
		for _, user := range a.users {
			if user.Active {
				activeUsers++
			}
		}
		stats["active_users"] = activeUsers

		roleCount := make(map[Role]int)
		for _, user := range a.users {
			roleCount[user.Role]++
		}
		stats["roles"] = roleCount
	}

	return stats
}

// GetUserFromContext extracts user from request context
func GetUserFromContext(r *http.Request) *User {
	if user, ok := r.Context().Value(UserContextKey).(*User); ok {
		return user
	}
	return nil
}
