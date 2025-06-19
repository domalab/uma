package middleware

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/domalab/uma/daemon/logger"
)

// Logging returns a middleware that logs HTTP requests with structured logging
func Logging() func(http.Handler) http.Handler {
	return LoggingWithConfig(DefaultLoggingConfig())
}

// LoggingWithConfig returns a logging middleware with custom configuration
func LoggingWithConfig(config LoggingConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Create a response writer wrapper to capture status code
			wrapper := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(wrapper, r)

			duration := time.Since(start)
			requestID := GetRequestIDFromContext(r)

			// Skip logging for certain paths if configured
			if shouldSkipLogging(r.URL.Path, config.SkipPaths) {
				return
			}

			// Use structured logging for API requests
			if config.StructuredLogging {
				logger.LogAPIRequest(requestID, r.Method, r.URL.Path, wrapper.statusCode, duration)
			}

			// Log additional context for errors
			if wrapper.statusCode >= 400 && config.LogErrors {
				context := map[string]interface{}{
					"user_agent":     r.UserAgent(),
					"content_type":   r.Header.Get("Content-Type"),
					"content_length": r.ContentLength,
					"query_params":   r.URL.RawQuery,
					"remote_addr":    r.RemoteAddr,
				}

				if wrapper.statusCode >= 500 {
					logger.LogErrorWithContext("api", "http_request",
						fmt.Errorf("HTTP %d error for %s %s", wrapper.statusCode, r.Method, r.URL.Path),
						requestID, context)
				}
			}

			// Backward compatibility: also log with colored output
			if config.ColoredOutput {
				if requestID != "" {
					logger.LightGreen("HTTP %s %s %d %v [%s]", r.Method, r.URL.Path, wrapper.statusCode, duration, requestID)
				} else {
					logger.LightGreen("HTTP %s %s %d %v", r.Method, r.URL.Path, wrapper.statusCode, duration)
				}
			}

			// Custom logger if provided
			if config.CustomLogger != nil {
				config.CustomLogger(LogEntry{
					RequestID:   requestID,
					Method:      r.Method,
					Path:        r.URL.Path,
					StatusCode:  wrapper.statusCode,
					Duration:    duration,
					UserAgent:   r.UserAgent(),
					RemoteAddr:  r.RemoteAddr,
					ContentType: r.Header.Get("Content-Type"),
					QueryParams: r.URL.RawQuery,
					Timestamp:   start,
				})
			}
		})
	}
}

// LoggingConfig represents logging middleware configuration
type LoggingConfig struct {
	StructuredLogging bool           `json:"structured_logging"`
	ColoredOutput     bool           `json:"colored_output"`
	LogErrors         bool           `json:"log_errors"`
	SkipPaths         []string       `json:"skip_paths"`
	CustomLogger      func(LogEntry) `json:"-"`
}

// LogEntry represents a log entry for HTTP requests
type LogEntry struct {
	RequestID   string        `json:"request_id"`
	Method      string        `json:"method"`
	Path        string        `json:"path"`
	StatusCode  int           `json:"status_code"`
	Duration    time.Duration `json:"duration"`
	UserAgent   string        `json:"user_agent"`
	RemoteAddr  string        `json:"remote_addr"`
	ContentType string        `json:"content_type"`
	QueryParams string        `json:"query_params"`
	Timestamp   time.Time     `json:"timestamp"`
}

// DefaultLoggingConfig returns a default logging configuration
func DefaultLoggingConfig() LoggingConfig {
	return LoggingConfig{
		StructuredLogging: true,
		ColoredOutput:     true,
		LogErrors:         true,
		SkipPaths: []string{
			"/health",
			"/metrics",
			"/favicon.ico",
		},
		CustomLogger: nil,
	}
}

// shouldSkipLogging determines if logging should be skipped for a path
func shouldSkipLogging(path string, skipPaths []string) bool {
	for _, skipPath := range skipPaths {
		if path == skipPath {
			return true
		}
	}
	return false
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	return rw.ResponseWriter.Write(b)
}

// Flush implements http.Flusher interface
func (rw *responseWriter) Flush() {
	if flusher, ok := rw.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

// Hijack implements http.Hijacker interface
func (rw *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := rw.ResponseWriter.(http.Hijacker); ok {
		return hijacker.Hijack()
	}
	return nil, nil, fmt.Errorf("response writer does not support hijacking")
}
