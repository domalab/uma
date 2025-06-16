package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/domalab/uma/daemon/logger"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// API Request Metrics
	apiRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "uma_api_requests_total",
			Help: "Total number of API requests",
		},
		[]string{"method", "endpoint", "status_code"},
	)

	apiRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "uma_api_request_duration_seconds",
			Help:    "Duration of API requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	// Bulk Operation Metrics
	bulkOperationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "uma_bulk_operation_duration_seconds",
			Help:    "Duration of bulk operations in seconds",
			Buckets: []float64{0.1, 0.5, 1.0, 2.5, 5.0, 10.0},
		},
		[]string{"operation"},
	)

	bulkOperationContainers = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "uma_bulk_operation_containers",
			Help:    "Number of containers processed in bulk operations",
			Buckets: []float64{1, 5, 10, 25, 50},
		},
		[]string{"operation"},
	)

	bulkOperationSuccessRate = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "uma_bulk_operation_success_rate",
			Help: "Success rate of bulk operations (percentage)",
		},
		[]string{"operation"},
	)

	// WebSocket Connection Metrics
	websocketConnections = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "uma_websocket_connections",
			Help: "Number of active WebSocket connections",
		},
		[]string{"endpoint"},
	)

	websocketMessagesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "uma_websocket_messages_total",
			Help: "Total number of WebSocket messages sent",
		},
		[]string{"endpoint", "message_type"},
	)

	// Health Check Metrics
	healthCheckDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "uma_health_check_duration_seconds",
			Help:    "Duration of health checks in seconds",
			Buckets: []float64{0.1, 0.5, 1.0, 2.0, 5.0, 10.0},
		},
	)

	healthCheckStatus = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "uma_health_check_status",
			Help: "Health check status (1 = healthy, 0 = unhealthy)",
		},
		[]string{"dependency"},
	)

	// System Metrics
	configLoadTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "uma_config_load_total",
			Help: "Total number of configuration loads",
		},
		[]string{"config_type", "status"},
	)

	validationErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "uma_validation_errors_total",
			Help: "Total number of validation errors",
		},
		[]string{"component", "operation"},
	)
)

// Note: responseWriter is defined in http_server.go

// metricsMiddleware adds metrics collection to HTTP requests
func (h *HTTPServer) metricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap response writer to capture status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: 200}

		// Process request
		next.ServeHTTP(wrapped, r)

		// Record metrics
		duration := time.Since(start)
		method := r.Method
		endpoint := r.URL.Path
		statusCode := strconv.Itoa(wrapped.statusCode)

		// Update Prometheus metrics
		apiRequestsTotal.WithLabelValues(method, endpoint, statusCode).Inc()
		apiRequestDuration.WithLabelValues(method, endpoint).Observe(duration.Seconds())

		// Log structured request
		requestID := h.getRequestIDFromContext(r)
		logger.LogAPIRequest(requestID, method, endpoint, wrapped.statusCode, duration)

		// Log metrics collection
		logger.LogMetricsCollection("api_request", duration.Seconds(), map[string]string{
			"method":      method,
			"endpoint":    endpoint,
			"status_code": statusCode,
		})
	})
}

// RecordBulkOperation records metrics for bulk operations
func RecordBulkOperation(operation string, total, succeeded, failed int, duration time.Duration, requestID string) {
	// Calculate success rate
	successRate := float64(succeeded) / float64(total) * 100

	// Update Prometheus metrics
	bulkOperationDuration.WithLabelValues(operation).Observe(duration.Seconds())
	bulkOperationContainers.WithLabelValues(operation).Observe(float64(total))
	bulkOperationSuccessRate.WithLabelValues(operation).Set(successRate)

	// Log structured bulk operation
	logger.LogBulkOperation(operation, total, succeeded, failed, duration, requestID)

	// Log metrics collection
	logger.LogMetricsCollection("bulk_operation", duration.Seconds(), map[string]string{
		"operation":    operation,
		"total":        strconv.Itoa(total),
		"succeeded":    strconv.Itoa(succeeded),
		"failed":       strconv.Itoa(failed),
		"success_rate": strconv.FormatFloat(successRate, 'f', 2, 64),
	})
}

// RecordWebSocketConnection records WebSocket connection metrics
func RecordWebSocketConnection(endpoint, event string, connectionCount int) {
	switch event {
	case "connect":
		websocketConnections.WithLabelValues(endpoint).Inc()
	case "disconnect":
		websocketConnections.WithLabelValues(endpoint).Dec()
	}

	// Generate client ID for logging (simplified)
	clientID := "client_" + strconv.Itoa(int(time.Now().Unix()))

	// Log structured WebSocket event
	logger.LogWebSocketConnection(endpoint, event, clientID, connectionCount)

	// Log metrics collection
	logger.LogMetricsCollection("websocket_connection", float64(connectionCount), map[string]string{
		"endpoint": endpoint,
		"event":    event,
	})
}

// RecordWebSocketMessage records WebSocket message metrics
func RecordWebSocketMessage(endpoint, messageType string) {
	websocketMessagesTotal.WithLabelValues(endpoint, messageType).Inc()

	// Log metrics collection
	logger.LogMetricsCollection("websocket_message", 1, map[string]string{
		"endpoint":     endpoint,
		"message_type": messageType,
	})
}

// RecordHealthCheck records health check metrics
func RecordHealthCheck(status string, dependencies map[string]string, duration time.Duration, requestID string) {
	// Record overall duration
	healthCheckDuration.Observe(duration.Seconds())

	// Record dependency statuses
	for dep, depStatus := range dependencies {
		if depStatus == "healthy" {
			healthCheckStatus.WithLabelValues(dep).Set(1)
		} else {
			healthCheckStatus.WithLabelValues(dep).Set(0)
		}
	}

	// Log structured health check
	logger.LogHealthCheck(status, dependencies, duration, requestID)

	// Log metrics collection
	logger.LogMetricsCollection("health_check", duration.Seconds(), map[string]string{
		"status": status,
	})
}

// RecordConfigLoad records configuration loading metrics
func RecordConfigLoad(configType, path string, success bool, errorMsg string) {
	status := "success"
	if !success {
		status = "error"
	}

	configLoadTotal.WithLabelValues(configType, status).Inc()

	// Log structured config load
	logger.LogConfigLoad(configType, path, success, errorMsg)

	// Log metrics collection
	logger.LogMetricsCollection("config_load", 1, map[string]string{
		"config_type": configType,
		"status":      status,
	})
}

// RecordValidationError records validation error metrics
func RecordValidationError(component, operation, requestID, errorMsg string) {
	validationErrorsTotal.WithLabelValues(component, operation).Inc()

	// Log structured validation error
	logger.LogValidationError(component, operation, requestID, errorMsg)

	// Log metrics collection
	logger.LogMetricsCollection("validation_error", 1, map[string]string{
		"component": component,
		"operation": operation,
	})
}

// handleMetrics serves Prometheus metrics
func (h *HTTPServer) handleMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Serve Prometheus metrics
	promhttp.Handler().ServeHTTP(w, r)
}
