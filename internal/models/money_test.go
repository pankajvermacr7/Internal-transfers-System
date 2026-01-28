package models

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestParseMoney(t *testing.T) {
	tests := []struct {
		input   string
		want    string
		wantErr bool
	}{
		{"100", "100", false},
		{"100.50", "100.5", false},
		{"0", "0", false},
		{"0.01", "0.01", false},
		{"-100", "-100", false},
		{"", "", true},
		{"abc", "", true},
		{"$100", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := ParseMoney(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("wantErr=%v, got %v", tt.wantErr, err)
				return
			}
			if !tt.wantErr {
				expected, _ := decimal.NewFromString(tt.want)
				if !result.Equal(expected) {
					t.Errorf("expected %s, got %s", tt.want, result)
				}
			}
		})
	}
}

func TestFormatMoney(t *testing.T) {
	tests := []struct {
		input decimal.Decimal
		want  string
	}{
		{decimal.NewFromInt(100), "100"},
		{decimal.NewFromFloat(100.50), "100.5"},
		{decimal.Zero, "0"},
	}

	for _, tt := range tests {
		if got := FormatMoney(tt.input); got != tt.want {
			t.Errorf("FormatMoney(%v) = %s, want %s", tt.input, got, tt.want)
		}
	}
}
