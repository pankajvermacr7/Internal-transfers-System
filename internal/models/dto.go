package models

// CreateAccountRequest represents the request body for creating a new account.
// POST /api/v1/accounts
type CreateAccountRequest struct {
	// AccountID is the unique identifier for the account.
	// Must be a positive integer provided by the client.
	AccountID int64 `json:"account_id"`

	// InitialBalance is the starting balance for the account.
	// Must be a valid decimal string (e.g., "1000.00", "0", "100.50").
	// Cannot be negative.
	InitialBalance string `json:"initial_balance"`
}

// GetAccountResponse represents the response body for account retrieval.
// GET /api/v1/accounts/{id}
type GetAccountResponse struct {
	// AccountID is the unique identifier of the account.
	AccountID int64 `json:"account_id"`

	// Balance is the current balance as a decimal string.
	// Returned as string to preserve decimal precision.
	Balance string `json:"balance"`
}

// CreateTransactionRequest represents the request body for creating a transfer.
// POST /api/v1/transactions
type CreateTransactionRequest struct {
	// SourceAccountID is the account from which funds will be deducted.
	// Must be a positive integer and the account must exist.
	SourceAccountID int64 `json:"source_account_id"`

	// DestinationAccountID is the account to which funds will be credited.
	// Must be a positive integer, must exist, and must differ from SourceAccountID.
	DestinationAccountID int64 `json:"destination_account_id"`

	// Amount is the transfer amount as a decimal string.
	// Must be a positive decimal (e.g., "100.00", "50.50").
	Amount string `json:"amount"`
}
