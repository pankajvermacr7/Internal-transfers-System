.PHONY: build run test test-integration lint migrate-up migrate-down docker-up docker-down help

# Build variables
BINARY_NAME=api
BUILD_DIR=./bin

# Go variables
GOBASE=$(shell pwd)
GOBIN=$(GOBASE)/bin

## help: Show this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'

## build: Build the application
build:
	@echo "Building..."
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/api

## run: Run the application
run:
	@go run ./cmd/api

## test: Run unit tests only (excludes integration tests)
test:
	@echo "Running unit tests..."
	@go test -v -race -short ./...

## test-unit: Run unit tests with coverage
test-unit:
	@echo "Running unit tests with coverage..."
	@go test -v -race -coverprofile=coverage-unit.out ./internal/... -short
	@go tool cover -func=coverage-unit.out | tail -1

## test-integration: Run integration tests (requires Docker)
test-integration:
	@echo "Running integration tests with testcontainers..."
	@go test -v -race -tags=integration ./internal/repository/... ./internal/service/... -run Integration

## test-all: Run all tests (unit + integration)
test-all:
	@echo "Running all tests..."
	@go test -v -race -tags=integration ./...

## test-coverage: Run all tests with coverage report
test-coverage:
	@echo "Running tests with coverage..."
	@go test -race -coverprofile=coverage.out -covermode=atomic -tags=integration ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"
	@go tool cover -func=coverage.out | tail -1

## test-transfer: Run transfer service tests specifically
test-transfer:
	@echo "Running transfer service tests..."
	@go test -v -race -tags=integration ./internal/service/... -run Transfer

## test-concurrent: Run concurrent/stress tests
test-concurrent:
	@echo "Running concurrent/stress tests..."
	@go test -v -race -tags=integration ./internal/service/... -run "Concurrent|Stress|Race|Deadlock"

## bench: Run benchmarks
bench:
	@echo "Running benchmarks..."
	@go test -bench=. -benchmem ./internal/...

## lint: Run linter
lint:
	@golangci-lint run ./...

## fmt: Format code
fmt:
	@go fmt ./...

## tidy: Tidy dependencies
tidy:
	@go mod tidy

## migrate-up: Run database migrations
migrate-up:
	@goose -dir ./migrations postgres "$(DATABASE_URL)" up

## migrate-down: Rollback last migration
migrate-down:
	@goose -dir ./migrations postgres "$(DATABASE_URL)" down

## migrate-status: Show migration status
migrate-status:
	@goose -dir ./migrations postgres "$(DATABASE_URL)" status

## docker-up: Start Docker containers
docker-up:
	@docker-compose up -d

## docker-down: Stop Docker containers
docker-down:
	@docker-compose down

## docker-logs: Show Docker logs
docker-logs:
	@docker-compose logs -f

## start: Start all services using start-service.sh
start:
	@./start-service.sh start

## stop: Stop all services
stop:
	@./start-service.sh stop

## restart: Restart all services
restart:
	@./start-service.sh restart

## status: Show service status
status:
	@./start-service.sh status

## docker-rebuild: Rebuild containers without cache
docker-rebuild:
	@./start-service.sh rebuild

## docker-cleanup: Stop services and remove volumes
docker-cleanup:
	@./start-service.sh cleanup

## clean: Clean build artifacts
clean:
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
