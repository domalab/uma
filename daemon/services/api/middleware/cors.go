package middleware

import (
	"net/http"
	"strings"
)

// CORS returns a middleware that handles Cross-Origin Resource Sharing (CORS) headers
func CORS() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Set CORS headers
			origin := r.Header.Get("Origin")
			
			// Allow all origins for internal API (since this is an internal-only service)
			// In production, you might want to restrict this to specific origins
			if origin != "" {
				// Validate origin to prevent malicious origins
				if isValidOrigin(origin) {
					w.Header().Set("Access-Control-Allow-Origin", origin)
				} else {
					// For invalid origins, set a safe default
					w.Header().Set("Access-Control-Allow-Origin", "*")
				}
			} else {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			}
			
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Request-ID")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Max-Age", "86400") // 24 hours

			// Handle preflight requests
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// isValidOrigin checks if the origin is valid and safe
func isValidOrigin(origin string) bool {
	// Block obviously malicious origins
	if strings.Contains(origin, "javascript:") ||
		strings.Contains(origin, "data:") ||
		strings.Contains(origin, "vbscript:") {
		return false
	}
	
	// For UMA, we're more permissive since it's an internal API
	// In a production environment, you'd want to maintain a whitelist
	return true
}
