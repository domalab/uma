package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

// Custom context key types to avoid collisions
type contextKey string

const (
	requestIDKey contextKey = "request_id"
	loggerKey    contextKey = "logger"
)

// LogLevel represents the severity level of a log entry
type LogLevel int

const (
	// LogLevelDebug for detailed debugging information
	LogLevelDebug LogLevel = iota
	// LogLevelInfo for general information
	LogLevelInfo
	// LogLevelWarn for warning conditions
	LogLevelWarn
	// LogLevelError for error conditions
	LogLevelError
	// LogLevelFatal for fatal errors that cause program termination
	LogLevelFatal
)

// String returns the string representation of the log level
func (l LogLevel) String() string {
	switch l {
	case LogLevelDebug:
		return "DEBUG"
	case LogLevelInfo:
		return "INFO"
	case LogLevelWarn:
		return "WARN"
	case LogLevelError:
		return "ERROR"
	case LogLevelFatal:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// LogEntry represents a structured log entry
type LogEntry struct {
	Timestamp  time.Time              `json:"timestamp"`
	Level      string                 `json:"level"`
	Message    string                 `json:"message"`
	RequestID  string                 `json:"request_id,omitempty"`
	UserID     string                 `json:"user_id,omitempty"`
	Component  string                 `json:"component,omitempty"`
	Operation  string                 `json:"operation,omitempty"`
	Duration   string                 `json:"duration,omitempty"`
	StatusCode int                    `json:"status_code,omitempty"`
	Method     string                 `json:"method,omitempty"`
	Path       string                 `json:"path,omitempty"`
	RemoteAddr string                 `json:"remote_addr,omitempty"`
	UserAgent  string                 `json:"user_agent,omitempty"`
	Error      string                 `json:"error,omitempty"`
	Fields     map[string]interface{} `json:"fields,omitempty"`
	Caller     string                 `json:"caller,omitempty"`
	StackTrace string                 `json:"stack_trace,omitempty"`
}

// StructuredLogger provides structured logging capabilities
type StructuredLogger struct {
	level     LogLevel
	output    io.Writer
	mu        sync.RWMutex
	component string
	requestID string
	userID    string
	fields    map[string]interface{}
}

// NewStructuredLogger creates a new structured logger
func NewStructuredLogger(level LogLevel, output io.Writer, component string) *StructuredLogger {
	if output == nil {
		output = os.Stdout
	}

	return &StructuredLogger{
		level:     level,
		output:    output,
		component: component,
		fields:    make(map[string]interface{}),
	}
}

// WithRequestID returns a new logger with the specified request ID
func (l *StructuredLogger) WithRequestID(requestID string) *StructuredLogger {
	l.mu.RLock()
	defer l.mu.RUnlock()

	newLogger := &StructuredLogger{
		level:     l.level,
		output:    l.output,
		component: l.component,
		requestID: requestID,
		userID:    l.userID,
		fields:    make(map[string]interface{}),
	}

	// Copy existing fields
	for k, v := range l.fields {
		newLogger.fields[k] = v
	}

	return newLogger
}

// WithUserID returns a new logger with the specified user ID
func (l *StructuredLogger) WithUserID(userID string) *StructuredLogger {
	l.mu.RLock()
	defer l.mu.RUnlock()

	newLogger := &StructuredLogger{
		level:     l.level,
		output:    l.output,
		component: l.component,
		requestID: l.requestID,
		userID:    userID,
		fields:    make(map[string]interface{}),
	}

	// Copy existing fields
	for k, v := range l.fields {
		newLogger.fields[k] = v
	}

	return newLogger
}

// WithFields returns a new logger with additional fields
func (l *StructuredLogger) WithFields(fields map[string]interface{}) *StructuredLogger {
	l.mu.RLock()
	defer l.mu.RUnlock()

	newLogger := &StructuredLogger{
		level:     l.level,
		output:    l.output,
		component: l.component,
		requestID: l.requestID,
		userID:    l.userID,
		fields:    make(map[string]interface{}),
	}

	// Copy existing fields
	for k, v := range l.fields {
		newLogger.fields[k] = v
	}

	// Add new fields
	for k, v := range fields {
		newLogger.fields[k] = v
	}

	return newLogger
}

// Debug logs a debug message
func (l *StructuredLogger) Debug(message string) {
	l.log(LogLevelDebug, message, nil)
}

// Debugf logs a formatted debug message
func (l *StructuredLogger) Debugf(format string, args ...interface{}) {
	l.log(LogLevelDebug, fmt.Sprintf(format, args...), nil)
}

// Info logs an info message
func (l *StructuredLogger) Info(message string) {
	l.log(LogLevelInfo, message, nil)
}

// Infof logs a formatted info message
func (l *StructuredLogger) Infof(format string, args ...interface{}) {
	l.log(LogLevelInfo, fmt.Sprintf(format, args...), nil)
}

// Warn logs a warning message
func (l *StructuredLogger) Warn(message string) {
	l.log(LogLevelWarn, message, nil)
}

// Warnf logs a formatted warning message
func (l *StructuredLogger) Warnf(format string, args ...interface{}) {
	l.log(LogLevelWarn, fmt.Sprintf(format, args...), nil)
}

// Error logs an error message
func (l *StructuredLogger) Error(message string, err error) {
	l.log(LogLevelError, message, err)
}

// Errorf logs a formatted error message
func (l *StructuredLogger) Errorf(format string, args ...interface{}) {
	l.log(LogLevelError, fmt.Sprintf(format, args...), nil)
}

// Fatal logs a fatal message and exits
func (l *StructuredLogger) Fatal(message string, err error) {
	l.log(LogLevelFatal, message, err)
	os.Exit(1)
}

// Fatalf logs a formatted fatal message and exits
func (l *StructuredLogger) Fatalf(format string, args ...interface{}) {
	l.log(LogLevelFatal, fmt.Sprintf(format, args...), nil)
	os.Exit(1)
}

// LogHTTPRequest logs an HTTP request with structured data
func (l *StructuredLogger) LogHTTPRequest(r *http.Request, statusCode int, duration time.Duration) {
	entry := l.createLogEntry(LogLevelInfo, "HTTP request processed")
	entry.Method = r.Method
	entry.Path = r.URL.Path
	entry.RemoteAddr = r.RemoteAddr
	entry.UserAgent = r.UserAgent()
	entry.StatusCode = statusCode
	entry.Duration = duration.String()
	entry.Operation = "http_request"

	l.writeLogEntry(entry)
}

// LogOperation logs a general operation with timing
func (l *StructuredLogger) LogOperation(operation string, duration time.Duration, err error) {
	level := LogLevelInfo
	message := fmt.Sprintf("Operation '%s' completed", operation)

	if err != nil {
		level = LogLevelError
		message = fmt.Sprintf("Operation '%s' failed", operation)
	}

	entry := l.createLogEntry(level, message)
	entry.Operation = operation
	entry.Duration = duration.String()

	if err != nil {
		entry.Error = err.Error()
	}

	l.writeLogEntry(entry)
}

// log is the internal logging method
func (l *StructuredLogger) log(level LogLevel, message string, err error) {
	if level < l.level {
		return
	}

	entry := l.createLogEntry(level, message)

	if err != nil {
		entry.Error = err.Error()

		// Add stack trace for errors
		if level >= LogLevelError {
			entry.StackTrace = l.getStackTrace()
		}
	}

	l.writeLogEntry(entry)
}

// createLogEntry creates a new log entry with common fields
func (l *StructuredLogger) createLogEntry(level LogLevel, message string) *LogEntry {
	l.mu.RLock()
	defer l.mu.RUnlock()

	entry := &LogEntry{
		Timestamp: time.Now().UTC(),
		Level:     level.String(),
		Message:   message,
		RequestID: l.requestID,
		UserID:    l.userID,
		Component: l.component,
		Caller:    l.getCaller(),
	}

	// Copy fields
	if len(l.fields) > 0 {
		entry.Fields = make(map[string]interface{})
		for k, v := range l.fields {
			entry.Fields[k] = v
		}
	}

	return entry
}

// writeLogEntry writes the log entry to the output
func (l *StructuredLogger) writeLogEntry(entry *LogEntry) {
	l.mu.Lock()
	defer l.mu.Unlock()

	jsonData, err := json.Marshal(entry)
	if err != nil {
		// Fallback to standard logging if JSON marshaling fails
		log.Printf("Failed to marshal log entry: %v", err)
		log.Printf("%s [%s] %s", entry.Timestamp.Format(time.RFC3339), entry.Level, entry.Message)
		return
	}

	l.output.Write(jsonData)
	l.output.Write([]byte("\n"))
}

// getCaller returns information about the calling function
func (l *StructuredLogger) getCaller() string {
	// Skip this function, the log function, and the public logging method
	_, file, line, ok := runtime.Caller(3)
	if !ok {
		return "unknown"
	}

	// Get just the filename, not the full path
	parts := strings.Split(file, "/")
	filename := parts[len(parts)-1]

	return fmt.Sprintf("%s:%d", filename, line)
}

// getStackTrace returns a stack trace for error logging
func (l *StructuredLogger) getStackTrace() string {
	buf := make([]byte, 1024)
	for {
		n := runtime.Stack(buf, false)
		if n < len(buf) {
			return string(buf[:n])
		}
		buf = make([]byte, 2*len(buf))
	}
}

// LoggingMiddleware creates middleware that logs HTTP requests
func LoggingMiddleware(logger *StructuredLogger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Get or generate request ID
			requestID := r.Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = GenerateRequestID()
				r.Header.Set("X-Request-ID", requestID)
			}

			// Create request-scoped logger
			requestLogger := logger.WithRequestID(requestID)

			// Add logger to request context
			ctx := context.WithValue(r.Context(), loggerKey, requestLogger)
			r = r.WithContext(ctx)

			// Wrap response writer to capture status code
			wrapped := &responseWriter{ResponseWriter: w, statusCode: 200}

			// Log request start
			requestLogger.WithFields(map[string]interface{}{
				"method":      r.Method,
				"path":        r.URL.Path,
				"remote_addr": r.RemoteAddr,
				"user_agent":  r.UserAgent(),
			}).Info("HTTP request started")

			// Process request
			next.ServeHTTP(wrapped, r)

			// Log request completion
			duration := time.Since(start)
			requestLogger.LogHTTPRequest(r, wrapped.statusCode, duration)
		})
	}
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

// GetLoggerFromContext retrieves the logger from the request context
func GetLoggerFromContext(ctx context.Context) *StructuredLogger {
	if logger, ok := ctx.Value(loggerKey).(*StructuredLogger); ok {
		return logger
	}
	// Return a default logger if none found
	return NewStructuredLogger(LogLevelInfo, os.Stdout, "unknown")
}
