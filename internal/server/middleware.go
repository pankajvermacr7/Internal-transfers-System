// Package server provides HTTP server implementation and middleware.
package server

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

// contextKey is a custom type for context keys to avoid collisions.
type contextKey string

const (
	// RequestIDKey is the context key for the request ID.
	RequestIDKey contextKey = "request_id"

	// RequestIDHeader is the HTTP header name for request ID.
	RequestIDHeader = "X-Request-ID"
)

// RequestIDMiddleware adds a unique request ID to each request.
// If the client provides a request ID via the X-Request-ID header, it is used.
// Otherwise, a new random ID is generated.
// The request ID is added to:
//   - The request context (for logging and tracing)
//   - The response header (for client correlation)
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check for existing request ID from client
		requestID := r.Header.Get(RequestIDHeader)
		if requestID == "" {
			requestID = generateRequestID()
		}

		// Add to response headers
		w.Header().Set(RequestIDHeader, requestID)

		// Add to request context
		ctx := context.WithValue(r.Context(), RequestIDKey, requestID)

		// Continue with the updated context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// generateRequestID generates a random 16-character hex string for request tracing.
// This provides sufficient uniqueness for request correlation without external dependencies.
func generateRequestID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		// Fallback to timestamp-based ID if random fails
		return "req-fallback"
	}
	return hex.EncodeToString(b)
}

// LoggingMiddleware logs HTTP requests with timing and status information.
// It captures:
//   - Request method, path, and remote address
//   - Response status code and size
//   - Request duration
//   - Request ID (if present)
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap response writer to capture status code
		wrapped := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// Process request
		next.ServeHTTP(wrapped, r)

		// Calculate duration
		duration := time.Since(start)

		// Get request ID from context
		requestID, _ := r.Context().Value(RequestIDKey).(string)

		// Log the request
		log.Info().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Str("remote_addr", r.RemoteAddr).
			Int("status", wrapped.statusCode).
			Int64("size", wrapped.bytesWritten).
			Dur("duration", duration).
			Str("request_id", requestID).
			Msg("HTTP request")
	})
}

// RecoveryMiddleware recovers from panics and returns a 500 error.
// It logs the panic with stack trace for debugging.
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				requestID, _ := r.Context().Value(RequestIDKey).(string)

				log.Error().
					Interface("panic", err).
					Str("path", r.URL.Path).
					Str("method", r.Method).
					Str("request_id", requestID).
					Msg("Panic recovered in HTTP handler")

				http.Error(w, `{"success":false,"error":"internal_error","message":"An unexpected error occurred"}`, http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// responseWriter wraps http.ResponseWriter to capture response metadata.
type responseWriter struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int64
}

// WriteHeader captures the status code before writing it.
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Write captures the number of bytes written.
func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.bytesWritten += int64(n)
	return n, err
}

// GetRequestID extracts the request ID from the context.
// Returns an empty string if no request ID is present.
func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(RequestIDKey).(string); ok {
		return id
	}
	return ""
}
