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
