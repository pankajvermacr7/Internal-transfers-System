package service

import (
	"context"
	"errors"
	"testing"

	"internal-transfers-system/internal/mocks"
	"internal-transfers-system/internal/models"

	"github.com/shopspring/decimal"
)

func TestAccountService_CreateAccount(t *testing.T) {
	tests := []struct {
		name      string
		request   *models.CreateAccountRequest
		setup     func(*mocks.MockAccountRepository)
		wantErr   error
		wantBal   string
	}{
		{
			name:    "success",
			request: &models.CreateAccountRequest{AccountID: 1, InitialBalance: "1000.00"},
			setup:   func(*mocks.MockAccountRepository) {},
			wantBal: "1000",
		},
		{
			name:    "zero balance",
			request: &models.CreateAccountRequest{AccountID: 2, InitialBalance: "0"},
			setup:   func(*mocks.MockAccountRepository) {},
			wantBal: "0",
		},
		{
			name:    "invalid balance",
			request: &models.CreateAccountRequest{AccountID: 1, InitialBalance: "abc"},
			setup:   func(*mocks.MockAccountRepository) {},
			wantErr: models.ErrInvalidAmount,
		},
		{
			name:    "negative balance",
			request: &models.CreateAccountRequest{AccountID: 1, InitialBalance: "-100"},
			setup:   func(*mocks.MockAccountRepository) {},
			wantErr: models.ErrInvalidAmount,
		},
		{
			name:    "duplicate account",
			request: &models.CreateAccountRequest{AccountID: 1, InitialBalance: "1000"},
			setup: func(m *mocks.MockAccountRepository) {
				m.SetAccount(&models.Account{AccountID: 1, Balance: decimal.NewFromInt(500)})
			},
			wantErr: models.ErrAccountAlreadyExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewMockAccountRepository()
			tt.setup(repo)

			svc := NewAccountService(repo)
			acc, err := svc.CreateAccount(context.Background(), tt.request)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("expected %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected: %v", err)
			}
			if acc.Balance.String() != tt.wantBal {
				t.Errorf("expected balance %s, got %s", tt.wantBal, acc.Balance)
			}
		})
	}
}

func TestAccountService_GetAccount(t *testing.T) {
	repo := mocks.NewMockAccountRepository()
	repo.SetAccount(&models.Account{AccountID: 1, Balance: decimal.NewFromInt(1000)})

	svc := NewAccountService(repo)

	// Found
	acc, err := svc.GetAccount(context.Background(), 1)
	if err != nil || acc.AccountID != 1 {
		t.Errorf("expected account 1, got err=%v", err)
	}

	// Not found
	_, err = svc.GetAccount(context.Background(), 999)
	if !errors.Is(err, models.ErrAccountNotFound) {
		t.Errorf("expected ErrAccountNotFound, got %v", err)
	}
}
