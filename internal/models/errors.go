package models

import (
	"errors"
	"fmt"
	"strings"
)

type ErrorCode string

const (
	CodeAccountNotFound      ErrorCode = "account_not_found"
	CodeInsufficientBalance  ErrorCode = "insufficient_balance"
	CodeInvalidAmount        ErrorCode = "invalid_amount"
	CodeCurrencyMismatch     ErrorCode = "currency_mismatch"
	CodeSameAccount          ErrorCode = "same_account"
	CodeTransferNotFound     ErrorCode = "transaction_not_found"
	CodeAccountAlreadyExists ErrorCode = "account_exists"
	CodeDuplicateTransaction ErrorCode = "duplicate_transaction"
	CodeDatabaseError        ErrorCode = "database_error"
	CodeTransactionFailed    ErrorCode = "transaction_failed"
	CodeInternalError        ErrorCode = "internal_error"
)

type DomainError struct {
	Code    ErrorCode
	Message string
	Cause   error
}

func (e *DomainError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *DomainError) Unwrap() error {
	return e.Cause
}

func (e *DomainError) Is(target error) bool {
	if t, ok := target.(*DomainError); ok {
		return e.Code == t.Code
	}
	return false
}

func NewDomainError(code ErrorCode, message string) *DomainError {
	return &DomainError{Code: code, Message: message}
}

func WrapError(code ErrorCode, message string, cause error) *DomainError {
	return &DomainError{Code: code, Message: message, Cause: cause}
}

var (
	ErrAccountNotFound = &DomainError{
		Code:    CodeAccountNotFound,
		Message: "account not found",
	}
	ErrInsufficientBalance = &DomainError{
		Code:    CodeInsufficientBalance,
		Message: "insufficient balance for this transaction",
	}
	ErrInvalidAmount = &DomainError{
		Code:    CodeInvalidAmount,
		Message: "amount must be a positive decimal value",
	}
	ErrCurrencyMismatch = &DomainError{
		Code:    CodeCurrencyMismatch,
		Message: "currency mismatch between accounts",
	}
	ErrSameAccount = &DomainError{
		Code:    CodeSameAccount,
		Message: "source and destination accounts cannot be the same",
	}
	ErrTransferNotFound = &DomainError{
		Code:    CodeTransferNotFound,
		Message: "transaction not found",
	}
	ErrAccountAlreadyExists = &DomainError{
		Code:    CodeAccountAlreadyExists,
		Message: "account with this ID already exists",
	}
	ErrDuplicateTransaction = &DomainError{
		Code:    CodeDuplicateTransaction,
		Message: "duplicate transaction detected",
	}
)

func IsDomainError(err error) (ErrorCode, bool) {
	var domainErr *DomainError
	if errors.As(err, &domainErr) {
		return domainErr.Code, true
	}
	return "", false
}

func IsRetryable(err error) bool {
	if err == nil {
		return false
	}
	errStr := strings.ToLower(err.Error())
	patterns := []string{"deadlock", "serialize", "connection", "timeout"}
	for _, p := range patterns {
		if strings.Contains(errStr, p) {
			return true
		}
	}
	return false
}
