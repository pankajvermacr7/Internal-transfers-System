// Package models provides domain models for the internal transfers system.
// These models represent the core business entities and are used across
// all layers of the application.
package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// Account represents a bank account in the system.
//
// Business rules:
//   - AccountID is provided by the client and must be unique
//   - Balance cannot be negative (enforced at database level)
//   - All monetary operations use decimal.Decimal for precision
//
// Uses db tags for go-kit/pgx reflection-based CRUD operations.
type Account struct {
	// AccountID is the unique identifier for the account.
	// This is provided by the client during account creation.
	AccountID int64 `db:"account_id" id:"true" json:"account_id"`

	// Balance is the current balance of the account.
	// Uses decimal.Decimal for precise monetary calculations.
	Balance decimal.Decimal `db:"balance" json:"balance"`

	// CreatedAt is the timestamp when the account was created.
	CreatedAt time.Time `db:"created_at" json:"created_at"`

	// UpdatedAt is the timestamp when the account was last updated.
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// TableName returns the database table name for Account.
// This can be used by go-kit/pgx for table resolution.
func (a Account) TableName() string {
	return "accounts"
}
