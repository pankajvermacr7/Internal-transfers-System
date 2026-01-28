package mocks

import (
	"context"
	"sync"
	"sync/atomic"

	"internal-transfers-system/internal/models"

	"github.com/jackc/pgx/v5"
)

type MockTransactionRepository struct {
	mu           sync.RWMutex
	transactions map[int64]*models.Transaction
	nextID       atomic.Int64

	CreateError         error
	GetByIDError        error
	GetByAccountIDError error
}

func NewMockTransactionRepository() *MockTransactionRepository {
	m := &MockTransactionRepository{transactions: make(map[int64]*models.Transaction)}
	m.nextID.Store(1)
	return m
}

func (m *MockTransactionRepository) Create(ctx context.Context, tx pgx.Tx, txn *models.Transaction) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.CreateError != nil {
		return m.CreateError
	}
	txn.TransactionID = m.nextID.Add(1) - 1
	m.transactions[txn.TransactionID] = &models.Transaction{
		TransactionID:        txn.TransactionID,
		SourceAccountID:      txn.SourceAccountID,
		DestinationAccountID: txn.DestinationAccountID,
		Amount:               txn.Amount,
	}
	return nil
}

func (m *MockTransactionRepository) GetByID(ctx context.Context, id int64) (*models.Transaction, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.GetByIDError != nil {
		return nil, m.GetByIDError
	}
	txn, exists := m.transactions[id]
	if !exists {
		return nil, models.ErrTransferNotFound
	}
	return &models.Transaction{
		TransactionID:        txn.TransactionID,
		SourceAccountID:      txn.SourceAccountID,
		DestinationAccountID: txn.DestinationAccountID,
		Amount:               txn.Amount,
	}, nil
}

func (m *MockTransactionRepository) GetByAccountID(ctx context.Context, accountID int64, limit, offset int) ([]*models.Transaction, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.GetByAccountIDError != nil {
		return nil, m.GetByAccountIDError
	}
	var result []*models.Transaction
	for _, txn := range m.transactions {
		if txn.SourceAccountID == accountID || txn.DestinationAccountID == accountID {
			result = append(result, txn)
		}
	}
	if offset >= len(result) {
		return []*models.Transaction{}, nil
	}
	end := offset + limit
	if end > len(result) {
		end = len(result)
	}
	return result[offset:end], nil
}

func (m *MockTransactionRepository) SetTransaction(txn *models.Transaction) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.transactions[txn.TransactionID] = txn
}
