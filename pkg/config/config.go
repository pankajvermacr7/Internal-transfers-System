package pkg

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/pankajvermacr7/go-kit/pgx"
)

// Config holds all configuration for the application.
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Log      LogConfig
}

// ServerConfig holds HTTP server configuration.
type ServerConfig struct {
	Host         string        `envconfig:"SERVER_HOST" default:"0.0.0.0"`
	Port         int           `envconfig:"SERVER_PORT" default:"8080"`
	ReadTimeout  time.Duration `envconfig:"SERVER_READ_TIMEOUT" default:"15s"`
	WriteTimeout time.Duration `envconfig:"SERVER_WRITE_TIMEOUT" default:"15s"`
	IdleTimeout  time.Duration `envconfig:"SERVER_IDLE_TIMEOUT" default:"60s"`
}

// Address returns the server address in host:port format.
func (s ServerConfig) Address() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

// DatabaseConfig holds database configuration aligned with go-kit/pgx.Config.
type DatabaseConfig struct {
	Host           string        `envconfig:"DB_HOST" default:"localhost"`
	Port           int           `envconfig:"DB_PORT" default:"5432"`
	Username       string        `envconfig:"DB_USERNAME" default:"postgres"`
	Password       string        `envconfig:"DB_PASSWORD" default:"postgres"`
	Database       string        `envconfig:"DB_DATABASE" default:"transfers"`
	SSLMode        string        `envconfig:"DB_SSL_MODE" default:"disable"`
	MaxConns       int           `envconfig:"DB_MAX_CONNS" default:"10"`
	Timeout        time.Duration `envconfig:"DB_TIMEOUT" default:"5s"`
	MigrationsPath string        `envconfig:"DB_MIGRATIONS_PATH" default:"migrations"`
}

// ToPgxConfig converts DatabaseConfig to go-kit/pgx.Config.
func (d DatabaseConfig) ToPgxConfig() pgx.Config {
	return pgx.Config{
		Host:     d.Host,
		Port:     d.Port,
		Username: d.Username,
		Password: d.Password,
		Database: d.Database,
		SSLMode:  d.SSLMode,
		MaxConns: d.MaxConns,
		Timeout:  d.Timeout,
	}
}

// DSN returns the database connection string.
func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		d.Username, d.Password, d.Host, d.Port, d.Database, d.SSLMode,
	)
}

// LogConfig holds logging configuration.
type LogConfig struct {
	Level  string `envconfig:"LOG_LEVEL" default:"info"`
	Format string `envconfig:"LOG_FORMAT" default:"json"` // json or console
}

// Load loads configuration from environment variables.
func Load() (*Config, error) {
	var cfg Config

	if err := envconfig.Process("", &cfg.Server); err != nil {
		return nil, fmt.Errorf("loading server config: %w", err)
	}

	if err := envconfig.Process("", &cfg.Database); err != nil {
		return nil, fmt.Errorf("loading database config: %w", err)
	}

	if err := envconfig.Process("", &cfg.Log); err != nil {
		return nil, fmt.Errorf("loading log config: %w", err)
	}

	return &cfg, nil
}
