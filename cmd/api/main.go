package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"internal-transfers-system/internal/server"
	config "internal-transfers-system/pkg/config"

	"github.com/pankajvermacr7/go-kit/logging"
	"github.com/pankajvermacr7/go-kit/pgx"
	"github.com/rs/zerolog/log"
)

func init() {
	logging.InitLogger()
}

func main() {
	log.Info().Msg("Starting Internal Transfers System...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	log.Info().
		Str("server_address", cfg.Server.Address()).
		Str("db_host", cfg.Database.Host).
		Int("db_port", cfg.Database.Port).
		Str("db_name", cfg.Database.Database).
		Str("log_level", cfg.Log.Level).
		Msg("Configuration loaded successfully")

	// Connect to database using go-kit
	db, err := pgx.NewDB(cfg.Database.ToPgxConfig())
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer db.Close()
	log.Info().Msg("Database connection established")

	// Run migrations
	if err := db.RunMigrationsFromDir(cfg.Database.MigrationsPath); err != nil {
		log.Fatal().Err(err).Msg("Failed to run migrations")
	}
	log.Info().Str("path", cfg.Database.MigrationsPath).Msg("Database migrations applied")

	// Create HTTP server
	srv := server.New(cfg.Server, db.GetPool())

	// Channel to listen for errors from server
	serverErrors := make(chan error, 1)

	// Start server in a goroutine
	go func() {
		serverErrors <- srv.Start()
	}()

	// Channel to listen for OS signals
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Block until we receive a signal or server error
	select {
	case err := <-serverErrors:
		log.Fatal().Err(err).Msg("Server error")

	case sig := <-shutdown:
		log.Info().Str("signal", sig.String()).Msg("Shutdown signal received")

		// Give outstanding requests 30 seconds to complete
		if err := srv.GracefulShutdown(30 * time.Second); err != nil {
			log.Error().Err(err).Msg("Graceful shutdown failed")
			os.Exit(1)
		}
	}

	log.Info().Msg("Server stopped")
}
