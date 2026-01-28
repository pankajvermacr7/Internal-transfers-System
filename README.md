# Internal Transfers System

A Go-based internal transfers service that handles financial transactions between accounts with proper concurrency control.

## Features

- Account creation and balance queries
- Atomic money transfers between accounts
- Row-level locking to prevent race conditions
- Retry logic for transient database errors
- Decimal precision for monetary calculations

## Prerequisites

- Go 1.24+
- Docker & Docker Compose
- PostgreSQL 16+ (or use Docker)

## Quick Start

```bash
# Clone and start services
git clone <repository-url>
cd Internal-transfers-System
make start

# Verify it's running
curl http://localhost:8080/health
```

## Running Locally

```bash
# Start database
docker-compose up -d postgres

# Copy env file
cp .env.example .env

# Run the application
make run
```

## API Endpoints

### Create Account
```bash
curl -X POST http://localhost:8080/api/v1/accounts \
  -H "Content-Type: application/json" \
  -d '{"account_id": 1, "initial_balance": "1000.00"}'
```

### Get Account Balance
```bash
curl http://localhost:8080/api/v1/accounts/1
```

### Transfer Money
```bash
curl -X POST http://localhost:8080/api/v1/transactions \
  -H "Content-Type: application/json" \
  -d '{"source_account_id": 1, "destination_account_id": 2, "amount": "100.00"}'
```

## Testing

```bash
# Unit tests
make test

# Integration tests (requires Docker)
make test-integration

# All tests with coverage
make test-coverage
```

## Project Structure

```
├── cmd/api/              # Application entry point
├── internal/
│   ├── handler/          # HTTP handlers
│   ├── service/          # Business logic
│   ├── repository/       # Data access
│   ├── models/           # Domain models
│   └── server/           # Server setup
└── pkg/config/           # Configuration
```

## Design Decisions

### Deadlock Prevention
Accounts are always locked in consistent order (lower ID first) to prevent deadlocks during concurrent transfers.

### Retry Logic
Transient database errors (deadlocks, serialization failures) trigger automatic retries with exponential backoff.

### Decimal Precision
Uses `shopspring/decimal` for precise monetary calculations instead of floating-point.

### Database Constraints
Business rules enforced at database level:
- `balance >= 0` - No negative balances
- `amount > 0` - Positive transfer amounts only
- `source != destination` - No self-transfers

## Assumptions

1. Account IDs are client-provided (not auto-generated)
2. Single currency - no multi-currency support
3. No authentication - designed for internal use
4. Synchronous processing - no async/queue-based transfers
5. No per-transaction or daily limits
6. Transactions are immutable once created

## Available Commands

| Command | Description |
|---------|-------------|
| `make start` | Start all services |
| `make stop` | Stop services |
| `make run` | Run locally |
| `make test` | Run unit tests |
| `make test-integration` | Run integration tests |
| `make test-coverage` | Generate coverage report |
