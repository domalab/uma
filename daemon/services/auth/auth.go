package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/domalab/uma/daemon/domain"
	"github.com/domalab/uma/daemon/logger"
)

// AuthService handles authentication and authorization
type AuthService struct {
	config domain.AuthConfig
}

// NewAuthService creates a new authentication service
func NewAuthService(config domain.AuthConfig) *AuthService {
	return &AuthService{
		config: config,
	}
}

// GenerateAPIKey generates a new API key
func (a *AuthService) GenerateAPIKey() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate API key: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// ValidateAPIKey validates an API key
func (a *AuthService) ValidateAPIKey(providedKey string) bool {
	if !a.config.Enabled {
		return true // Authentication disabled
	}

	if a.config.APIKey == "" {
		logger.Yellow("API key authentication enabled but no key configured")
		return false
	}

	// Use constant-time comparison to prevent timing attacks
	return subtle.ConstantTimeCompare([]byte(providedKey), []byte(a.config.APIKey)) == 1
}

// AuthMiddleware returns an HTTP middleware for authentication
func (a *AuthService) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !a.config.Enabled {
			next.ServeHTTP(w, r)
			return
		}

		// Check for API key in header
		apiKey := r.Header.Get("X-API-Key")
		if apiKey == "" {
			// Check for API key in Authorization header
			auth := r.Header.Get("Authorization")
			if strings.HasPrefix(auth, "Bearer ") {
				apiKey = strings.TrimPrefix(auth, "Bearer ")
			}
		}

		if apiKey == "" {
			a.writeUnauthorized(w, "Missing API key")
			return
		}

		if !a.ValidateAPIKey(apiKey) {
			a.writeUnauthorized(w, "Invalid API key")
			return
		}

		// Add authentication info to request context if needed
		next.ServeHTTP(w, r)
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
