package models

import (
	"errors"
	"fmt"
	"testing"
)

func TestDomainError(t *testing.T) {
	err := NewDomainError(CodeAccountNotFound, "not found")
	if err.Code != CodeAccountNotFound {
		t.Errorf("expected %s, got %s", CodeAccountNotFound, err.Code)
	}

	wrapped := WrapError(CodeDatabaseError, "db failed", fmt.Errorf("conn error"))
	if wrapped.Cause == nil {
		t.Error("expected cause")
	}
	if !errors.Is(wrapped, wrapped.Cause) {
		t.Error("Unwrap should return cause")
	}
}

func TestDomainError_Is(t *testing.T) {
	if !errors.Is(ErrAccountNotFound, &DomainError{Code: CodeAccountNotFound}) {
		t.Error("should match by code")
	}
	if errors.Is(ErrAccountNotFound, &DomainError{Code: CodeInsufficientBalance}) {
		t.Error("should not match different code")
	}
}

func TestIsDomainError(t *testing.T) {
	code, ok := IsDomainError(ErrAccountNotFound)
	if !ok || code != CodeAccountNotFound {
		t.Errorf("expected (CodeAccountNotFound, true), got (%s, %v)", code, ok)
	}

	_, ok = IsDomainError(fmt.Errorf("random"))
	if ok {
		t.Error("should not match non-domain error")
	}
}

func TestIsRetryable(t *testing.T) {
	tests := []struct {
		err  error
		want bool
	}{
		{nil, false},
		{fmt.Errorf("deadlock detected"), true},
		{fmt.Errorf("DEADLOCK DETECTED"), true},
		{fmt.Errorf("could not serialize access"), true},
		{fmt.Errorf("timeout"), true},
		{ErrAccountNotFound, false},
		{fmt.Errorf("random error"), false},
	}

	for _, tt := range tests {
		name := "nil"
		if tt.err != nil {
			name = tt.err.Error()
		}
		t.Run(name, func(t *testing.T) {
			if got := IsRetryable(tt.err); got != tt.want {
				t.Errorf("IsRetryable() = %v, want %v", got, tt.want)
			}
		})
	}
}
