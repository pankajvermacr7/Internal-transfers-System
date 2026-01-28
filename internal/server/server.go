// Package server provides the HTTP server implementation for the internal transfers API.
// It handles server lifecycle, routing, and middleware configuration.
package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"internal-transfers-system/internal/handler"
	"internal-transfers-system/internal/repository"
	"internal-transfers-system/internal/service"
	config "internal-transfers-system/pkg/config"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

// Server represents the HTTP server for the internal transfers API.
// It manages the server lifecycle, routing, and dependency injection.
type Server struct {
	httpServer *http.Server
	router     *http.ServeMux
	db         *pgxpool.Pool

	// Handlers for different API endpoints
	accountHandler     *handler.AccountHandler
	transactionHandler *handler.TransactionHandler
}

// New creates a new Server instance with all dependencies configured.
// It sets up:
//   - Database repositories
//   - Business logic services
//   - HTTP handlers
//   - Middleware chain (recovery, request ID, logging)
//   - Route registration
func New(cfg config.ServerConfig, db *pgxpool.Pool) *Server {
	router := http.NewServeMux()

	// Create repositories (data access layer)
	accountRepo := repository.NewAccountRepository(db)
	transactionRepo := repository.NewTransactionRepository(db)

	// Create services (business logic layer)
	accountService := service.NewAccountService(accountRepo)
	transferService := service.NewTransferService(accountRepo, transactionRepo)

	// Create handlers (presentation layer)
	accountHandler := handler.NewAccountHandler(accountService)
	transactionHandler := handler.NewTransactionHandler(transferService)

	srv := &Server{
		router: router,
		db:     db,
		httpServer: &http.Server{
			Addr:         cfg.Address(),
			ReadTimeout:  cfg.ReadTimeout,
			WriteTimeout: cfg.WriteTimeout,
			IdleTimeout:  cfg.IdleTimeout,
		},
		accountHandler:     accountHandler,
		transactionHandler: transactionHandler,
	}

	// Register routes with handlers
	srv.registerRoutes()

	// Apply middleware chain (order matters: outermost first)
	// Recovery -> RequestID -> Logging -> Router
	handler := RecoveryMiddleware(
		RequestIDMiddleware(
			LoggingMiddleware(router),
		),
	)
	srv.httpServer.Handler = handler

	return srv
}

// registerRoutes sets up all HTTP routes for the API.
// Routes are organized by resource type and versioned under /api/v1.
func (s *Server) registerRoutes() {
	// Health check endpoints (no versioning for infrastructure endpoints)
	s.router.HandleFunc("GET /health", s.handleHealth)
	s.router.HandleFunc("GET /ready", s.handleReady)

	// Account endpoints
	// POST /api/v1/accounts - Create a new account
	// GET /api/v1/accounts/{id} - Get account details
	s.router.HandleFunc("POST /api/v1/accounts", s.accountHandler.CreateAccount)
	s.router.HandleFunc("GET /api/v1/accounts/{id}", s.accountHandler.GetAccount)

	// Transaction endpoints
	// POST /api/v1/transactions - Create a money transfer
	s.router.HandleFunc("POST /api/v1/transactions", s.transactionHandler.CreateTransaction)
}

// Start begins listening for HTTP requests.
func (s *Server) Start() error {
	log.Info().
		Str("address", s.httpServer.Addr).
		Msg("Starting HTTP server")

	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("HTTP server error: %w", err)
	}

	return nil
}

// Shutdown gracefully stops the HTTP server.
func (s *Server) Shutdown(ctx context.Context) error {
	log.Info().Msg("Shutting down HTTP server...")

	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown error: %w", err)
	}

	log.Info().Msg("HTTP server stopped")
	return nil
}

// GracefulShutdown waits for the given duration before forcing shutdown.
func (s *Server) GracefulShutdown(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return s.Shutdown(ctx)
}
