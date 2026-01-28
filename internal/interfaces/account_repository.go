// Package interfaces defines contracts for the data access layer.
// These interfaces enable dependency injection, facilitate testing,
// and enforce separation of concerns between layers.
package interfaces

import (
	"context"

	"internal-transfers-system/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
)

// AccountRepository defines the contract for account data operations.
// Implementations must ensure thread-safety and proper transaction handling.
type AccountRepository interface {
	// Create inserts a new account into the database.
	// The account's CreatedAt and UpdatedAt fields are populated from the database.
	// Returns an error if the account already exists (duplicate key) or on database failure.
	Create(ctx context.Context, account *models.Account) error

	// GetByID retrieves an account by its ID.
	// Returns ErrAccountNotFound if the account does not exist.
	GetByID(ctx context.Context, accountID int64) (*models.Account, error)

	// GetByIDForUpdate retrieves an account with a row-level lock for update.
	// This prevents other transactions from modifying or locking the row until
	// the current transaction completes. Must be called within a transaction.
	//
	// The FOR UPDATE lock ensures:
	//   - No other transaction can modify this row until we commit/rollback
	//   - Concurrent transfers to/from this account are serialized
	//   - Balance consistency is maintained during multi-step operations
	//
	// Returns ErrAccountNotFound if the account does not exist.
	GetByIDForUpdate(ctx context.Context, tx pgx.Tx, accountID int64) (*models.Account, error)

	// UpdateBalance updates the balance of an account within a transaction.
	// Returns an error if the update fails or if no rows were affected (account not found).
	// The database CHECK constraint ensures the balance cannot go negative.
	UpdateBalance(ctx context.Context, tx pgx.Tx, accountID int64, newBalance decimal.Decimal) error

	// Exists checks if an account with the given ID exists.
	// Returns (false, nil) if the account doesn't exist, (true, nil) if it does.
	Exists(ctx context.Context, accountID int64) (bool, error)

	// BeginTx starts a new database transaction with appropriate isolation level.
	// The caller is responsible for calling Commit() or Rollback() on the returned transaction.
	BeginTx(ctx context.Context) (pgx.Tx, error)
}
