package service

import (
	"context"
	"time"

	"internal-transfers-system/internal/interfaces"
	"internal-transfers-system/internal/models"

	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"
)

type TransferServiceConfig struct {
	MaxRetries     int
	RetryBaseDelay time.Duration
}

func DefaultTransferConfig() TransferServiceConfig {
	return TransferServiceConfig{
		MaxRetries:     3,
		RetryBaseDelay: 100 * time.Millisecond,
	}
}

type TransferService struct {
	accountRepo     interfaces.AccountRepository
	transactionRepo interfaces.TransactionRepository
	config          TransferServiceConfig
}

func NewTransferService(
	accountRepo interfaces.AccountRepository,
	transactionRepo interfaces.TransactionRepository,
) *TransferService {
	return NewTransferServiceWithConfig(accountRepo, transactionRepo, DefaultTransferConfig())
}

func NewTransferServiceWithConfig(
	accountRepo interfaces.AccountRepository,
	transactionRepo interfaces.TransactionRepository,
	config TransferServiceConfig,
) *TransferService {
	return &TransferService{
		accountRepo:     accountRepo,
		transactionRepo: transactionRepo,
		config:          config,
	}
}

func (s *TransferService) Transfer(ctx context.Context, req *models.CreateTransactionRequest) (*models.Transaction, error) {
	if req.SourceAccountID == req.DestinationAccountID {
		return nil, models.ErrSameAccount
	}

	amount, err := models.ParseMoney(req.Amount)
	if err != nil {
		log.Debug().Err(err).Str("amount", req.Amount).Msg("Invalid amount format")
		return nil, models.ErrInvalidAmount
	}
	if amount.LessThanOrEqual(decimal.Zero) {
		log.Debug().Str("amount", req.Amount).Msg("Amount must be positive")
		return nil, models.ErrInvalidAmount
	}

	var transaction *models.Transaction
	var lastErr error

	for attempt := 0; attempt <= s.config.MaxRetries; attempt++ {
		if attempt > 0 {
			delay := s.config.RetryBaseDelay * time.Duration(1<<uint(attempt-1))
			log.Debug().Int("attempt", attempt).Dur("delay", delay).Msg("Retrying transfer after transient error")

			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}

		transaction, lastErr = s.executeTransfer(ctx, req.SourceAccountID, req.DestinationAccountID, amount)
		if lastErr == nil {
			return transaction, nil
		}

		if !models.IsRetryable(lastErr) {
			return nil, lastErr
		}

		log.Warn().Err(lastErr).Int("attempt", attempt+1).Int("maxRetries", s.config.MaxRetries).Msg("Transfer failed with retryable error")
	}

	return nil, models.WrapError(models.CodeTransactionFailed, "transfer failed after retries", lastErr)
}

func (s *TransferService) executeTransfer(ctx context.Context, sourceID, destID int64, amount decimal.Decimal) (*models.Transaction, error) {
	tx, err := s.accountRepo.BeginTx(ctx)
	if err != nil {
		return nil, models.WrapError(models.CodeDatabaseError, "failed to begin transaction", err)
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && err.Error() != "tx is closed" {
			log.Error().Err(err).Msg("Failed to rollback transaction")
		}
	}()

	// Lock accounts in consistent order (lower ID first) to prevent deadlocks
	firstID, secondID := sourceID, destID
	if firstID > secondID {
		firstID, secondID = secondID, firstID
	}

	first, err := s.accountRepo.GetByIDForUpdate(ctx, tx, firstID)
	if err != nil {
		return nil, err
	}
	second, err := s.accountRepo.GetByIDForUpdate(ctx, tx, secondID)
	if err != nil {
		return nil, err
	}

	var sourceAccount, destAccount *models.Account
	if firstID == sourceID {
		sourceAccount, destAccount = first, second
	} else {
		sourceAccount, destAccount = second, first
	}

	if sourceAccount.Balance.LessThan(amount) {
		log.Debug().
			Int64("sourceAccountID", sourceID).
			Str("balance", sourceAccount.Balance.String()).
			Str("amount", amount.String()).
			Msg("Insufficient balance for transfer")
		return nil, models.ErrInsufficientBalance
	}

	newSourceBalance := sourceAccount.Balance.Sub(amount)
	newDestBalance := destAccount.Balance.Add(amount)

	if err := s.accountRepo.UpdateBalance(ctx, tx, sourceAccount.AccountID, newSourceBalance); err != nil {
		return nil, models.WrapError(models.CodeDatabaseError, "failed to update source balance", err)
	}

	if err := s.accountRepo.UpdateBalance(ctx, tx, destAccount.AccountID, newDestBalance); err != nil {
		return nil, models.WrapError(models.CodeDatabaseError, "failed to update destination balance", err)
	}

	transaction := &models.Transaction{
		SourceAccountID:      sourceID,
		DestinationAccountID: destID,
		Amount:               amount,
	}
	if err := s.transactionRepo.Create(ctx, tx, transaction); err != nil {
		return nil, models.WrapError(models.CodeDatabaseError, "failed to create transaction record", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, models.WrapError(models.CodeDatabaseError, "failed to commit transaction", err)
	}

	log.Info().
		Int64("transactionID", transaction.TransactionID).
		Int64("sourceAccountID", sourceID).
		Int64("destAccountID", destID).
		Str("amount", amount.String()).
		Msg("Transfer completed successfully")

	return transaction, nil
}

func (s *TransferService) GetTransaction(ctx context.Context, transactionID int64) (*models.Transaction, error) {
	return s.transactionRepo.GetByID(ctx, transactionID)
}

const (
	DefaultPageSize = 20
	MaxPageSize     = 100
)

func (s *TransferService) GetAccountTransactions(ctx context.Context, accountID int64, limit, offset int) ([]*models.Transaction, error) {
	if limit <= 0 {
		limit = DefaultPageSize
	}
	if limit > MaxPageSize {
		limit = MaxPageSize
	}
	if offset < 0 {
		offset = 0
	}

	return s.transactionRepo.GetByAccountID(ctx, accountID, limit, offset)
}
