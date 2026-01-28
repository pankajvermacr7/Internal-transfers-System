//go:build integration

package repository

import (
	"context"
	"testing"

	"internal-transfers-system/internal/models"

	"github.com/shopspring/decimal"
)

func setupTxnRepo(t *testing.T) (*TransactionRepository, *AccountRepository) {
	t.Helper()
	if err := testSuite.Clean(); err != nil {
		t.Fatalf("clean: %v", err)
	}
	return NewTransactionRepository(testSuite.Pool()), NewAccountRepository(testSuite.Pool())
}

func TestTransactionRepository_Create(t *testing.T) {
	txnRepo, accRepo := setupTxnRepo(t)
	ctx := context.Background()

	accRepo.Create(ctx, &models.Account{AccountID: 1, Balance: decimal.NewFromInt(1000)})
	accRepo.Create(ctx, &models.Account{AccountID: 2, Balance: decimal.NewFromInt(500)})

	tx, _ := accRepo.BeginTx(ctx)
	txn := &models.Transaction{
		SourceAccountID:      1,
		DestinationAccountID: 2,
		Amount:               decimal.NewFromInt(100),
	}
	if err := txnRepo.Create(ctx, tx, txn); err != nil {
		tx.Rollback(ctx)
		t.Fatalf("create: %v", err)
	}
	tx.Commit(ctx)

	if txn.TransactionID == 0 {
		t.Error("expected transaction ID")
	}
}

func TestTransactionRepository_GetByID(t *testing.T) {
	txnRepo, accRepo := setupTxnRepo(t)
	ctx := context.Background()

	accRepo.Create(ctx, &models.Account{AccountID: 1, Balance: decimal.NewFromInt(1000)})
	accRepo.Create(ctx, &models.Account{AccountID: 2, Balance: decimal.NewFromInt(500)})

	tx, _ := accRepo.BeginTx(ctx)
	txn := &models.Transaction{
		SourceAccountID: 1, DestinationAccountID: 2, Amount: decimal.NewFromInt(100),
	}
	txnRepo.Create(ctx, tx, txn)
	tx.Commit(ctx)

	found, err := txnRepo.GetByID(ctx, txn.TransactionID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if !found.Amount.Equal(decimal.NewFromInt(100)) {
		t.Errorf("expected 100, got %s", found.Amount)
	}
}

func TestTransactionRepository_GetByID_NotFound(t *testing.T) {
	txnRepo, _ := setupTxnRepo(t)
	_, err := txnRepo.GetByID(context.Background(), 999)
	if err != models.ErrTransferNotFound {
		t.Errorf("expected ErrTransferNotFound, got %v", err)
	}
}

func TestTransactionRepository_GetByAccountID(t *testing.T) {
	txnRepo, accRepo := setupTxnRepo(t)
	ctx := context.Background()

	accRepo.Create(ctx, &models.Account{AccountID: 1, Balance: decimal.NewFromInt(1000)})
	accRepo.Create(ctx, &models.Account{AccountID: 2, Balance: decimal.NewFromInt(1000)})

	for i := 0; i < 5; i++ {
		tx, _ := accRepo.BeginTx(ctx)
		txnRepo.Create(ctx, tx, &models.Transaction{
			SourceAccountID: 1, DestinationAccountID: 2, Amount: decimal.NewFromInt(10),
		})
		tx.Commit(ctx)
	}

	txns, _ := txnRepo.GetByAccountID(ctx, 1, 10, 0)
	if len(txns) != 5 {
		t.Errorf("expected 5, got %d", len(txns))
	}

	// Pagination
	txns, _ = txnRepo.GetByAccountID(ctx, 1, 3, 0)
	if len(txns) != 3 {
		t.Errorf("expected 3, got %d", len(txns))
	}
}
