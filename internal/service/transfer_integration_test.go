//go:build integration

package service

import (
	"context"
	"errors"
	"os"
	"sync"
	"sync/atomic"
	"testing"

	"internal-transfers-system/internal/models"
	"internal-transfers-system/internal/repository"
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

func setup(t *testing.T) (*TransferService, *AccountService, *repository.AccountRepository) {
	t.Helper()
	if err := testSuite.Clean(); err != nil {
		t.Fatalf("clean: %v", err)
	}
	accRepo := repository.NewAccountRepository(testSuite.Pool())
	txnRepo := repository.NewTransactionRepository(testSuite.Pool())
	return NewTransferService(accRepo, txnRepo), NewAccountService(accRepo), accRepo
}

func createAccount(t *testing.T, svc *AccountService, id int64, balance string) {
	t.Helper()
	_, err := svc.CreateAccount(context.Background(), &models.CreateAccountRequest{
		AccountID: id, InitialBalance: balance,
	})
	if err != nil {
		t.Fatalf("create account: %v", err)
	}
}

func TestIntegration_BasicTransfer(t *testing.T) {
	transferSvc, accSvc, accRepo := setup(t)
	ctx := context.Background()

	createAccount(t, accSvc, 1, "1000")
	createAccount(t, accSvc, 2, "500")

	txn, err := transferSvc.Transfer(ctx, &models.CreateTransactionRequest{
		SourceAccountID: 1, DestinationAccountID: 2, Amount: "100",
	})
	if err != nil {
		t.Fatalf("transfer: %v", err)
	}
	if txn.TransactionID == 0 {
		t.Error("expected transaction ID")
	}

	acc1, _ := accRepo.GetByID(ctx, 1)
	acc2, _ := accRepo.GetByID(ctx, 2)
	if !acc1.Balance.Equal(decimal.NewFromInt(900)) {
		t.Errorf("source balance: got %s", acc1.Balance)
	}
	if !acc2.Balance.Equal(decimal.NewFromInt(600)) {
		t.Errorf("dest balance: got %s", acc2.Balance)
	}
}

func TestIntegration_InsufficientBalance(t *testing.T) {
	transferSvc, accSvc, _ := setup(t)

	createAccount(t, accSvc, 1, "100")
	createAccount(t, accSvc, 2, "500")

	_, err := transferSvc.Transfer(context.Background(), &models.CreateTransactionRequest{
		SourceAccountID: 1, DestinationAccountID: 2, Amount: "200",
	})
	if !errors.Is(err, models.ErrInsufficientBalance) {
		t.Errorf("expected insufficient balance, got %v", err)
	}
}

func TestIntegration_ConcurrentTransfers(t *testing.T) {
	transferSvc, accSvc, accRepo := setup(t)
	ctx := context.Background()

	createAccount(t, accSvc, 1, "10000")
	createAccount(t, accSvc, 2, "10000")

	var wg sync.WaitGroup
	var success atomic.Int32

	// 50 transfers each direction
	for i := 0; i < 50; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			_, err := transferSvc.Transfer(ctx, &models.CreateTransactionRequest{
				SourceAccountID: 1, DestinationAccountID: 2, Amount: "10",
			})
			if err == nil {
				success.Add(1)
			}
		}()
		go func() {
			defer wg.Done()
			_, err := transferSvc.Transfer(ctx, &models.CreateTransactionRequest{
				SourceAccountID: 2, DestinationAccountID: 1, Amount: "10",
			})
			if err == nil {
				success.Add(1)
			}
		}()
	}
	wg.Wait()

	t.Logf("successful transfers: %d", success.Load())

	// Total balance must be conserved
	acc1, _ := accRepo.GetByID(ctx, 1)
	acc2, _ := accRepo.GetByID(ctx, 2)
	total := acc1.Balance.Add(acc2.Balance)
	if !total.Equal(decimal.NewFromInt(20000)) {
		t.Errorf("balance mismatch: %s + %s = %s", acc1.Balance, acc2.Balance, total)
	}
}

func TestIntegration_RaceForSameBalance(t *testing.T) {
	transferSvc, accSvc, accRepo := setup(t)
	ctx := context.Background()

	createAccount(t, accSvc, 1, "100")
	createAccount(t, accSvc, 2, "0")

	var wg sync.WaitGroup
	var success atomic.Int32

	// 20 goroutines try to transfer entire balance
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := transferSvc.Transfer(ctx, &models.CreateTransactionRequest{
				SourceAccountID: 1, DestinationAccountID: 2, Amount: "100",
			})
			if err == nil {
				success.Add(1)
			}
		}()
	}
	wg.Wait()

	// Exactly one should succeed
	if success.Load() != 1 {
		t.Errorf("expected 1 success, got %d", success.Load())
	}

	acc1, _ := accRepo.GetByID(ctx, 1)
	acc2, _ := accRepo.GetByID(ctx, 2)
	if !acc1.Balance.IsZero() || !acc2.Balance.Equal(decimal.NewFromInt(100)) {
		t.Errorf("unexpected balances: %s, %s", acc1.Balance, acc2.Balance)
	}
}
