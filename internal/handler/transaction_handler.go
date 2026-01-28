package handler

import (
	"net/http"
	"time"

	"internal-transfers-system/internal/models"
	"internal-transfers-system/internal/service"
	"internal-transfers-system/internal/validator"

	"github.com/rs/zerolog/log"
)

type TransactionResponse struct {
	TransactionID        int64  `json:"transaction_id"`
	SourceAccountID      int64  `json:"source_account_id"`
	DestinationAccountID int64  `json:"destination_account_id"`
	Amount               string `json:"amount"`
	CreatedAt            string `json:"created_at"`
}

type TransactionHandler struct {
	transferService *service.TransferService
}

func NewTransactionHandler(transferService *service.TransferService) *TransactionHandler {
	return &TransactionHandler{transferService: transferService}
}

func (h *TransactionHandler) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req models.CreateTransactionRequest
	if err := decodeJSONBody(r, &req); err != nil {
		log.Debug().Err(err).Msg("Failed to decode create transaction request")
		writeError(w, http.StatusBadRequest, "invalid_json", "Invalid JSON body")
		return
	}

	if errs := validator.ValidateCreateTransaction(&req); len(errs) > 0 {
		log.Debug().
			Int64("sourceAccountID", req.SourceAccountID).
			Int64("destAccountID", req.DestinationAccountID).
			Str("amount", req.Amount).
			Interface("errors", errs).
			Msg("Create transaction validation failed")
		writeValidationError(w, errs)
		return
	}

	txn, err := h.transferService.Transfer(ctx, &req)
	if err != nil {
		handleServiceError(ctx, w, err)
		return
	}

	resp := TransactionResponse{
		TransactionID:        txn.TransactionID,
		SourceAccountID:      txn.SourceAccountID,
		DestinationAccountID: txn.DestinationAccountID,
		Amount:               txn.Amount.String(),
		CreatedAt:            txn.CreatedAt.Format(time.RFC3339),
	}
	writeSuccess(w, http.StatusCreated, resp)
}
