package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	// Logger is the global structured logger instance
	Logger zerolog.Logger

	// Maintain backward compatibility with existing logger functions
	initialized bool
)

func init() {
	initStructuredLogger()
}

// initStructuredLogger initializes the structured logger with UMA-specific configuration
func initStructuredLogger() {
	// Configure zerolog for human-readable output in development
	zerolog.TimeFieldFormat = time.RFC3339

	// Create console writer for colored output (similar to existing logger)
	consoleWriter := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: "15:04:05",
		NoColor:    false,
	}

	// Initialize the global logger with UMA context
	Logger = zerolog.New(consoleWriter).
		With().
		Timestamp().
		Str("service", "uma").
		Logger()

	initialized = true
}

// Structured logging functions for different components

// LogAPIRequest logs API requests with structured fields
func LogAPIRequest(requestID, method, path string, statusCode int, duration time.Duration) {
	Logger.Info().
		Str("component", "api").
		Str("request_id", requestID).
		Str("method", method).
		Str("path", path).
		Int("status_code", statusCode).
		Dur("duration", duration).
		Msg("API request completed")
}

// LogBulkOperation logs bulk Docker operations with detailed metrics
func LogBulkOperation(operation string, total, succeeded, failed int, duration time.Duration, requestID string) {
	Logger.Info().
		Str("component", "docker").
		Str("operation_type", "bulk").
		Str("operation", operation).
		Str("request_id", requestID).
		Int("total", total).
		Int("succeeded", succeeded).
		Int("failed", failed).
		Dur("duration", duration).
		Float64("success_rate", float64(succeeded)/float64(total)*100).
		Msg("Bulk operation completed")
}

// LogWebSocketConnection logs WebSocket connection events
func LogWebSocketConnection(endpoint, event, clientID string, connectionCount int) {
	Logger.Info().
		Str("component", "websocket").
		Str("endpoint", endpoint).
		Str("event", event).
		Str("client_id", clientID).
		Int("active_connections", connectionCount).
		Msg("WebSocket connection event")
}

// LogHealthCheck logs health check results with dependency status
func LogHealthCheck(status string, dependencies map[string]string, duration time.Duration, requestID string) {
	event := Logger.Info().
		Str("component", "health").
		Str("status", status).
		Str("request_id", requestID).
		Dur("duration", duration)

	// Add dependency statuses as structured fields
	for service, serviceStatus := range dependencies {
		event = event.Str("dep_"+service, serviceStatus)
	}

	event.Msg("Health check completed")
}

// LogConfigLoad logs configuration loading events
func LogConfigLoad(configType, path string, success bool, errorMsg string) {
	event := Logger.Info().
		Str("component", "config").
		Str("config_type", configType).
		Str("path", path).
		Bool("success", success)

	if !success && errorMsg != "" {
		event = event.Str("error", errorMsg)
	}

	event.Msg("Configuration loaded")
}

// LogValidationError logs input validation errors with details
func LogValidationError(component, operation, requestID, errorMsg string) {
	Logger.Warn().
		Str("component", component).
		Str("operation", operation).
		Str("request_id", requestID).
		Str("validation_error", errorMsg).
		Msg("Input validation failed")
}

// LogMetricsCollection logs metrics collection events
func LogMetricsCollection(metricType string, value float64, labels map[string]string) {
	event := Logger.Debug().
		Str("component", "metrics").
		Str("metric_type", metricType).
		Float64("value", value)

	// Add labels as structured fields
	for key, val := range labels {
		event = event.Str("label_"+key, val)
	}

	event.Msg("Metric collected")
}

// LogAsyncOperation logs async operation lifecycle events
func LogAsyncOperation(operationID, operationType, event string, progress int, requestID string, context map[string]interface{}) {
	event_logger := Logger.Info().
		Str("component", "async").
		Str("operation_id", operationID).
		Str("operation_type", operationType).
		Str("event", event).
		Int("progress", progress).
		Str("request_id", requestID)

	// Add additional context fields
	for key, value := range context {
		event_logger = event_logger.Interface(key, value)
	}

	event_logger.Msg("Async operation event")
}

// LogCacheOperation logs cache hit/miss and performance metrics
func LogCacheOperation(operation, key string, hit bool, duration time.Duration, size int, requestID string) {
	Logger.Debug().
		Str("component", "cache").
		Str("operation", operation).
		Str("key", key).
		Bool("hit", hit).
		Dur("duration", duration).
		Int("size_bytes", size).
		Str("request_id", requestID).
		Msg("Cache operation")
}

// LogRateLimitEvent logs rate limiting events
func LogRateLimitEvent(clientIP, operationType string, allowed bool, remaining int, resetTime time.Time, requestID string) {
	level := Logger.Debug()
	if !allowed {
		level = Logger.Warn()
	}

	level.
		Str("component", "rate_limiter").
		Str("client_ip", clientIP).
		Str("operation_type", operationType).
		Bool("allowed", allowed).
		Int("remaining", remaining).
		Time("reset_time", resetTime).
		Str("request_id", requestID).
		Msg("Rate limit check")
}

// LogPerformanceMetrics logs detailed performance metrics for operations
func LogPerformanceMetrics(operation string, duration time.Duration, memoryUsed int64, cpuTime time.Duration, requestID string, context map[string]interface{}) {
	event := Logger.Info().
		Str("component", "performance").
		Str("operation", operation).
		Dur("duration", duration).
		Int64("memory_used_bytes", memoryUsed).
		Dur("cpu_time", cpuTime).
		Str("request_id", requestID)

	// Add performance context
	for key, value := range context {
		event = event.Interface(key, value)
	}

	event.Msg("Performance metrics")
}

// LogSecurityEvent logs security-related events
func LogSecurityEvent(eventType, clientIP, userAgent, requestID string, success bool, details map[string]interface{}) {
	level := Logger.Info()
	if !success {
		level = Logger.Warn()
	}

	event := level.
		Str("component", "security").
		Str("event_type", eventType).
		Str("client_ip", clientIP).
		Str("user_agent", userAgent).
		Str("request_id", requestID).
		Bool("success", success)

	// Add security event details
	for key, value := range details {
		event = event.Interface(key, value)
	}

	event.Msg("Security event")
}

// LogErrorWithContext logs errors with rich context information
func LogErrorWithContext(component, operation string, err error, requestID string, context map[string]interface{}) {
	event := Logger.Error().
		Str("component", component).
		Str("operation", operation).
		Err(err).
		Str("request_id", requestID)

	// Add error context
	for key, value := range context {
		event = event.Interface(key, value)
	}

	event.Msg("Operation failed")
}

// Backward compatibility functions that wrap the existing logger behavior
// These maintain the same interface as the existing logger package

// Info logs an info message (backward compatible)
func Info(format string, args ...interface{}) {
	if initialized {
		Logger.Info().Msgf(format, args...)
	} else {
		// Fallback to existing logger if not initialized
		log.Info().Msgf(format, args...)
	}
}

// Warn logs a warning message (backward compatible)
func Warn(format string, args ...interface{}) {
	if initialized {
		Logger.Warn().Msgf(format, args...)
	} else {
		log.Warn().Msgf(format, args...)
	}
}

// Error logs an error message (backward compatible)
func Error(format string, args ...interface{}) {
	if initialized {
		Logger.Error().Msgf(format, args...)
	} else {
		log.Error().Msgf(format, args...)
	}
}

// Debug logs a debug message (backward compatible)
func Debug(format string, args ...interface{}) {
	if initialized {
		Logger.Debug().Msgf(format, args...)
	} else {
		log.Debug().Msgf(format, args...)
	}
}

// Fatal logs a fatal message and exits (backward compatible)
func Fatal(format string, args ...interface{}) {
	if initialized {
		Logger.Fatal().Msgf(format, args...)
	} else {
		log.Fatal().Msgf(format, args...)
	}
}

// GetLogger returns the structured logger instance for advanced usage
func GetLogger() zerolog.Logger {
	return Logger
}

// WithContext creates a logger with additional context fields
func WithContext(fields map[string]interface{}) zerolog.Logger {
	ctx := Logger.With()
	for key, value := range fields {
		ctx = ctx.Interface(key, value)
	}
	return ctx.Logger()
}
