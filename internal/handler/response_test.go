package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"internal-transfers-system/internal/validator"
)

func TestWriteJSON(t *testing.T) {
	rec := httptest.NewRecorder()
	writeJSON(rec, http.StatusOK, map[string]string{"key": "value"})

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json; charset=utf-8" {
		t.Errorf("wrong content-type: %s", ct)
	}
}

func TestWriteError(t *testing.T) {
	rec := httptest.NewRecorder()
	writeError(rec, http.StatusBadRequest, "test_error", "test message")

	var resp ErrorResponse
	json.Unmarshal(rec.Body.Bytes(), &resp)

	if resp.Success {
		t.Error("expected success=false")
	}
	if resp.Error != "test_error" {
		t.Errorf("expected test_error, got %s", resp.Error)
	}
}

func TestWriteValidationError(t *testing.T) {
	errs := validator.ValidationErrors{
		{Field: "field1", Message: "required"},
	}

	rec := httptest.NewRecorder()
	writeValidationError(rec, errs)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}

	var resp ValidationErrorResponse
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if len(resp.Errors) != 1 {
		t.Errorf("expected 1 error, got %d", len(resp.Errors))
	}
}
