package mocks

import (
	"context"
	"sync"

	"internal-transfers-system/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/shopspring/decimal"
)

type MockAccountRepository struct {
	mu       sync.RWMutex
	accounts map[int64]*models.Account

	CreateError           error
	GetByIDError          error
	GetByIDForUpdateError error
	UpdateBalanceError    error
	ExistsError           error
	BeginTxError          error

	OnGetByIDForUpdate func(ctx context.Context, tx interface{}, accountID int64) (*models.Account, error)
}

func NewMockAccountRepository() *MockAccountRepository {
	return &MockAccountRepository{accounts: make(map[int64]*models.Account)}
}

func (m *MockAccountRepository) Create(ctx context.Context, account *models.Account) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.CreateError != nil {
		return m.CreateError
	}
	if _, exists := m.accounts[account.AccountID]; exists {
		return models.ErrAccountAlreadyExists
	}
	m.accounts[account.AccountID] = &models.Account{
		AccountID: account.AccountID,
		Balance:   account.Balance,
	}
	return nil
}

func (m *MockAccountRepository) GetByID(ctx context.Context, id int64) (*models.Account, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.GetByIDError != nil {
		return nil, m.GetByIDError
	}
	acc, exists := m.accounts[id]
	if !exists {
		return nil, models.ErrAccountNotFound
	}
	return &models.Account{AccountID: acc.AccountID, Balance: acc.Balance}, nil
}

func (m *MockAccountRepository) GetByIDForUpdate(ctx context.Context, tx pgx.Tx, id int64) (*models.Account, error) {
	if m.OnGetByIDForUpdate != nil {
		return m.OnGetByIDForUpdate(ctx, tx, id)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.GetByIDForUpdateError != nil {
		return nil, m.GetByIDForUpdateError
	}
	acc, exists := m.accounts[id]
	if !exists {
		return nil, models.ErrAccountNotFound
	}
	return &models.Account{AccountID: acc.AccountID, Balance: acc.Balance}, nil
}

func (m *MockAccountRepository) UpdateBalance(ctx context.Context, tx pgx.Tx, id int64, balance decimal.Decimal) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.UpdateBalanceError != nil {
		return m.UpdateBalanceError
	}
	acc, exists := m.accounts[id]
	if !exists {
		return models.ErrAccountNotFound
	}
	acc.Balance = balance
	return nil
}

func (m *MockAccountRepository) Exists(ctx context.Context, id int64) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.ExistsError != nil {
		return false, m.ExistsError
	}
	_, exists := m.accounts[id]
	return exists, nil
}

func (m *MockAccountRepository) BeginTx(ctx context.Context) (pgx.Tx, error) {
	if m.BeginTxError != nil {
		return nil, m.BeginTxError
	}
	return &MockTx{}, nil
}

func (m *MockAccountRepository) SetAccount(acc *models.Account) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.accounts[acc.AccountID] = acc
}

func (m *MockAccountRepository) GetAccount(id int64) (*models.Account, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	acc, exists := m.accounts[id]
	return acc, exists
}

func (m *MockAccountRepository) GetAccountUnsafe(id int64) (*models.Account, bool) {
	acc, exists := m.accounts[id]
	return acc, exists
}

type MockTx struct{}

func (m *MockTx) Begin(ctx context.Context) (pgx.Tx, error)         { return &MockTx{}, nil }
func (m *MockTx) Commit(ctx context.Context) error                  { return nil }
func (m *MockTx) Rollback(ctx context.Context) error                { return nil }
func (m *MockTx) CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error) {
	return 0, nil
}
func (m *MockTx) SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults { return nil }
func (m *MockTx) LargeObjects() pgx.LargeObjects                               { return pgx.LargeObjects{} }
func (m *MockTx) Prepare(ctx context.Context, name, sql string) (*pgconn.StatementDescription, error) {
	return nil, nil
}
func (m *MockTx) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, nil
}
func (m *MockTx) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) { return nil, nil }
func (m *MockTx) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row        { return nil }
func (m *MockTx) Conn() *pgx.Conn                                                      { return nil }
