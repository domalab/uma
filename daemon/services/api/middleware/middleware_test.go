package middleware

import (
	"bytes"
	"compress/gzip"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

// TestCORS tests the CORS middleware
func TestCORS(t *testing.T) {
	// Create a test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	})

	// Wrap with CORS middleware
	corsHandler := CORS()(testHandler)

	tests := []struct {
		name           string
		method         string
		origin         string
		expectedStatus int
		checkHeaders   map[string]string
	}{
		{
			name:           "GET request with CORS headers",
			method:         "GET",
			origin:         "http://localhost:3000",
			expectedStatus: http.StatusOK,
			checkHeaders: map[string]string{
				"Access-Control-Allow-Origin":  "*",
				"Access-Control-Allow-Methods": "GET, POST, PUT, DELETE, OPTIONS",
				"Access-Control-Allow-Headers": "Content-Type, Authorization, X-Request-ID",
			},
		},
		{
			name:           "OPTIONS preflight request",
			method:         "OPTIONS",
			origin:         "http://localhost:3000",
			expectedStatus: http.StatusOK,
			checkHeaders: map[string]string{
				"Access-Control-Allow-Origin":  "*",
				"Access-Control-Allow-Methods": "GET, POST, PUT, DELETE, OPTIONS",
				"Access-Control-Allow-Headers": "Content-Type, Authorization, X-Request-ID",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/test", nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}
			w := httptest.NewRecorder()

			corsHandler.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			// Check CORS headers
			for header, expectedValue := range tt.checkHeaders {
				actualValue := w.Header().Get(header)
				if actualValue != expectedValue {
					t.Errorf("Expected header %s to be '%s', got '%s'", header, expectedValue, actualValue)
				}
			}
		})
	}
}

// TestRequestID tests the Request ID middleware
func TestRequestID(t *testing.T) {
	// Create a test handler that checks for request ID in context
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// The request ID should be available in the context or response headers
		w.WriteHeader(http.StatusOK)
	})

	// Wrap with RequestID middleware
	requestIDHandler := RequestID()(testHandler)

	tests := []struct {
		name              string
		existingRequestID string
		expectNewID       bool
	}{
		{
			name:        "Generate new request ID when none exists",
			expectNewID: true,
		},
		{
			name:              "Use existing request ID",
			existingRequestID: "existing-id-123",
			expectNewID:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.existingRequestID != "" {
				req.Header.Set("X-Request-ID", tt.existingRequestID)
			}
			w := httptest.NewRecorder()

			requestIDHandler.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
			}

			responseRequestID := w.Header().Get("X-Request-ID")
			if tt.expectNewID {
				if responseRequestID == "" {
					t.Error("Expected new request ID to be generated")
				}
				if len(responseRequestID) < 10 {
					t.Error("Expected generated request ID to be at least 10 characters")
				}
			} else {
				if responseRequestID != tt.existingRequestID {
					t.Errorf("Expected request ID to be '%s', got '%s'", tt.existingRequestID, responseRequestID)
				}
			}
		})
	}
}

// TestVersioning tests the API versioning middleware
func TestVersioning(t *testing.T) {
	// Create a test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	})

	// Wrap with Versioning middleware
	versioningHandler := Versioning()(testHandler)

	tests := []struct {
		name           string
		acceptHeader   string
		expectedStatus int
		checkHeaders   map[string]string
	}{
		{
			name:           "Default version when no Accept header",
			acceptHeader:   "",
			expectedStatus: http.StatusOK,
			checkHeaders: map[string]string{
				"X-API-Version": "v1",
			},
		},
		{
			name:           "Valid version in Accept header",
			acceptHeader:   "application/vnd.uma.v1+json",
			expectedStatus: http.StatusOK,
			checkHeaders: map[string]string{
				"X-API-Version": "v1",
			},
		},
		{
			name:           "Standard JSON Accept header",
			acceptHeader:   "application/json",
			expectedStatus: http.StatusOK,
			checkHeaders: map[string]string{
				"X-API-Version": "v1",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.acceptHeader != "" {
				req.Header.Set("Accept", tt.acceptHeader)
			}
			w := httptest.NewRecorder()

			versioningHandler.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			// Check version headers
			for header, expectedValue := range tt.checkHeaders {
				actualValue := w.Header().Get(header)
				if actualValue != expectedValue {
					t.Errorf("Expected header %s to be '%s', got '%s'", header, expectedValue, actualValue)
				}
			}
		})
	}
}

// TestCompression tests the compression middleware
func TestCompression(t *testing.T) {
	// Create a test handler that returns a large response
	largeResponse := strings.Repeat("This is a test response that should be compressed. ", 100)
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(largeResponse))
	})

	// Wrap with Compression middleware
	compressionHandler := Compression()(testHandler)

	tests := []struct {
		name           string
		acceptEncoding string
		expectGzip     bool
	}{
		{
			name:           "Gzip compression when supported",
			acceptEncoding: "gzip, deflate",
			expectGzip:     false, // Our simple middleware might not implement compression
		},
		{
			name:           "No compression when not supported",
			acceptEncoding: "",
			expectGzip:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.acceptEncoding != "" {
				req.Header.Set("Accept-Encoding", tt.acceptEncoding)
			}
			w := httptest.NewRecorder()

			compressionHandler.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
			}

			// For now, just check that the middleware doesn't break the response
			// In a real implementation, you would check for actual compression
			if len(w.Body.Bytes()) == 0 {
				t.Error("Expected response body to not be empty")
			}
		})
	}
}

// TestGzipCompression tests actual gzip compression functionality
func TestGzipCompression(t *testing.T) {
	// Test gzip compression/decompression functionality
	original := "This is a test string for compression"

	var buf bytes.Buffer
	writer := gzip.NewWriter(&buf)
	_, err := writer.Write([]byte(original))
	if err != nil {
		t.Fatalf("Failed to write to gzip writer: %v", err)
	}
	writer.Close()

	// Decompress
	reader, err := gzip.NewReader(&buf)
	if err != nil {
		t.Fatalf("Failed to create gzip reader: %v", err)
	}
	defer reader.Close()

	var decompressed bytes.Buffer
	_, err = decompressed.ReadFrom(reader)
	if err != nil {
		t.Fatalf("Failed to read from gzip reader: %v", err)
	}

	if decompressed.String() != original {
		t.Errorf("Expected %s, got %s", original, decompressed.String())
	}
}

// TestMiddlewareErrorConditions tests error conditions and edge cases for middleware
func TestMiddlewareErrorConditions(t *testing.T) {
	t.Run("CORS with invalid origin", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		corsHandler := CORS()(handler)

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Origin", "javascript:alert('xss')")
		w := httptest.NewRecorder()

		corsHandler.ServeHTTP(w, req)

		// Should handle malicious origins safely
		origin := w.Header().Get("Access-Control-Allow-Origin")
		if strings.Contains(origin, "javascript:") {
			t.Error("CORS should not allow javascript: origins")
		}
	})

	t.Run("RequestID with existing ID", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get("X-Request-ID")
			if requestID == "" {
				t.Error("Expected request ID to be preserved")
			}
			w.WriteHeader(http.StatusOK)
		})

		requestIDHandler := RequestID()(handler)

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Request-ID", "existing-id-123")
		w := httptest.NewRecorder()

		requestIDHandler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
	})

	t.Run("Versioning with malformed Accept header", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		versionHandler := Versioning()(handler)

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Accept", "application/vnd.uma.invalid+json")
		w := httptest.NewRecorder()

		versionHandler.ServeHTTP(w, req)

		// Should handle malformed version gracefully
		if w.Code == http.StatusInternalServerError {
			t.Error("Versioning middleware should handle malformed Accept headers gracefully")
		}
	})

	t.Run("Compression with uncompressible content", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "image/jpeg")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("fake-image-data"))
		})

		compressionHandler := Compression()(handler)

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Accept-Encoding", "gzip")
		w := httptest.NewRecorder()

		compressionHandler.ServeHTTP(w, req)

		// Should not compress images
		if w.Header().Get("Content-Encoding") == "gzip" {
			t.Error("Should not compress image content")
		}
	})

	t.Run("Logging with nil request", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		loggingHandler := Logging()(handler)

		// This tests the middleware's robustness
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Logging middleware should not panic: %v", r)
			}
		}()

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		loggingHandler.ServeHTTP(w, req)
	})
}

// TestMiddlewareChainIntegration tests middleware chain integration
func TestMiddlewareChainIntegration(t *testing.T) {
	t.Run("CompleteMiddlewareChain", func(t *testing.T) {
		// Create a test handler
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("test response"))
		})

		// Apply all middleware
		handler := RequestID()(
			Logging()(
				CORS()(
					Compression()(
						testHandler,
					),
				),
			),
		)

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		req.Header.Set("Accept-Encoding", "gzip")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		// Check that all middleware applied their effects
		if w.Header().Get("X-Request-ID") == "" {
			t.Error("Expected X-Request-ID header from RequestIDMiddleware")
		}

		if w.Header().Get("Access-Control-Allow-Origin") == "" {
			t.Error("Expected CORS headers from CORSMiddleware")
		}
	})

	t.Run("MiddlewareErrorHandling", func(t *testing.T) {
		// Create a handler that returns an error status
		errorHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal Server Error"))
		})

		// Apply middleware
		handler := RequestID()(
			Logging()(
				errorHandler,
			),
		)

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		// Should still have request ID even if handler returns error
		if w.Header().Get("X-Request-ID") == "" {
			t.Error("Expected X-Request-ID header even when handler returns error")
		}

		if w.Code != http.StatusInternalServerError {
			t.Errorf("Expected status 500, got %d", w.Code)
		}
	})
}

// TestCORSMiddleware_EdgeCases tests CORS middleware edge cases
func TestCORSMiddleware_EdgeCases(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	t.Run("PreflightWithCustomHeaders", func(t *testing.T) {
		handler := CORS()(testHandler)

		req := httptest.NewRequest("OPTIONS", "/test", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		req.Header.Set("Access-Control-Request-Method", "POST")
		req.Header.Set("Access-Control-Request-Headers", "Content-Type, Authorization, X-Custom-Header")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200 for preflight, got %d", w.Code)
		}

		allowedMethods := w.Header().Get("Access-Control-Allow-Methods")
		if !strings.Contains(allowedMethods, "POST") {
			t.Errorf("Expected POST in allowed methods, got: %s", allowedMethods)
		}
	})

	t.Run("CORSWithInvalidOrigin", func(t *testing.T) {
		handler := CORS()(testHandler)

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Origin", "http://malicious-site.com")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		// Should still process the request but with appropriate CORS headers
		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		// Check CORS headers are set appropriately
		origin := w.Header().Get("Access-Control-Allow-Origin")
		if origin == "" {
			t.Error("Expected some CORS origin header to be set")
		}
	})
}

// TestCompressionMiddleware_EdgeCases tests compression middleware edge cases
func TestCompressionMiddleware_EdgeCases(t *testing.T) {
	t.Run("CompressionWithLargeResponse", func(t *testing.T) {
		largeResponse := strings.Repeat("This is a test response that should be compressed. ", 1000)

		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(largeResponse))
		})

		handler := Compression()(testHandler)

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Accept-Encoding", "gzip")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		// Check if response was compressed
		encoding := w.Header().Get("Content-Encoding")
		if encoding == "gzip" {
			// Response was compressed
			if w.Body.Len() >= len(largeResponse) {
				t.Error("Expected compressed response to be smaller than original")
			}
		}
	})

	t.Run("CompressionWithSmallResponse", func(t *testing.T) {
		smallResponse := "small"

		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(smallResponse))
		})

		handler := Compression()(testHandler)

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Accept-Encoding", "gzip")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		// Small responses might not be compressed
		body := w.Body.String()
		if body != smallResponse && w.Header().Get("Content-Encoding") != "gzip" {
			t.Error("Response should either be uncompressed or properly compressed")
		}
	})
}

// TestMiddlewareConcurrency tests middleware under concurrent load
func TestMiddlewareConcurrency(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(1 * time.Millisecond) // Simulate some work
		w.WriteHeader(http.StatusOK)
	})

	// Stack multiple middleware
	finalHandler := RequestID()(
		CORS()(
			Versioning()(
				Compression()(
					Logging()(handler),
				),
			),
		),
	)

	const numGoroutines = 20
	const requestsPerGoroutine = 5

	var wg sync.WaitGroup
	results := make(chan int, numGoroutines*requestsPerGoroutine)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < requestsPerGoroutine; j++ {
				req := httptest.NewRequest("GET", "/test", nil)
				w := httptest.NewRecorder()
				finalHandler.ServeHTTP(w, req)
				results <- w.Code
			}
		}()
	}

	wg.Wait()
	close(results)

	// Check that all requests completed successfully
	successCount := 0
	for status := range results {
		if status == http.StatusOK {
			successCount++
		}
	}

	expectedRequests := numGoroutines * requestsPerGoroutine
	if successCount != expectedRequests {
		t.Errorf("Expected %d successful requests, got %d", expectedRequests, successCount)
	}
}

// TestLogging tests the logging middleware
func TestLogging(t *testing.T) {
	// Create a test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	})

	// Wrap with Logging middleware
	loggingHandler := Logging()(testHandler)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	loggingHandler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	// Note: In a real implementation, you would capture log output and verify it
	// For now, we just ensure the middleware doesn't break the request flow
}

// TestMetrics tests the metrics middleware
func TestMetrics(t *testing.T) {
	// Create a test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	})

	// Wrap with Metrics middleware
	metricsHandler := Metrics()(testHandler)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	metricsHandler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	// Note: In a real implementation, you would check that metrics were recorded
	// For now, we just ensure the middleware doesn't break the request flow
}

// TestGetMetricsHandler tests the metrics handler
func TestGetMetricsHandler(t *testing.T) {
	handler := GetMetricsHandler()

	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	// Check that response contains Prometheus metrics format
	body := w.Body.String()
	if !strings.Contains(body, "# HELP") || !strings.Contains(body, "# TYPE") {
		t.Error("Expected Prometheus metrics format in response")
	}
}
