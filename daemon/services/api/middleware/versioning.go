package middleware

import (
	"context"
	"net/http"
	"regexp"
	"strings"
)

// Context key for API version
const (
	APIVersionKey contextKey = "api_version"
)

// Versioning returns a middleware that handles API version negotiation
func Versioning() func(http.Handler) http.Handler {
	return VersioningWithConfig(DefaultVersioningConfig())
}

// VersioningWithConfig returns a versioning middleware with custom configuration
func VersioningWithConfig(config VersioningConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Parse Accept header for version negotiation
			acceptHeader := r.Header.Get("Accept")
			apiVersion := parseAPIVersion(acceptHeader, config)

			// Set the negotiated version in request context
			ctx := context.WithValue(r.Context(), APIVersionKey, apiVersion)
			r = r.WithContext(ctx)

			// Set version information in response headers
			w.Header().Set("X-API-Version", apiVersion)
			w.Header().Set("X-API-Supported-Versions", strings.Join(config.SupportedVersions, ", "))

			// Add deprecation warning if version is deprecated
			if isVersionDeprecated(apiVersion, config.DeprecatedVersions) {
				w.Header().Set("X-API-Deprecation-Warning", "This API version is deprecated")
			}

			next.ServeHTTP(w, r)
		})
	}
}

// VersioningConfig represents versioning middleware configuration
type VersioningConfig struct {
	DefaultVersion      string   `json:"default_version"`
	SupportedVersions   []string `json:"supported_versions"`
	DeprecatedVersions  []string `json:"deprecated_versions"`
	HeaderName          string   `json:"header_name"`
	AcceptHeaderPattern string   `json:"accept_header_pattern"`
}

// DefaultVersioningConfig returns a default versioning configuration
func DefaultVersioningConfig() VersioningConfig {
	return VersioningConfig{
		DefaultVersion:      "v1",
		SupportedVersions:   []string{"v1"},
		DeprecatedVersions:  []string{},
		HeaderName:          "X-API-Version",
		AcceptHeaderPattern: `application/vnd\.uma\.([^+]+)\+json`,
	}
}

// parseAPIVersion extracts API version from Accept header
func parseAPIVersion(acceptHeader string, config VersioningConfig) string {
	if acceptHeader == "" {
		return config.DefaultVersion
	}

	// Try to match the versioned media type pattern
	re := regexp.MustCompile(config.AcceptHeaderPattern)
	matches := re.FindStringSubmatch(acceptHeader)

	if len(matches) > 1 {
		version := matches[1]
		// Check if the version is supported
		for _, supportedVersion := range config.SupportedVersions {
			if version == supportedVersion {
				return version
			}
		}
	}

	// Check for simple version patterns like "v1", "v2"
	versionPattern := regexp.MustCompile(`v(\d+)`)
	matches = versionPattern.FindStringSubmatch(acceptHeader)
	if len(matches) > 0 {
		version := matches[0]
		for _, supportedVersion := range config.SupportedVersions {
			if version == supportedVersion {
				return version
			}
		}
	}

	return config.DefaultVersion
}

// isVersionDeprecated checks if a version is deprecated
func isVersionDeprecated(version string, deprecatedVersions []string) bool {
	for _, deprecatedVersion := range deprecatedVersions {
		if version == deprecatedVersion {
			return true
		}
	}
	return false
}

// GetAPIVersionFromContext gets the API version from request context
func GetAPIVersionFromContext(r *http.Request) string {
	if version, ok := r.Context().Value(APIVersionKey).(string); ok {
		return version
	}
	return "v1" // Default fallback
}

// VersionedHandler represents a handler for a specific API version
type VersionedHandler struct {
	Version string
	Handler http.HandlerFunc
}

// VersionRouter routes requests to different handlers based on API version
type VersionRouter struct {
	handlers map[string]http.HandlerFunc
	fallback http.HandlerFunc
}

// NewVersionRouter creates a new version router
func NewVersionRouter() *VersionRouter {
	return &VersionRouter{
		handlers: make(map[string]http.HandlerFunc),
	}
}

// AddHandler adds a handler for a specific version
func (vr *VersionRouter) AddHandler(version string, handler http.HandlerFunc) {
	vr.handlers[version] = handler
}

// SetFallback sets the fallback handler for unsupported versions
func (vr *VersionRouter) SetFallback(handler http.HandlerFunc) {
	vr.fallback = handler
}

// ServeHTTP implements http.Handler interface
func (vr *VersionRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	version := GetAPIVersionFromContext(r)

	if handler, exists := vr.handlers[version]; exists {
		handler.ServeHTTP(w, r)
		return
	}

	if vr.fallback != nil {
		vr.fallback.ServeHTTP(w, r)
		return
	}

	// No handler found and no fallback
	http.Error(w, "Unsupported API version", http.StatusNotAcceptable)
}
