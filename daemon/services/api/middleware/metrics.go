package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// HTTP request metrics
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "uma_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status_code"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "uma_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint", "status_code"},
	)

	httpRequestsInFlight = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "uma_http_requests_in_flight",
			Help: "Number of HTTP requests currently being processed",
		},
	)

	httpRequestSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "uma_http_request_size_bytes",
			Help:    "HTTP request size in bytes",
			Buckets: prometheus.ExponentialBuckets(100, 10, 8),
		},
		[]string{"method", "endpoint"},
	)

	httpResponseSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "uma_http_response_size_bytes",
			Help:    "HTTP response size in bytes",
			Buckets: prometheus.ExponentialBuckets(100, 10, 8),
		},
		[]string{"method", "endpoint", "status_code"},
	)
)

// Metrics returns a middleware that collects HTTP metrics
func Metrics() func(http.Handler) http.Handler {
	return MetricsWithConfig(DefaultMetricsConfig())
}

// MetricsWithConfig returns a metrics middleware with custom configuration
func MetricsWithConfig(config MetricsConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip metrics collection for certain paths if configured
			if shouldSkipMetrics(r.URL.Path, config.SkipPaths) {
				next.ServeHTTP(w, r)
				return
			}

			start := time.Now()
			httpRequestsInFlight.Inc()
			defer httpRequestsInFlight.Dec()

			// Normalize endpoint for metrics (remove IDs, etc.)
			endpoint := normalizeEndpoint(r.URL.Path, config.EndpointNormalization)

			// Record request size
			if r.ContentLength > 0 {
				httpRequestSize.WithLabelValues(r.Method, endpoint).Observe(float64(r.ContentLength))
			}

			// Create a response writer wrapper to capture status code and response size
			wrapper := &metricsResponseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
				responseSize:   0,
			}

			next.ServeHTTP(wrapper, r)

			duration := time.Since(start)
			statusCode := strconv.Itoa(wrapper.statusCode)

			// Record metrics
			httpRequestsTotal.WithLabelValues(r.Method, endpoint, statusCode).Inc()
			httpRequestDuration.WithLabelValues(r.Method, endpoint, statusCode).Observe(duration.Seconds())

			if wrapper.responseSize > 0 {
				httpResponseSize.WithLabelValues(r.Method, endpoint, statusCode).Observe(float64(wrapper.responseSize))
			}
		})
	}
}

// MetricsConfig represents metrics middleware configuration
type MetricsConfig struct {
	SkipPaths             []string          `json:"skip_paths"`
	EndpointNormalization map[string]string `json:"endpoint_normalization"`
	CustomLabels          map[string]string `json:"custom_labels"`
}

// DefaultMetricsConfig returns a default metrics configuration
func DefaultMetricsConfig() MetricsConfig {
	return MetricsConfig{
		SkipPaths: []string{
			"/metrics",
			"/favicon.ico",
		},
		EndpointNormalization: map[string]string{
			// Normalize endpoints with IDs
			"/api/v1/docker/containers/": "/api/v1/docker/containers/{id}",
			"/api/v1/vms/":               "/api/v1/vms/{name}",
			"/api/v1/operations/":        "/api/v1/operations/{id}",
		},
		CustomLabels: map[string]string{
			"service": "uma",
		},
	}
}

// shouldSkipMetrics determines if metrics collection should be skipped for a path
func shouldSkipMetrics(path string, skipPaths []string) bool {
	for _, skipPath := range skipPaths {
		if path == skipPath {
			return true
		}
	}
	return false
}

// normalizeEndpoint normalizes endpoint paths for consistent metrics
func normalizeEndpoint(path string, normalization map[string]string) string {
	for pattern, normalized := range normalization {
		if len(path) > len(pattern) && path[:len(pattern)] == pattern {
			return normalized
		}
	}
	return path
}

// metricsResponseWriter wraps http.ResponseWriter to capture metrics
type metricsResponseWriter struct {
	http.ResponseWriter
	statusCode   int
	responseSize int
}

func (mrw *metricsResponseWriter) WriteHeader(code int) {
	mrw.statusCode = code
	mrw.ResponseWriter.WriteHeader(code)
}

func (mrw *metricsResponseWriter) Write(b []byte) (int, error) {
	size, err := mrw.ResponseWriter.Write(b)
	mrw.responseSize += size
	return size, err
}

// RecordCustomMetric records a custom metric with labels
func RecordCustomMetric(name string, value float64, labels map[string]string) {
	// This would be implemented based on specific metric requirements
	// For now, it's a placeholder for custom metrics
}

// GetMetricsHandler returns the Prometheus metrics handler
func GetMetricsHandler() http.Handler {
	return promhttp.Handler()
}

// RegisterCustomMetrics allows registration of custom metrics
func RegisterCustomMetrics(collectors ...prometheus.Collector) error {
	for _, collector := range collectors {
		if err := prometheus.Register(collector); err != nil {
			return err
		}
	}
	return nil
}
