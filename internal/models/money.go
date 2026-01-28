package models

import (
	"fmt"

	"github.com/shopspring/decimal"
)

func ParseMoney(s string) (decimal.Decimal, error) {
	if s == "" {
		return decimal.Decimal{}, fmt.Errorf("empty amount string")
	}
	d, err := decimal.NewFromString(s)
	if err != nil {
		return decimal.Decimal{}, fmt.Errorf("invalid decimal %q: %w", s, err)
	}
	return d, nil
}

func FormatMoney(d decimal.Decimal) string {
	return d.String()
}
