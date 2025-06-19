package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

// Context key types for type safety
type contextKey string

const (
	RequestIDKey contextKey = "request_id"
)

// RequestID returns a middleware that adds request ID tracking for debugging and tracing
func RequestID() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if request already has an ID from client
			requestID := r.Header.Get("X-Request-ID")

			// Generate a new request ID if not provided
			if requestID == "" {
				requestID = generateRequestID()
			}

			// Set the request ID in response headers
			w.Header().Set("X-Request-ID", requestID)

			// Add request ID to request context for use in handlers
			ctx := context.WithValue(r.Context(), RequestIDKey, requestID)
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}

// RequestIDWithConfig returns a request ID middleware with custom configuration
func RequestIDWithConfig(config RequestIDConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var requestID string

			// Check if request already has an ID from client
			if config.AllowClientRequestID {
				requestID = r.Header.Get(config.RequestIDHeader)
			}

			// Generate a new request ID if not provided or not allowed
			if requestID == "" {
				if config.Generator != nil {
					requestID = config.Generator()
				} else {
					requestID = generateRequestID()
				}
			}

			// Validate request ID format if validator is provided
			if config.Validator != nil && !config.Validator(requestID) {
				requestID = generateRequestID()
			}

			// Set the request ID in response headers
			w.Header().Set(config.ResponseIDHeader, requestID)

			// Add request ID to request context for use in handlers
			ctx := context.WithValue(r.Context(), config.ContextKey, requestID)
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}

// RequestIDConfig represents request ID middleware configuration
type RequestIDConfig struct {
	AllowClientRequestID bool              `json:"allow_client_request_id"`
	RequestIDHeader      string            `json:"request_id_header"`
	ResponseIDHeader     string            `json:"response_id_header"`
	ContextKey           contextKey        `json:"-"`
	Generator            func() string     `json:"-"`
	Validator            func(string) bool `json:"-"`
}

// DefaultRequestIDConfig returns a default request ID configuration
func DefaultRequestIDConfig() RequestIDConfig {
	return RequestIDConfig{
		AllowClientRequestID: true,
		RequestIDHeader:      "X-Request-ID",
		ResponseIDHeader:     "X-Request-ID",
		ContextKey:           RequestIDKey,
		Generator:            generateRequestID,
		Validator:            nil, // No validation by default
	}
}

// generateRequestID creates a unique request ID for tracing using UUID
func generateRequestID() string {
	return uuid.New().String()
}

// GetRequestIDFromContext gets the request ID from request context
func GetRequestIDFromContext(r *http.Request) string {
	if requestID, ok := r.Context().Value(RequestIDKey).(string); ok {
		return requestID
	}
	return ""
}

// GetRequestIDFromContextWithKey gets the request ID from request context using a custom key
func GetRequestIDFromContextWithKey(r *http.Request, key contextKey) string {
	if requestID, ok := r.Context().Value(key).(string); ok {
		return requestID
	}
	return ""
}
