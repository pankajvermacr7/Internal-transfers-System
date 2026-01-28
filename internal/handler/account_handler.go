package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"internal-transfers-system/internal/models"
	"internal-transfers-system/internal/service"
	"internal-transfers-system/internal/validator"

	"github.com/rs/zerolog/log"
)

type AccountHandler struct {
	accountService *service.AccountService
}

func NewAccountHandler(accountService *service.AccountService) *AccountHandler {
	return &AccountHandler{accountService: accountService}
}

func (h *AccountHandler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req models.CreateAccountRequest
	if err := decodeJSONBody(r, &req); err != nil {
		log.Debug().Err(err).Msg("Failed to decode create account request")
		writeError(w, http.StatusBadRequest, "invalid_json", "Invalid JSON body")
		return
	}

	if errs := validator.ValidateCreateAccount(&req); len(errs) > 0 {
		log.Debug().Int64("accountID", req.AccountID).Interface("errors", errs).Msg("Create account validation failed")
		writeValidationError(w, errs)
		return
	}

	account, err := h.accountService.CreateAccount(ctx, &req)
	if err != nil {
		handleServiceError(ctx, w, err)
		return
	}

	resp := models.GetAccountResponse{
		AccountID: account.AccountID,
		Balance:   account.Balance.String(),
	}
	writeSuccess(w, http.StatusCreated, resp)
}

func (h *AccountHandler) GetAccount(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idStr := r.PathValue("id")
	accountID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		log.Debug().Str("id", idStr).Msg("Invalid account ID format")
		writeError(w, http.StatusBadRequest, "invalid_id", "Account ID must be a valid integer")
		return
	}
	if accountID <= 0 {
		log.Debug().Int64("id", accountID).Msg("Account ID must be positive")
		writeError(w, http.StatusBadRequest, "invalid_id", "Account ID must be a positive integer")
		return
	}

	account, err := h.accountService.GetAccount(ctx, accountID)
	if err != nil {
		handleServiceError(ctx, w, err)
		return
	}

	resp := models.GetAccountResponse{
		AccountID: account.AccountID,
		Balance:   account.Balance.String(),
	}
	writeSuccess(w, http.StatusOK, resp)
}

func decodeJSONBody(r *http.Request, target interface{}) error {
	const maxBodySize = 1 << 20
	r.Body = http.MaxBytesReader(nil, r.Body, maxBodySize)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(target); err != nil {
		return err
	}

	if decoder.More() {
		return errors.New("body must only contain a single JSON object")
	}

	return nil
}

func handleServiceError(ctx context.Context, w http.ResponseWriter, err error) {
	if errors.Is(err, context.Canceled) {
		log.Debug().Msg("Request cancelled by client")
		writeError(w, http.StatusBadRequest, "request_cancelled", "Request was cancelled")
		return
	}
	if errors.Is(err, context.DeadlineExceeded) {
		log.Warn().Msg("Request timeout")
		writeError(w, http.StatusGatewayTimeout, "timeout", "Request timed out")
		return
	}

	var domainErr *models.DomainError
	if errors.As(err, &domainErr) {
		status, errorCode, message := mapDomainError(domainErr)
		if status >= 500 {
			log.Error().Err(err).Str("code", string(domainErr.Code)).Msg("Internal error")
		}
		writeError(w, status, errorCode, message)
		return
	}

	switch {
	case errors.Is(err, models.ErrAccountNotFound):
		writeError(w, http.StatusNotFound, "account_not_found", "Account not found")
	case errors.Is(err, models.ErrAccountAlreadyExists):
		writeError(w, http.StatusConflict, "account_exists", "Account already exists")
	case errors.Is(err, models.ErrInsufficientBalance):
		writeError(w, http.StatusUnprocessableEntity, "insufficient_balance", "Insufficient balance for this transaction")
	case errors.Is(err, models.ErrInvalidAmount):
		writeError(w, http.StatusBadRequest, "invalid_amount", "Amount must be a positive decimal value")
	case errors.Is(err, models.ErrSameAccount):
		writeError(w, http.StatusBadRequest, "same_account", "Source and destination accounts cannot be the same")
	case errors.Is(err, models.ErrTransferNotFound):
		writeError(w, http.StatusNotFound, "transaction_not_found", "Transaction not found")
	case errors.Is(err, models.ErrDuplicateTransaction):
		writeError(w, http.StatusConflict, "duplicate_transaction", "Duplicate transaction detected")
	case errors.Is(err, io.EOF):
		writeError(w, http.StatusBadRequest, "invalid_json", "Request body is empty")
	default:
		log.Error().Err(err).Msg("Unexpected error in handler")
		writeError(w, http.StatusInternalServerError, "internal_error", "An unexpected error occurred. Please try again later.")
	}
}

func mapDomainError(err *models.DomainError) (status int, errorCode string, message string) {
	switch err.Code {
	case models.CodeAccountNotFound:
		return http.StatusNotFound, string(err.Code), err.Message
	case models.CodeAccountAlreadyExists:
		return http.StatusConflict, string(err.Code), err.Message
	case models.CodeInsufficientBalance:
		return http.StatusUnprocessableEntity, string(err.Code), err.Message
	case models.CodeInvalidAmount:
		return http.StatusBadRequest, string(err.Code), err.Message
	case models.CodeSameAccount:
		return http.StatusBadRequest, string(err.Code), err.Message
	case models.CodeTransferNotFound:
		return http.StatusNotFound, string(err.Code), err.Message
	case models.CodeDuplicateTransaction:
		return http.StatusConflict, string(err.Code), err.Message
	case models.CodeDatabaseError, models.CodeTransactionFailed, models.CodeInternalError:
		return http.StatusInternalServerError, "internal_error", "An unexpected error occurred. Please try again later."
	default:
		return http.StatusInternalServerError, "internal_error", "An unexpected error occurred. Please try again later."
	}
}
