package service

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"internal-transfers-system/internal/mocks"
	"internal-transfers-system/internal/models"

	"github.com/shopspring/decimal"
)

func TestTransferService_Transfer(t *testing.T) {
	tests := []struct {
		name          string
		request       *models.CreateTransactionRequest
		setupMock     func(*mocks.MockAccountRepository, *mocks.MockTransactionRepository)
		expectedError error
		validate      func(*testing.T, *models.Transaction, *mocks.MockAccountRepository)
	}{
		{
			name: "successful transfer",
			request: &models.CreateTransactionRequest{
				SourceAccountID:      1,
				DestinationAccountID: 2,
				Amount:               "100.00",
			},
			setupMock: func(accRepo *mocks.MockAccountRepository, _ *mocks.MockTransactionRepository) {
				accRepo.SetAccount(&models.Account{AccountID: 1, Balance: decimal.NewFromInt(1000)})
				accRepo.SetAccount(&models.Account{AccountID: 2, Balance: decimal.NewFromInt(500)})
			},
			validate: func(t *testing.T, txn *models.Transaction, accRepo *mocks.MockAccountRepository) {
				if !txn.Amount.Equal(decimal.NewFromInt(100)) {
					t.Errorf("expected amount 100, got %s", txn.Amount)
				}
				src, _ := accRepo.GetAccount(1)
				if !src.Balance.Equal(decimal.NewFromInt(900)) {
					t.Errorf("expected source balance 900, got %s", src.Balance)
				}
			},
		},
		{
			name: "same account",
			request: &models.CreateTransactionRequest{
				SourceAccountID:      1,
				DestinationAccountID: 1,
				Amount:               "100.00",
			},
			setupMock:     func(*mocks.MockAccountRepository, *mocks.MockTransactionRepository) {},
			expectedError: models.ErrSameAccount,
		},
		{
			name: "invalid amount",
			request: &models.CreateTransactionRequest{
				SourceAccountID:      1,
				DestinationAccountID: 2,
				Amount:               "abc",
			},
			setupMock:     func(*mocks.MockAccountRepository, *mocks.MockTransactionRepository) {},
			expectedError: models.ErrInvalidAmount,
		},
		{
			name: "zero amount",
			request: &models.CreateTransactionRequest{
				SourceAccountID:      1,
				DestinationAccountID: 2,
				Amount:               "0",
			},
			setupMock:     func(*mocks.MockAccountRepository, *mocks.MockTransactionRepository) {},
			expectedError: models.ErrInvalidAmount,
		},
		{
			name: "source not found",
			request: &models.CreateTransactionRequest{
				SourceAccountID:      999,
				DestinationAccountID: 2,
				Amount:               "100.00",
			},
			setupMock: func(accRepo *mocks.MockAccountRepository, _ *mocks.MockTransactionRepository) {
				accRepo.SetAccount(&models.Account{AccountID: 2, Balance: decimal.NewFromInt(500)})
			},
			expectedError: models.ErrAccountNotFound,
		},
		{
			name: "insufficient balance",
			request: &models.CreateTransactionRequest{
				SourceAccountID:      1,
				DestinationAccountID: 2,
				Amount:               "2000.00",
			},
			setupMock: func(accRepo *mocks.MockAccountRepository, _ *mocks.MockTransactionRepository) {
				accRepo.SetAccount(&models.Account{AccountID: 1, Balance: decimal.NewFromInt(1000)})
				accRepo.SetAccount(&models.Account{AccountID: 2, Balance: decimal.NewFromInt(500)})
			},
			expectedError: models.ErrInsufficientBalance,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accRepo := mocks.NewMockAccountRepository()
			txnRepo := mocks.NewMockTransactionRepository()
			tt.setupMock(accRepo, txnRepo)

			svc := NewTransferService(accRepo, txnRepo)
			txn, err := svc.Transfer(context.Background(), tt.request)

			if tt.expectedError != nil {
				if !errors.Is(err, tt.expectedError) {
					t.Errorf("expected %v, got %v", tt.expectedError, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.validate != nil {
				tt.validate(t, txn, accRepo)
			}
		})
	}
}

func TestTransferService_LockOrdering(t *testing.T) {
	accRepo := mocks.NewMockAccountRepository()
	txnRepo := mocks.NewMockTransactionRepository()

	accRepo.SetAccount(&models.Account{AccountID: 1, Balance: decimal.NewFromInt(1000)})
	accRepo.SetAccount(&models.Account{AccountID: 2, Balance: decimal.NewFromInt(500)})

	var lockOrder []int64
	var mu sync.Mutex

	accRepo.OnGetByIDForUpdate = func(_ context.Context, _ interface{}, id int64) (*models.Account, error) {
		mu.Lock()
		lockOrder = append(lockOrder, id)
		mu.Unlock()
		acc, _ := accRepo.GetAccountUnsafe(id)
		if acc == nil {
			return nil, models.ErrAccountNotFound
		}
		return acc, nil
	}

	svc := NewTransferService(accRepo, txnRepo)

	svc.Transfer(context.Background(), &models.CreateTransactionRequest{
		SourceAccountID:      2,
		DestinationAccountID: 1,
		Amount:               "100.00",
	})

	if len(lockOrder) != 2 || lockOrder[0] != 1 || lockOrder[1] != 2 {
		t.Errorf("expected lock order [1,2], got %v", lockOrder)
	}
}

func TestTransferService_RetryOnDeadlock(t *testing.T) {
	accRepo := mocks.NewMockAccountRepository()
	txnRepo := mocks.NewMockTransactionRepository()

	accRepo.SetAccount(&models.Account{AccountID: 1, Balance: decimal.NewFromInt(1000)})
	accRepo.SetAccount(&models.Account{AccountID: 2, Balance: decimal.NewFromInt(500)})

	var calls atomic.Int32
	accRepo.OnGetByIDForUpdate = func(_ context.Context, _ interface{}, id int64) (*models.Account, error) {
		if calls.Add(1) <= 2 {
			return nil, errors.New("deadlock detected")
		}
		acc, _ := accRepo.GetAccountUnsafe(id)
		return acc, nil
	}

	config := TransferServiceConfig{MaxRetries: 3, RetryBaseDelay: time.Millisecond}
	svc := NewTransferServiceWithConfig(accRepo, txnRepo, config)

	txn, err := svc.Transfer(context.Background(), &models.CreateTransactionRequest{
		SourceAccountID:      1,
		DestinationAccountID: 2,
		Amount:               "100.00",
	})

	if err != nil {
		t.Fatalf("expected success after retry, got: %v", err)
	}
	if txn == nil {
		t.Error("expected transaction")
	}
}

func TestTransferService_GetTransaction(t *testing.T) {
	accRepo := mocks.NewMockAccountRepository()
	txnRepo := mocks.NewMockTransactionRepository()

	txnRepo.SetTransaction(&models.Transaction{
		TransactionID:        1,
		SourceAccountID:      1,
		DestinationAccountID: 2,
		Amount:               decimal.NewFromInt(100),
	})

	svc := NewTransferService(accRepo, txnRepo)

	txn, err := svc.GetTransaction(context.Background(), 1)
	if err != nil || txn.TransactionID != 1 {
		t.Errorf("expected transaction 1, got err=%v", err)
	}

	_, err = svc.GetTransaction(context.Background(), 999)
	if !errors.Is(err, models.ErrTransferNotFound) {
		t.Errorf("expected ErrTransferNotFound, got %v", err)
	}
}
