package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// TestLogLevel tests the LogLevel type and its string representation
func TestLogLevel(t *testing.T) {
	tests := []struct {
		level    LogLevel
		expected string
	}{
		{LogLevelDebug, "DEBUG"},
		{LogLevelInfo, "INFO"},
		{LogLevelWarn, "WARN"},
		{LogLevelError, "ERROR"},
		{LogLevelFatal, "FATAL"},
		{LogLevel(999), "UNKNOWN"},
	}

	for _, test := range tests {
		if test.level.String() != test.expected {
			t.Errorf("Expected %s, got %s", test.expected, test.level.String())
		}
	}
}

// TestStructuredLogger tests the StructuredLogger functionality
func TestStructuredLogger(t *testing.T) {
	t.Run("NewStructuredLogger", func(t *testing.T) {
		var buf bytes.Buffer
		logger := NewStructuredLogger(LogLevelInfo, &buf, "test-component")

		if logger == nil {
			t.Error("Expected non-nil logger")
			return
		}
		if logger.level != LogLevelInfo {
			t.Errorf("Expected level %v, got %v", LogLevelInfo, logger.level)
		}
		if logger.component != "test-component" {
			t.Errorf("Expected component 'test-component', got '%s'", logger.component)
		}
	})

	t.Run("WithRequestID", func(t *testing.T) {
		var buf bytes.Buffer
		logger := NewStructuredLogger(LogLevelInfo, &buf, "test")
		requestLogger := logger.WithRequestID("req-123")

		if requestLogger.requestID != "req-123" {
			t.Errorf("Expected request ID 'req-123', got '%s'", requestLogger.requestID)
		}

		// Original logger should not be modified
		if logger.requestID != "" {
			t.Error("Original logger should not have request ID")
		}
	})

	t.Run("WithUserID", func(t *testing.T) {
		var buf bytes.Buffer
		logger := NewStructuredLogger(LogLevelInfo, &buf, "test")
		userLogger := logger.WithUserID("user-456")

		if userLogger.userID != "user-456" {
			t.Errorf("Expected user ID 'user-456', got '%s'", userLogger.userID)
		}

		// Original logger should not be modified
		if logger.userID != "" {
			t.Error("Original logger should not have user ID")
		}
	})

	t.Run("WithFields", func(t *testing.T) {
		var buf bytes.Buffer
		logger := NewStructuredLogger(LogLevelInfo, &buf, "test")

		fields := map[string]interface{}{
			"key1": "value1",
			"key2": 42,
		}
		fieldLogger := logger.WithFields(fields)

		if len(fieldLogger.fields) != 2 {
			t.Errorf("Expected 2 fields, got %d", len(fieldLogger.fields))
		}
		if fieldLogger.fields["key1"] != "value1" {
			t.Error("Expected field key1 to have value 'value1'")
		}
		if fieldLogger.fields["key2"] != 42 {
			t.Error("Expected field key2 to have value 42")
		}

		// Original logger should not be modified
		if len(logger.fields) != 0 {
			t.Error("Original logger should not have fields")
		}
	})
}

// TestLoggingMethods tests the various logging methods
func TestLoggingMethods(t *testing.T) {
	t.Run("InfoLogging", func(t *testing.T) {
		var buf bytes.Buffer
		logger := NewStructuredLogger(LogLevelInfo, &buf, "test")

		logger.Info("Test info message")

		var entry LogEntry
		if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
			t.Fatalf("Failed to unmarshal log entry: %v", err)
		}

		if entry.Level != "INFO" {
			t.Errorf("Expected level INFO, got %s", entry.Level)
		}
		if entry.Message != "Test info message" {
			t.Errorf("Expected message 'Test info message', got '%s'", entry.Message)
		}
		if entry.Component != "test" {
			t.Errorf("Expected component 'test', got '%s'", entry.Component)
		}
	})

	t.Run("ErrorLogging", func(t *testing.T) {
		var buf bytes.Buffer
		logger := NewStructuredLogger(LogLevelInfo, &buf, "test")

		testErr := &testError{message: "test error"}
		logger.Error("Test error message", testErr)

		var entry LogEntry
		if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
			t.Fatalf("Failed to unmarshal log entry: %v", err)
		}

		if entry.Level != "ERROR" {
			t.Errorf("Expected level ERROR, got %s", entry.Level)
		}
		if entry.Error != "test error" {
			t.Errorf("Expected error 'test error', got '%s'", entry.Error)
		}
		if entry.StackTrace == "" {
			t.Error("Expected stack trace for error log")
		}
	})

	t.Run("DebugLoggingFiltered", func(t *testing.T) {
		var buf bytes.Buffer
		logger := NewStructuredLogger(LogLevelInfo, &buf, "test")

		logger.Debug("This should not appear")

		if buf.Len() > 0 {
			t.Error("Debug message should be filtered out at INFO level")
		}
	})

	t.Run("FormattedLogging", func(t *testing.T) {
		var buf bytes.Buffer
		logger := NewStructuredLogger(LogLevelInfo, &buf, "test")

		logger.Infof("Test message with %s and %d", "string", 42)

		var entry LogEntry
		if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
			t.Fatalf("Failed to unmarshal log entry: %v", err)
		}

		expected := "Test message with string and 42"
		if entry.Message != expected {
			t.Errorf("Expected message '%s', got '%s'", expected, entry.Message)
		}
	})
}

// TestHTTPRequestLogging tests HTTP request logging
func TestHTTPRequestLogging(t *testing.T) {
	t.Run("LogHTTPRequest", func(t *testing.T) {
		var buf bytes.Buffer
		logger := NewStructuredLogger(LogLevelInfo, &buf, "test")

		req := httptest.NewRequest("GET", "/api/v2/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		req.Header.Set("User-Agent", "Test-Agent/1.0")

		duration := 150 * time.Millisecond
		logger.LogHTTPRequest(req, 200, duration)

		var entry LogEntry
		if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
			t.Fatalf("Failed to unmarshal log entry: %v", err)
		}

		if entry.Method != "GET" {
			t.Errorf("Expected method GET, got %s", entry.Method)
		}
		if entry.Path != "/api/v2/test" {
			t.Errorf("Expected path '/api/v2/test', got '%s'", entry.Path)
		}
		if entry.StatusCode != 200 {
			t.Errorf("Expected status code 200, got %d", entry.StatusCode)
		}
		if entry.RemoteAddr != "192.168.1.1:12345" {
			t.Errorf("Expected remote addr '192.168.1.1:12345', got '%s'", entry.RemoteAddr)
		}
		if entry.UserAgent != "Test-Agent/1.0" {
			t.Errorf("Expected user agent 'Test-Agent/1.0', got '%s'", entry.UserAgent)
		}
		if entry.Duration != duration.String() {
			t.Errorf("Expected duration '%s', got '%s'", duration.String(), entry.Duration)
		}
		if entry.Operation != "http_request" {
			t.Errorf("Expected operation 'http_request', got '%s'", entry.Operation)
		}
	})
}

// TestOperationLogging tests operation logging
func TestOperationLogging(t *testing.T) {
	t.Run("SuccessfulOperation", func(t *testing.T) {
		var buf bytes.Buffer
		logger := NewStructuredLogger(LogLevelInfo, &buf, "test")

		duration := 50 * time.Millisecond
		logger.LogOperation("database_query", duration, nil)

		var entry LogEntry
		if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
			t.Fatalf("Failed to unmarshal log entry: %v", err)
		}

		if entry.Level != "INFO" {
			t.Errorf("Expected level INFO, got %s", entry.Level)
		}
		if entry.Operation != "database_query" {
			t.Errorf("Expected operation 'database_query', got '%s'", entry.Operation)
		}
		if entry.Duration != duration.String() {
			t.Errorf("Expected duration '%s', got '%s'", duration.String(), entry.Duration)
		}
		if entry.Error != "" {
			t.Error("Expected no error for successful operation")
		}
	})

	t.Run("FailedOperation", func(t *testing.T) {
		var buf bytes.Buffer
		logger := NewStructuredLogger(LogLevelInfo, &buf, "test")

		duration := 25 * time.Millisecond
		testErr := &testError{message: "operation failed"}
		logger.LogOperation("database_query", duration, testErr)

		var entry LogEntry
		if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
			t.Fatalf("Failed to unmarshal log entry: %v", err)
		}

		if entry.Level != "ERROR" {
			t.Errorf("Expected level ERROR, got %s", entry.Level)
		}
		if entry.Error != "operation failed" {
			t.Errorf("Expected error 'operation failed', got '%s'", entry.Error)
		}
	})
}

// TestLoggingMiddleware tests the logging middleware
func TestLoggingMiddleware(t *testing.T) {
	t.Run("MiddlewareLogging", func(t *testing.T) {
		var buf bytes.Buffer
		logger := NewStructuredLogger(LogLevelInfo, &buf, "middleware")

		// Create test handler
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get logger from context
			ctxLogger := GetLoggerFromContext(r.Context())
			ctxLogger.Info("Handler executed")

			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		})

		// Wrap with logging middleware
		wrappedHandler := LoggingMiddleware(logger)(handler)

		// Create test request
		req := httptest.NewRequest("POST", "/api/v2/test", nil)
		req.Header.Set("User-Agent", "Test-Client/1.0")
		w := httptest.NewRecorder()

		// Execute request
		wrappedHandler.ServeHTTP(w, req)

		// Check that logs were written
		logOutput := buf.String()
		if logOutput == "" {
			t.Error("Expected log output from middleware")
		}

		// Should contain multiple log entries
		lines := strings.Split(strings.TrimSpace(logOutput), "\n")
		if len(lines) < 2 {
			t.Errorf("Expected at least 2 log entries, got %d", len(lines))
		}

		// Parse first log entry (request start)
		var startEntry LogEntry
		if err := json.Unmarshal([]byte(lines[0]), &startEntry); err != nil {
			t.Fatalf("Failed to unmarshal start log entry: %v", err)
		}

		if startEntry.Message != "HTTP request started" {
			t.Errorf("Expected start message, got '%s'", startEntry.Message)
		}
		if startEntry.RequestID == "" {
			t.Error("Expected request ID in start log")
		}

		// Parse last log entry (request completion)
		var endEntry LogEntry
		if err := json.Unmarshal([]byte(lines[len(lines)-1]), &endEntry); err != nil {
			t.Fatalf("Failed to unmarshal end log entry: %v", err)
		}

		if endEntry.Message != "HTTP request processed" {
			t.Errorf("Expected processed message, got '%s'", endEntry.Message)
		}
		if endEntry.StatusCode != 200 {
			t.Errorf("Expected status code 200, got %d", endEntry.StatusCode)
		}
		if endEntry.Method != "POST" {
			t.Errorf("Expected method POST, got %s", endEntry.Method)
		}
	})

	t.Run("GetLoggerFromContext", func(t *testing.T) {
		// Test with context that has logger
		var buf bytes.Buffer
		logger := NewStructuredLogger(LogLevelInfo, &buf, "test")
		ctx := context.WithValue(context.Background(), loggerKey, logger)

		retrievedLogger := GetLoggerFromContext(ctx)
		if retrievedLogger != logger {
			t.Error("Expected to retrieve the same logger from context")
		}

		// Test with context that doesn't have logger
		emptyCtx := context.Background()
		defaultLogger := GetLoggerFromContext(emptyCtx)
		if defaultLogger == nil {
			t.Error("Expected default logger when none in context")
		}
	})
}

// testError is a simple error implementation for testing
type testError struct {
	message string
}

func (e *testError) Error() string {
	return e.message
}
