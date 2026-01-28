package handler

import (
	"encoding/json"
	"testing"
	"time"

	"internal-transfers-system/internal/models"

	"github.com/shopspring/decimal"
)

func TestTransactionResponse_Format(t *testing.T) {
	txn := &models.Transaction{
		TransactionID:        1,
		SourceAccountID:      100,
		DestinationAccountID: 200,
		Amount:               decimal.RequireFromString("150.50"),
		CreatedAt:            time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
	}

	resp := TransactionResponse{
		TransactionID:        txn.TransactionID,
		SourceAccountID:      txn.SourceAccountID,
		DestinationAccountID: txn.DestinationAccountID,
		Amount:               txn.Amount.String(),
		CreatedAt:            txn.CreatedAt.Format(time.RFC3339),
	}

	if resp.TransactionID != 1 {
		t.Errorf("expected 1, got %d", resp.TransactionID)
	}
	if resp.Amount != "150.5" {
		t.Errorf("expected 150.5, got %s", resp.Amount)
	}
}

func TestTransactionResponse_JSON(t *testing.T) {
	resp := TransactionResponse{
		TransactionID:        1,
		SourceAccountID:      100,
		DestinationAccountID: 200,
		Amount:               "150.50",
		CreatedAt:            "2024-01-15T10:30:00Z",
	}

	data, _ := json.Marshal(resp)
	var parsed map[string]interface{}
	json.Unmarshal(data, &parsed)

	expected := []string{"transaction_id", "source_account_id", "destination_account_id", "amount", "created_at"}
	for _, field := range expected {
		if _, ok := parsed[field]; !ok {
			t.Errorf("missing field %q", field)
		}
	}
}
