package testutil

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

type TestContainerSuite struct {
	container *postgres.PostgresContainer
	pool      *pgxpool.Pool
}

func NewTestContainerSuite() (*TestContainerSuite, error) {
	ctx := context.Background()

	container, err := postgres.Run(ctx, "postgres:16-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("start container: %w", err)
	}

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		container.Terminate(ctx)
		return nil, fmt.Errorf("connection string: %w", err)
	}

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		container.Terminate(ctx)
		return nil, fmt.Errorf("pool: %w", err)
	}

	suite := &TestContainerSuite{container: container, pool: pool}
	if err := suite.runMigrations(ctx); err != nil {
		suite.Teardown()
		return nil, err
	}

	return suite, nil
}

func (s *TestContainerSuite) Pool() *pgxpool.Pool {
	return s.pool
}

func (s *TestContainerSuite) Clean() error {
	_, err := s.pool.Exec(context.Background(), `
		TRUNCATE transactions RESTART IDENTITY CASCADE;
		TRUNCATE accounts RESTART IDENTITY CASCADE;
	`)
	return err
}

func (s *TestContainerSuite) Teardown() {
	if s.pool != nil {
		s.pool.Close()
	}
	if s.container != nil {
		s.container.Terminate(context.Background())
	}
}

func (s *TestContainerSuite) runMigrations(ctx context.Context) error {
	_, err := s.pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS accounts (
			account_id BIGINT PRIMARY KEY,
			balance NUMERIC NOT NULL CHECK (balance >= 0),
			created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
		);
		
		CREATE OR REPLACE FUNCTION set_updated_at() RETURNS TRIGGER AS $$
		BEGIN NEW.updated_at = now(); RETURN NEW; END;
		$$ LANGUAGE plpgsql;
		
		DROP TRIGGER IF EXISTS trg_accounts_updated ON accounts;
		CREATE TRIGGER trg_accounts_updated BEFORE UPDATE ON accounts
		FOR EACH ROW EXECUTE FUNCTION set_updated_at();
		
		CREATE TABLE IF NOT EXISTS transactions (
			transaction_id BIGSERIAL PRIMARY KEY,
			source_account_id BIGINT NOT NULL REFERENCES accounts(account_id),
			destination_account_id BIGINT NOT NULL REFERENCES accounts(account_id),
			amount NUMERIC NOT NULL CHECK (amount > 0),
			created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			CHECK (source_account_id <> destination_account_id)
		);
		
		CREATE INDEX IF NOT EXISTS idx_txn_source ON transactions(source_account_id, created_at DESC);
		CREATE INDEX IF NOT EXISTS idx_txn_dest ON transactions(destination_account_id, created_at DESC);
	`)
	return err
}
