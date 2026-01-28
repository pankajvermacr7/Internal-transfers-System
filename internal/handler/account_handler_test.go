package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"internal-transfers-system/internal/models"
)

func TestDecodeJSONBody(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		wantErr bool
	}{
		{"valid", `{"account_id": 1, "initial_balance": "100.00"}`, false},
		{"empty", "", true},
		{"malformed", `{invalid}`, true},
		{"unknown fields", `{"account_id": 1, "extra": "field"}`, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(tt.body))
			var target models.CreateAccountRequest
			err := decodeJSONBody(req, &target)
			if (err != nil) != tt.wantErr {
				t.Errorf("wantErr=%v, got err=%v", tt.wantErr, err)
			}
		})
	}
}

func TestHandleServiceError(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   string
	}{
		{"not found", models.ErrAccountNotFound, http.StatusNotFound, "account_not_found"},
		{"already exists", models.ErrAccountAlreadyExists, http.StatusConflict, "account_exists"},
		{"insufficient", models.ErrInsufficientBalance, http.StatusUnprocessableEntity, "insufficient_balance"},
		{"invalid amount", models.ErrInvalidAmount, http.StatusBadRequest, "invalid_amount"},
		{"same account", models.ErrSameAccount, http.StatusBadRequest, "same_account"},
		{"timeout", context.DeadlineExceeded, http.StatusGatewayTimeout, "timeout"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			handleServiceError(context.Background(), rec, tt.err)

			if rec.Code != tt.wantStatus {
				t.Errorf("expected %d, got %d", tt.wantStatus, rec.Code)
			}

			var resp ErrorResponse
			json.Unmarshal(rec.Body.Bytes(), &resp)
			if resp.Error != tt.wantCode {
				t.Errorf("expected %q, got %q", tt.wantCode, resp.Error)
			}
		})
	}
}

func TestMapDomainError(t *testing.T) {
	tests := []struct {
		code       models.ErrorCode
		wantStatus int
	}{
		{models.CodeAccountNotFound, http.StatusNotFound},
		{models.CodeAccountAlreadyExists, http.StatusConflict},
		{models.CodeInsufficientBalance, http.StatusUnprocessableEntity},
		{models.CodeInvalidAmount, http.StatusBadRequest},
		{models.CodeDatabaseError, http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(string(tt.code), func(t *testing.T) {
			err := &models.DomainError{Code: tt.code}
			status, _, _ := mapDomainError(err)
			if status != tt.wantStatus {
				t.Errorf("expected %d, got %d", tt.wantStatus, status)
			}
		})
	}
}
