package server

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

// ServiceName is the canonical name of this service.
const ServiceName = "internal-transfers-system"

// ServiceVersion is the current version of this service.
// TODO: This should be set at build time via ldflags.
const ServiceVersion = "1.0.0"

// HealthResponse represents the health check response.
// This response indicates whether the service process is running.
type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Service   string    `json:"service"`
	Version   string    `json:"version,omitempty"`
}

// ReadyResponse represents the readiness check response.
// This response indicates whether the service is ready to accept traffic,
// including the status of all dependent services (database, etc.).
type ReadyResponse struct {
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Checks    map[string]string `json:"checks,omitempty"`
}

// handleHealth returns the health status of the service.
//
// This endpoint is used by container orchestrators (Docker, Kubernetes) for:
//   - Liveness probes: Determines if the container should be restarted
//   - Load balancer health checks: Determines if traffic should be routed
//
// The health check is lightweight and does not verify external dependencies.
// Use /ready for full readiness checks including database connectivity.
//
// Response: 200 OK with HealthResponse body
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().UTC(),
		Service:   ServiceName,
		Version:   ServiceVersion,
	}

	writeServerJSON(w, http.StatusOK, response)
}

// handleReady returns the readiness status of the service.
//
// This endpoint is used by container orchestrators for:
//   - Readiness probes: Determines if the pod should receive traffic
//   - Deployment strategies: Determines when new pods are ready
//
// The readiness check verifies:
//   - Database connectivity and responsiveness
//
// Responses:
//   - 200 OK: Service is ready to accept traffic
//   - 503 Service Unavailable: Service is not ready (dependencies failing)
func (s *Server) handleReady(w http.ResponseWriter, r *http.Request) {
	checks := make(map[string]string)
	statusCode := http.StatusOK
	readyStatus := "ready"

	// Check database connectivity with timeout
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	if s.db == nil {
		checks["database"] = "uninitialized"
		statusCode = http.StatusServiceUnavailable
		readyStatus = "not_ready"
		log.Warn().Msg("Readiness check failed: database not initialized")
	} else if err := s.db.Ping(ctx); err != nil {
		checks["database"] = "unavailable"
		statusCode = http.StatusServiceUnavailable
		readyStatus = "not_ready"
		log.Warn().Err(err).Msg("Readiness check failed: database ping error")
	} else {
		checks["database"] = "ok"
	}

	response := ReadyResponse{
		Status:    readyStatus,
		Timestamp: time.Now().UTC(),
		Checks:    checks,
	}

	writeServerJSON(w, statusCode, response)
}

// writeServerJSON writes a JSON response with the given status code.
// This is a server-specific helper that doesn't depend on the handler package.
func writeServerJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Error().Err(err).Msg("Failed to encode JSON response")
	}
}
