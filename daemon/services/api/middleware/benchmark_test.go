package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// BenchmarkRequestIDMiddleware tests request ID middleware performance
func BenchmarkRequestIDMiddleware(b *testing.B) {
	handler := RequestID()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}
}

// BenchmarkLoggingMiddleware tests logging middleware performance
func BenchmarkLoggingMiddleware(b *testing.B) {
	handler := Logging()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}
}

// BenchmarkCORSMiddleware tests CORS middleware performance
func BenchmarkCORSMiddleware(b *testing.B) {
	handler := CORS()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}
}

// BenchmarkCompressionMiddleware tests compression middleware performance
func BenchmarkCompressionMiddleware(b *testing.B) {
	largeResponse := strings.Repeat("This is test data for compression. ", 1000)

	handler := Compression()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(largeResponse))
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Accept-Encoding", "gzip")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}
}

// BenchmarkVersioningMiddleware tests versioning middleware performance
func BenchmarkVersioningMiddleware(b *testing.B) {
	handler := Versioning()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/v2/test", nil)
	req.Header.Set("Accept", "application/vnd.uma.v2+json")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}
}

// BenchmarkMetricsMiddleware tests metrics middleware performance
func BenchmarkMetricsMiddleware(b *testing.B) {
	handler := Metrics()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}
}

// BenchmarkMiddlewareChain tests complete middleware chain performance
func BenchmarkMiddlewareChain(b *testing.B) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "success"}`))
	})

	// Build complete middleware chain
	handler := RequestID()(
		Logging()(
			CORS()(
				Versioning()(
					Compression()(
						Metrics()(
							testHandler,
						),
					),
				),
			),
		),
	)

	req := httptest.NewRequest("GET", "/api/v2/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Accept-Encoding", "gzip")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}
}

// BenchmarkConcurrentMiddleware tests concurrent middleware performance
func BenchmarkConcurrentMiddleware(b *testing.B) {
	handler := RequestID()(
		Logging()(
			CORS()(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				}),
			),
		),
	)

	req := httptest.NewRequest("GET", "/test", nil)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
		}
	})
}

// BenchmarkCORSPreflight tests CORS preflight request performance
func BenchmarkCORSPreflight(b *testing.B) {
	handler := CORS()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("OPTIONS", "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "Content-Type, Authorization")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}
}

// BenchmarkCompressionWithoutAcceptEncoding tests compression middleware without compression
func BenchmarkCompressionWithoutAcceptEncoding(b *testing.B) {
	largeResponse := strings.Repeat("This is test data. ", 1000)

	handler := Compression()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(largeResponse))
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	// No Accept-Encoding header

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}
}

// BenchmarkMiddlewareMemoryAllocation tests memory allocation in middleware
func BenchmarkMiddlewareMemoryAllocation(b *testing.B) {
	handler := RequestID()(
		Logging()(
			Metrics()(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{"status": "ok"}`))
				}),
			),
		),
	)

	req := httptest.NewRequest("GET", "/test", nil)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}
}

// BenchmarkLargeRequestMiddleware tests middleware with large requests
func BenchmarkLargeRequestMiddleware(b *testing.B) {
	largeBody := strings.Repeat(`{"data": "test"}`, 10000)

	handler := RequestID()(
		Logging()(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}),
		),
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/test", strings.NewReader(largeBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}
}

// BenchmarkMiddlewareErrorHandling tests middleware performance with error responses
func BenchmarkMiddlewareErrorHandling(b *testing.B) {
	handler := RequestID()(
		Logging()(
			Metrics()(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(`{"error": "Internal server error"}`))
				}),
			),
		),
	)

	req := httptest.NewRequest("GET", "/test", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}
}
