package interfaces

import (
	"context"

	"internal-transfers-system/internal/models"

	"github.com/jackc/pgx/v5"
)

// TransactionRepository defines the contract for transaction data operations.
// Implementations must ensure thread-safety and proper error handling.
type TransactionRepository interface {
	// Create inserts a new transaction record within a database transaction.
	// The transaction's TransactionID and CreatedAt fields are populated from the database.
	//
	// This method must be called within an active database transaction (tx).
	// The caller is responsible for committing or rolling back the transaction.
	//
	// The database enforces:
	//   - amount > 0 via CHECK constraint
	//   - source and destination accounts exist via FOREIGN KEY constraints
	//   - source != destination via CHECK constraint
	Create(ctx context.Context, tx pgx.Tx, transaction *models.Transaction) error

	// GetByID retrieves a transaction by its ID.
	// Returns ErrTransferNotFound if the transaction does not exist.
	GetByID(ctx context.Context, transactionID int64) (*models.Transaction, error)

	// GetByAccountID retrieves transactions for a given account with pagination.
	// Returns transactions where the account is either source or destination,
	// ordered by creation time (newest first).
	//
	// Parameters:
	//   - accountID: The account to get transactions for
	//   - limit: Maximum number of transactions to return
	//   - offset: Number of transactions to skip for pagination
	//
	// Returns an empty slice if no transactions are found (not an error).
	GetByAccountID(ctx context.Context, accountID int64, limit, offset int) ([]*models.Transaction, error)
}
