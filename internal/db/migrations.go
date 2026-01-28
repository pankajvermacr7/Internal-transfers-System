package db

import "embed"

// MigrationsFS contains the embedded SQL migration files.
//
//go:embed migrations/*.sql
var MigrationsFS embed.FS

// MigrationsDir is the directory path within the embedded filesystem.
const MigrationsDir = "migrations"
