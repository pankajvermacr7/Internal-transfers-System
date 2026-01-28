package repository

import (
	"context"
	"errors"
	"fmt"

	"internal-transfers-system/internal/interfaces"
	"internal-transfers-system/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Compile-time check to ensure TransactionRepository implements interfaces.TransactionRepository.
var _ interfaces.TransactionRepository = (*TransactionRepository)(nil)

// TransactionRepository provides data access operations for transactions.
// All methods are safe for concurrent use.
type TransactionRepository struct {
	db *pgxpool.Pool
}

// NewTransactionRepository creates a new TransactionRepository with the given connection pool.
func NewTransactionRepository(db *pgxpool.Pool) *TransactionRepository {
	return &TransactionRepository{db: db}
}

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
func (r *TransactionRepository) Create(ctx context.Context, tx pgx.Tx, transaction *models.Transaction) error {
	query := `
		INSERT INTO transactions (source_account_id, destination_account_id, amount, created_at)
		VALUES ($1, $2, $3, NOW())
		RETURNING transaction_id, created_at`

	err := tx.QueryRow(ctx, query,
		transaction.SourceAccountID,
		transaction.DestinationAccountID,
		transaction.Amount,
	).Scan(&transaction.TransactionID, &transaction.CreatedAt)

	if err != nil {
		return fmt.Errorf("insert transaction: %w", err)
	}
	return nil
}

// GetByID retrieves a transaction by its ID.
// Returns ErrTransferNotFound if the transaction does not exist.
func (r *TransactionRepository) GetByID(ctx context.Context, transactionID int64) (*models.Transaction, error) {
	query := `
		SELECT transaction_id, source_account_id, destination_account_id, amount, created_at
		FROM transactions
		WHERE transaction_id = $1`

	txn := &models.Transaction{}
	err := r.db.QueryRow(ctx, query, transactionID).
		Scan(&txn.TransactionID, &txn.SourceAccountID, &txn.DestinationAccountID, &txn.Amount, &txn.CreatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, models.ErrTransferNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get transaction %d: %w", transactionID, err)
	}
	return txn, nil
}

// GetByAccountID retrieves transactions for a given account with pagination.
// Returns transactions where the account is either source or destination,
// ordered by creation time (newest first).
//
// Parameters:
//   - accountID: The account to get transactions for
//   - limit: Maximum number of transactions to return (should be validated by caller)
//   - offset: Number of transactions to skip for pagination
//
// Returns an empty slice if no transactions are found (not an error).
func (r *TransactionRepository) GetByAccountID(ctx context.Context, accountID int64, limit, offset int) ([]*models.Transaction, error) {
	query := `
		SELECT transaction_id, source_account_id, destination_account_id, amount, created_at
		FROM transactions
		WHERE source_account_id = $1 OR destination_account_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.Query(ctx, query, accountID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("query transactions for account %d: %w", accountID, err)
	}
	defer rows.Close()

	// Pre-allocate with expected capacity to reduce allocations
	transactions := make([]*models.Transaction, 0, limit)

	for rows.Next() {
		txn := &models.Transaction{}
		if err := rows.Scan(
			&txn.TransactionID,
			&txn.SourceAccountID,
			&txn.DestinationAccountID,
			&txn.Amount,
			&txn.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan transaction row: %w", err)
		}
		transactions = append(transactions, txn)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate transaction rows: %w", err)
	}

	return transactions, nil
}
