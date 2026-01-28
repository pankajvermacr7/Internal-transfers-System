package validator

import (
	"fmt"

	"internal-transfers-system/internal/models"

	"github.com/shopspring/decimal"
)

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return ""
	}
	return fmt.Sprintf("validation failed: %d error(s)", len(e))
}

func (e ValidationErrors) HasErrors() bool {
	return len(e) > 0
}

func ValidateCreateAccount(req *models.CreateAccountRequest) ValidationErrors {
	var errs ValidationErrors

	if req.AccountID <= 0 {
		errs = append(errs, ValidationError{Field: "account_id", Message: "must be a positive integer"})
	}

	if req.InitialBalance == "" {
		errs = append(errs, ValidationError{Field: "initial_balance", Message: "is required"})
	} else {
		balance, err := decimal.NewFromString(req.InitialBalance)
		if err != nil {
			errs = append(errs, ValidationError{Field: "initial_balance", Message: "must be a valid decimal number"})
		} else if balance.LessThan(decimal.Zero) {
			errs = append(errs, ValidationError{Field: "initial_balance", Message: "cannot be negative"})
		}
	}

	return errs
}

func ValidateCreateTransaction(req *models.CreateTransactionRequest) ValidationErrors {
	var errs ValidationErrors

	if req.SourceAccountID <= 0 {
		errs = append(errs, ValidationError{Field: "source_account_id", Message: "must be a positive integer"})
	}

	if req.DestinationAccountID <= 0 {
		errs = append(errs, ValidationError{Field: "destination_account_id", Message: "must be a positive integer"})
	}

	if req.SourceAccountID > 0 && req.DestinationAccountID > 0 && req.SourceAccountID == req.DestinationAccountID {
		errs = append(errs, ValidationError{Field: "destination_account_id", Message: "cannot be the same as source_account_id"})
	}

	if req.Amount == "" {
		errs = append(errs, ValidationError{Field: "amount", Message: "is required"})
	} else {
		amount, err := decimal.NewFromString(req.Amount)
		if err != nil {
			errs = append(errs, ValidationError{Field: "amount", Message: "must be a valid decimal number"})
		} else if amount.LessThanOrEqual(decimal.Zero) {
			errs = append(errs, ValidationError{Field: "amount", Message: "must be greater than zero"})
		}
	}

	return errs
}
