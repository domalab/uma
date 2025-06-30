package middleware

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
)

// Compression returns a middleware that adds gzip compression for large responses
func Compression() func(http.Handler) http.Handler {
	return CompressionWithConfig(DefaultCompressionConfig())
}

// CompressionWithConfig returns a compression middleware with custom configuration
func CompressionWithConfig(config CompressionConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if client accepts gzip encoding
			if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
				next.ServeHTTP(w, r)
				return
			}

			// Only compress responses for certain endpoints or large responses
			shouldCompress := shouldCompressResponse(r, config)
			if !shouldCompress {
				next.ServeHTTP(w, r)
				return
			}

			// Create gzip writer
			gz := gzip.NewWriter(w)
			defer gz.Close()

			// Set compression headers
			w.Header().Set("Content-Encoding", "gzip")
			w.Header().Set("Vary", "Accept-Encoding")

			// Wrap response writer
			gzw := &gzipResponseWriter{
				ResponseWriter: w,
				Writer:         gz,
			}

			next.ServeHTTP(gzw, r)
		})
	}
}

// CompressionConfig represents compression middleware configuration
type CompressionConfig struct {
	Level             int      `json:"level"`              // Compression level (1-9)
	MinLength         int      `json:"min_length"`         // Minimum response length to compress
	CompressiblePaths []string `json:"compressible_paths"` // Paths that should be compressed
	CompressibleTypes []string `json:"compressible_types"` // Content types that should be compressed
	ExcludedPaths     []string `json:"excluded_paths"`     // Paths that should not be compressed
	ExcludedTypes     []string `json:"excluded_types"`     // Content types that should not be compressed
}

// DefaultCompressionConfig returns a default compression configuration
func DefaultCompressionConfig() CompressionConfig {
	return CompressionConfig{
		Level:     gzip.DefaultCompression,
		MinLength: 1024, // 1KB minimum
		CompressiblePaths: []string{
			"/api/v2/storage/config",
			"/api/v2/storage/layout",
			"/api/v2/containers/list",
			"/api/v2/vms/list",
			"/api/v2/system/info",
		},
		CompressibleTypes: []string{
			"application/json",
			"application/xml",
			"text/plain",
			"text/html",
			"text/css",
			"text/javascript",
			"application/javascript",
		},
		ExcludedPaths: []string{
			"/api/v2/stream", // WebSocket streaming endpoint
		},
		ExcludedTypes: []string{
			"image/",
			"video/",
			"audio/",
			"application/octet-stream",
		},
	}
}

// shouldCompressResponse determines if a response should be compressed
func shouldCompressResponse(r *http.Request, config CompressionConfig) bool {
	path := r.URL.Path

	// Check excluded paths first
	for _, excludedPath := range config.ExcludedPaths {
		if strings.HasPrefix(path, excludedPath) {
			return false
		}
	}

	// Check compressible paths
	for _, compressiblePath := range config.CompressiblePaths {
		if strings.HasPrefix(path, compressiblePath) {
			return true
		}
	}

	// Check content type if available in request
	contentType := r.Header.Get("Content-Type")
	if contentType != "" {
		// Check excluded types
		for _, excludedType := range config.ExcludedTypes {
			if strings.HasPrefix(contentType, excludedType) {
				return false
			}
		}

		// Check compressible types
		for _, compressibleType := range config.CompressibleTypes {
			if strings.HasPrefix(contentType, compressibleType) {
				return true
			}
		}
	}

	return false
}

// gzipResponseWriter wraps http.ResponseWriter to compress responses
type gzipResponseWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func (w *gzipResponseWriter) WriteHeader(code int) {
	w.ResponseWriter.WriteHeader(code)
}

// Flush implements http.Flusher interface
func (w *gzipResponseWriter) Flush() {
	if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

// Hijack implements http.Hijacker interface
func (w *gzipResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := w.ResponseWriter.(http.Hijacker); ok {
		return hijacker.Hijack()
	}
	return nil, nil, fmt.Errorf("response writer does not support hijacking")
}
