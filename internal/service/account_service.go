package service

import (
	"context"
	"strings"

	"internal-transfers-system/internal/interfaces"
	"internal-transfers-system/internal/models"

	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"
)

type AccountService struct {
	accountRepo interfaces.AccountRepository
}

func NewAccountService(accountRepo interfaces.AccountRepository) *AccountService {
	return &AccountService{accountRepo: accountRepo}
}

func (s *AccountService) CreateAccount(ctx context.Context, req *models.CreateAccountRequest) (*models.Account, error) {
	balance, err := models.ParseMoney(req.InitialBalance)
	if err != nil {
		log.Debug().Err(err).Str("initialBalance", req.InitialBalance).Msg("Invalid initial balance format")
		return nil, models.ErrInvalidAmount
	}

	if balance.LessThan(decimal.Zero) {
		log.Debug().Str("initialBalance", req.InitialBalance).Msg("Initial balance cannot be negative")
		return nil, models.ErrInvalidAmount
	}

	exists, err := s.accountRepo.Exists(ctx, req.AccountID)
	if err != nil {
		log.Error().Err(err).Int64("accountID", req.AccountID).Msg("Failed to check account existence")
		return nil, models.WrapError(models.CodeDatabaseError, "failed to check account existence", err)
	}
	if exists {
		log.Debug().Int64("accountID", req.AccountID).Msg("Account already exists")
		return nil, models.ErrAccountAlreadyExists
	}

	account := &models.Account{
		AccountID: req.AccountID,
		Balance:   balance,
	}

	if err := s.accountRepo.Create(ctx, account); err != nil {
		if isDuplicateKeyError(err) {
			return nil, models.ErrAccountAlreadyExists
		}
		log.Error().Err(err).Int64("accountID", req.AccountID).Msg("Failed to create account")
		return nil, models.WrapError(models.CodeDatabaseError, "failed to create account", err)
	}

	log.Info().Int64("accountID", account.AccountID).Str("balance", account.Balance.String()).Msg("Account created successfully")

	return account, nil
}

func (s *AccountService) GetAccount(ctx context.Context, accountID int64) (*models.Account, error) {
	return s.accountRepo.GetByID(ctx, accountID)
}

func isDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "duplicate key") || strings.Contains(errStr, "23505")
}
