//go:build integration

package repository

import (
	"context"
	"os"
	"testing"
	"time"

	"internal-transfers-system/internal/models"
	"internal-transfers-system/internal/testutil"

	"github.com/shopspring/decimal"
)

var testSuite *testutil.TestContainerSuite

func TestMain(m *testing.M) {
	var err error
	testSuite, err = testutil.NewTestContainerSuite()
	if err != nil {
		panic("failed to start test container: " + err.Error())
	}
	code := m.Run()
	testSuite.Teardown()
	os.Exit(code)
}

func setupAccountRepo(t *testing.T) *AccountRepository {
	t.Helper()
	if err := testSuite.Clean(); err != nil {
		t.Fatalf("clean: %v", err)
	}
	return NewAccountRepository(testSuite.Pool())
}

func TestAccountRepository_Create(t *testing.T) {
	repo := setupAccountRepo(t)
	ctx := context.Background()

	acc := &models.Account{AccountID: 1, Balance: decimal.NewFromInt(1000)}
	if err := repo.Create(ctx, acc); err != nil {
		t.Fatalf("create: %v", err)
	}
	if acc.CreatedAt.IsZero() {
		t.Error("expected CreatedAt")
	}
}

func TestAccountRepository_Create_Duplicate(t *testing.T) {
	repo := setupAccountRepo(t)
	ctx := context.Background()

	repo.Create(ctx, &models.Account{AccountID: 1, Balance: decimal.NewFromInt(1000)})
	err := repo.Create(ctx, &models.Account{AccountID: 1, Balance: decimal.NewFromInt(500)})
	if err == nil {
		t.Error("expected duplicate error")
	}
}

func TestAccountRepository_GetByID(t *testing.T) {
	repo := setupAccountRepo(t)
	ctx := context.Background()

	repo.Create(ctx, &models.Account{AccountID: 1, Balance: decimal.NewFromInt(1000)})

	acc, err := repo.GetByID(ctx, 1)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if !acc.Balance.Equal(decimal.NewFromInt(1000)) {
		t.Errorf("expected 1000, got %s", acc.Balance)
	}
}

func TestAccountRepository_GetByID_NotFound(t *testing.T) {
	repo := setupAccountRepo(t)
	_, err := repo.GetByID(context.Background(), 999)
	if err != models.ErrAccountNotFound {
		t.Errorf("expected ErrAccountNotFound, got %v", err)
	}
}

func TestAccountRepository_UpdateBalance(t *testing.T) {
	repo := setupAccountRepo(t)
	ctx := context.Background()

	repo.Create(ctx, &models.Account{AccountID: 1, Balance: decimal.NewFromInt(1000)})

	tx, _ := repo.BeginTx(ctx)
	repo.UpdateBalance(ctx, tx, 1, decimal.NewFromInt(500))
	tx.Commit(ctx)

	acc, _ := repo.GetByID(ctx, 1)
	if !acc.Balance.Equal(decimal.NewFromInt(500)) {
		t.Errorf("expected 500, got %s", acc.Balance)
	}
}

func TestAccountRepository_GetByIDForUpdate_Locking(t *testing.T) {
	repo := setupAccountRepo(t)
	ctx := context.Background()

	repo.Create(ctx, &models.Account{AccountID: 1, Balance: decimal.NewFromInt(1000)})

	tx1, _ := repo.BeginTx(ctx)
	defer tx1.Rollback(ctx)
	repo.GetByIDForUpdate(ctx, tx1, 1)

	// Second tx should block
	lockAcquired := make(chan bool)
	go func() {
		tx2, _ := repo.BeginTx(ctx)
		defer tx2.Rollback(ctx)
		repo.GetByIDForUpdate(ctx, tx2, 1)
		lockAcquired <- true
	}()

	select {
	case <-lockAcquired:
		t.Error("second tx should block")
	case <-time.After(100 * time.Millisecond):
		// expected
	}

	tx1.Rollback(ctx)

	select {
	case <-lockAcquired:
		// expected
	case <-time.After(2 * time.Second):
		t.Error("second tx should acquire lock")
	}
}

func TestAccountRepository_Exists(t *testing.T) {
	repo := setupAccountRepo(t)
	ctx := context.Background()

	exists, _ := repo.Exists(ctx, 1)
	if exists {
		t.Error("should not exist")
	}

	repo.Create(ctx, &models.Account{AccountID: 1, Balance: decimal.NewFromInt(1000)})

	exists, _ = repo.Exists(ctx, 1)
	if !exists {
		t.Error("should exist")
	}
}
