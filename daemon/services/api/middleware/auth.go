package middleware

import (
	"net/http"

	"github.com/domalab/uma/daemon/services/auth"
)

// Auth returns a middleware that handles authentication
func Auth(authService *auth.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if authService != nil && authService.IsEnabled() {
			return authService.AuthMiddleware(next)
		}
		return next
	}
}

// AuthWithConfig returns an authentication middleware with custom configuration
func AuthWithConfig(authService *auth.AuthService, config AuthConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip authentication for excluded paths
			if shouldSkipAuth(r.URL.Path, config.ExcludedPaths) {
				next.ServeHTTP(w, r)
				return
			}

			// Skip authentication for certain methods if configured
			if shouldSkipAuthForMethod(r.Method, config.ExcludedMethods) {
				next.ServeHTTP(w, r)
				return
			}

			// Use auth service if available and enabled
			if authService != nil && authService.IsEnabled() {
				authService.AuthMiddleware(next).ServeHTTP(w, r)
				return
			}

			// If auth is required but service is not available, deny access
			if config.RequireAuth {
				http.Error(w, "Authentication required", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// AuthConfig represents authentication middleware configuration
type AuthConfig struct {
	RequireAuth     bool     `json:"require_auth"`
	ExcludedPaths   []string `json:"excluded_paths"`
	ExcludedMethods []string `json:"excluded_methods"`
	TokenHeader     string   `json:"token_header"`
	TokenPrefix     string   `json:"token_prefix"`
}

// DefaultAuthConfig returns a default authentication configuration
func DefaultAuthConfig() AuthConfig {
	return AuthConfig{
		RequireAuth: false, // Default to optional auth for backward compatibility
		ExcludedPaths: []string{
			"/api/v1/health",
			"/api/v1/auth/login",
			"/metrics",
			// Removed OpenAPI documentation paths - system removed
		},
		ExcludedMethods: []string{
			"OPTIONS",
		},
		TokenHeader: "Authorization",
		TokenPrefix: "Bearer ",
	}
}

// shouldSkipAuth determines if authentication should be skipped for a path
func shouldSkipAuth(path string, excludedPaths []string) bool {
	for _, excludedPath := range excludedPaths {
		if path == excludedPath || (len(excludedPath) > 0 && excludedPath[len(excludedPath)-1] == '/' &&
			len(path) > len(excludedPath) && path[:len(excludedPath)] == excludedPath) {
			return true
		}
	}
	return false
}

// shouldSkipAuthForMethod determines if authentication should be skipped for a method
func shouldSkipAuthForMethod(method string, excludedMethods []string) bool {
	for _, excludedMethod := range excludedMethods {
		if method == excludedMethod {
			return true
		}
	}
	return false
}

// RequireAuth returns a middleware that requires authentication
func RequireAuth(authService *auth.AuthService) func(http.Handler) http.Handler {
	config := DefaultAuthConfig()
	config.RequireAuth = true
	return AuthWithConfig(authService, config)
}

// OptionalAuth returns a middleware that allows optional authentication
func OptionalAuth(authService *auth.AuthService) func(http.Handler) http.Handler {
	config := DefaultAuthConfig()
	config.RequireAuth = false
	return AuthWithConfig(authService, config)
}
