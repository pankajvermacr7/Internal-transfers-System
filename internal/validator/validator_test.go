package validator

import (
	"testing"

	"internal-transfers-system/internal/models"
)

func TestValidateCreateAccount(t *testing.T) {
	tests := []struct {
		name    string
		req     *models.CreateAccountRequest
		wantErr bool
	}{
		{"valid", &models.CreateAccountRequest{AccountID: 1, InitialBalance: "1000"}, false},
		{"zero id", &models.CreateAccountRequest{AccountID: 0, InitialBalance: "1000"}, true},
		{"negative id", &models.CreateAccountRequest{AccountID: -1, InitialBalance: "1000"}, true},
		{"missing balance", &models.CreateAccountRequest{AccountID: 1, InitialBalance: ""}, true},
		{"invalid balance", &models.CreateAccountRequest{AccountID: 1, InitialBalance: "abc"}, true},
		{"negative balance", &models.CreateAccountRequest{AccountID: 1, InitialBalance: "-100"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := ValidateCreateAccount(tt.req)
			if (len(errs) > 0) != tt.wantErr {
				t.Errorf("wantErr=%v, got %v errors", tt.wantErr, len(errs))
			}
		})
	}
}

func TestValidateCreateTransaction(t *testing.T) {
	tests := []struct {
		name    string
		req     *models.CreateTransactionRequest
		wantErr bool
	}{
		{"valid", &models.CreateTransactionRequest{SourceAccountID: 1, DestinationAccountID: 2, Amount: "100"}, false},
		{"zero source", &models.CreateTransactionRequest{SourceAccountID: 0, DestinationAccountID: 2, Amount: "100"}, true},
		{"zero dest", &models.CreateTransactionRequest{SourceAccountID: 1, DestinationAccountID: 0, Amount: "100"}, true},
		{"same account", &models.CreateTransactionRequest{SourceAccountID: 1, DestinationAccountID: 1, Amount: "100"}, true},
		{"missing amount", &models.CreateTransactionRequest{SourceAccountID: 1, DestinationAccountID: 2, Amount: ""}, true},
		{"zero amount", &models.CreateTransactionRequest{SourceAccountID: 1, DestinationAccountID: 2, Amount: "0"}, true},
		{"negative amount", &models.CreateTransactionRequest{SourceAccountID: 1, DestinationAccountID: 2, Amount: "-100"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := ValidateCreateTransaction(tt.req)
			if (len(errs) > 0) != tt.wantErr {
				t.Errorf("wantErr=%v, got %v errors", tt.wantErr, len(errs))
			}
		})
	}
}
