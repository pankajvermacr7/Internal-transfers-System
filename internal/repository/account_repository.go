package repository

import (
	"context"
	"errors"
	"fmt"

	"internal-transfers-system/internal/interfaces"
	"internal-transfers-system/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

// Compile-time check to ensure AccountRepository implements interfaces.AccountRepository.
var _ interfaces.AccountRepository = (*AccountRepository)(nil)

// AccountRepository provides data access operations for accounts.
// All methods are safe for concurrent use.
type AccountRepository struct {
	db *pgxpool.Pool
}

// NewAccountRepository creates a new AccountRepository with the given connection pool.
func NewAccountRepository(db *pgxpool.Pool) *AccountRepository {
	return &AccountRepository{db: db}
}

// Create inserts a new account into the database.
// The account's CreatedAt and UpdatedAt fields are populated from the database.
// Returns an error if the account already exists (duplicate key) or on database failure.
func (r *AccountRepository) Create(ctx context.Context, account *models.Account) error {
	query := `
		INSERT INTO accounts (account_id, balance, created_at, updated_at)
		VALUES ($1, $2, NOW(), NOW())
		RETURNING created_at, updated_at`

	err := r.db.QueryRow(ctx, query, account.AccountID, account.Balance).
		Scan(&account.CreatedAt, &account.UpdatedAt)
	if err != nil {
		return fmt.Errorf("insert account %d: %w", account.AccountID, err)
	}
	return nil
}

// GetByID retrieves an account by its ID.
// Returns ErrAccountNotFound if the account does not exist.
func (r *AccountRepository) GetByID(ctx context.Context, accountID int64) (*models.Account, error) {
	query := `
		SELECT account_id, balance, created_at, updated_at
		FROM accounts
		WHERE account_id = $1`

	account := &models.Account{}
	err := r.db.QueryRow(ctx, query, accountID).
		Scan(&account.AccountID, &account.Balance, &account.CreatedAt, &account.UpdatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, models.ErrAccountNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get account %d: %w", accountID, err)
	}
	return account, nil
}

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
func (r *AccountRepository) GetByIDForUpdate(ctx context.Context, tx pgx.Tx, accountID int64) (*models.Account, error) {
	query := `
		SELECT account_id, balance, created_at, updated_at
		FROM accounts
		WHERE account_id = $1
		FOR UPDATE`

	account := &models.Account{}
	err := tx.QueryRow(ctx, query, accountID).
		Scan(&account.AccountID, &account.Balance, &account.CreatedAt, &account.UpdatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, models.ErrAccountNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get account %d for update: %w", accountID, err)
	}
	return account, nil
}

// UpdateBalance updates the balance of an account within a transaction.
// Returns an error if the update fails or if no rows were affected (account not found).
// The database CHECK constraint ensures the balance cannot go negative.
func (r *AccountRepository) UpdateBalance(ctx context.Context, tx pgx.Tx, accountID int64, newBalance decimal.Decimal) error {
	query := `UPDATE accounts SET balance = $1, updated_at = NOW() WHERE account_id = $2`

	result, err := tx.Exec(ctx, query, newBalance, accountID)
	if err != nil {
		return fmt.Errorf("update balance for account %d: %w", accountID, err)
	}

	// Verify exactly one row was updated
	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return models.ErrAccountNotFound
	}
	if rowsAffected != 1 {
		return fmt.Errorf("expected 1 row affected, got %d for account %d", rowsAffected, accountID)
	}

	return nil
}

// Exists checks if an account with the given ID exists.
// Returns (false, nil) if the account doesn't exist, (true, nil) if it does.
func (r *AccountRepository) Exists(ctx context.Context, accountID int64) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM accounts WHERE account_id = $1)`
	var exists bool
	err := r.db.QueryRow(ctx, query, accountID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check account %d exists: %w", accountID, err)
	}
	return exists, nil
}

// BeginTx starts a new database transaction with READ COMMITTED isolation level.
// This isolation level prevents dirty reads while allowing better concurrency.
// The caller is responsible for calling Commit() or Rollback() on the returned transaction.
func (r *AccountRepository) BeginTx(ctx context.Context) (pgx.Tx, error) {
	txOptions := pgx.TxOptions{
		IsoLevel:   pgx.ReadCommitted,
		AccessMode: pgx.ReadWrite,
	}
	tx, err := r.db.BeginTx(ctx, txOptions)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	return tx, nil
}
