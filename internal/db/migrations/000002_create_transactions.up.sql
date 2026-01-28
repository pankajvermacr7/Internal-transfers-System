CREATE TABLE IF NOT EXISTS transactions (
  transaction_id         BIGSERIAL PRIMARY KEY,
  source_account_id      BIGINT NOT NULL REFERENCES accounts(account_id) ON DELETE RESTRICT,
  destination_account_id BIGINT NOT NULL REFERENCES accounts(account_id) ON DELETE RESTRICT,
  amount                 NUMERIC NOT NULL CHECK (amount > 0),
  created_at             TIMESTAMPTZ NOT NULL DEFAULT now(),
  CHECK (source_account_id <> destination_account_id)
);

CREATE INDEX IF NOT EXISTS idx_transactions_source_created_at
  ON transactions (source_account_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_transactions_destination_created_at
  ON transactions (destination_account_id, created_at DESC);
