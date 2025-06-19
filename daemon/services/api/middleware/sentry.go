package middleware

import (
	"fmt"
	"net/http"

	"github.com/getsentry/sentry-go"
)

// Sentry returns a middleware that adds Sentry error tracking to HTTP requests
func Sentry() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Create a new Sentry hub for this request
			hub := sentry.GetHubFromContext(r.Context())
			if hub == nil {
				hub = sentry.CurrentHub().Clone()
			}

			// Add request context to Sentry
			hub.Scope().SetTag("http.method", r.Method)
			hub.Scope().SetTag("http.url", r.URL.Path)
			hub.Scope().SetTag("http.user_agent", r.UserAgent())
			hub.Scope().SetTag("http.remote_addr", r.RemoteAddr)

			// Add request ID if available
			if requestID := r.Header.Get("X-Request-ID"); requestID != "" {
				hub.Scope().SetTag("request_id", requestID)
			}

			// Create new context with Sentry hub
			ctx := sentry.SetHubOnContext(r.Context(), hub)
			r = r.WithContext(ctx)

			// Wrap response writer to capture status codes
			wrapped := &sentryResponseWriter{ResponseWriter: w, statusCode: 200}

			// Recover from panics and send to Sentry
			defer func() {
				if err := recover(); err != nil {
					hub.RecoverWithContext(ctx, err)
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			}()

			// Call next handler
			next.ServeHTTP(wrapped, r)

			// Capture HTTP errors (4xx/5xx) in Sentry
			if wrapped.statusCode >= 400 {
				hub.Scope().SetLevel(sentry.LevelError)
				hub.CaptureMessage(fmt.Sprintf("HTTP %d: %s %s", wrapped.statusCode, r.Method, r.URL.Path))
			}
		})
	}
}

// sentryResponseWriter wraps http.ResponseWriter to capture status codes
type sentryResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *sentryResponseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}
