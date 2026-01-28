package handler

import (
	"encoding/json"
	"net/http"

	"internal-transfers-system/internal/validator"

	"github.com/rs/zerolog/log"
)

type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
}

type ErrorResponse struct {
	Success   bool   `json:"success"`
	Error     string `json:"error"`
	Message   string `json:"message"`
	RequestID string `json:"request_id,omitempty"`
}

type ValidationErrorResponse struct {
	Success bool                        `json:"success"`
	Error   string                      `json:"error"`
	Errors  []validator.ValidationError `json:"errors"`
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Error().Err(err).Msg("Failed to encode JSON response")
	}
}

func writeSuccess(w http.ResponseWriter, status int, data interface{}) {
	writeJSON(w, status, data)
}

func writeError(w http.ResponseWriter, status int, errorCode, message string) {
	requestID := w.Header().Get("X-Request-ID")

	writeJSON(w, status, ErrorResponse{
		Success:   false,
		Error:     errorCode,
		Message:   message,
		RequestID: requestID,
	})
}

func writeValidationError(w http.ResponseWriter, errs validator.ValidationErrors) {
	writeJSON(w, http.StatusBadRequest, ValidationErrorResponse{
		Success: false,
		Error:   "validation_failed",
		Errors:  errs,
	})
}

func writeInternalError(w http.ResponseWriter, err error) {
	log.Error().Err(err).Msg("Internal server error")
	writeError(w, http.StatusInternalServerError, "internal_error", "An unexpected error occurred. Please try again later.")
}
