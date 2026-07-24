package store

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// Migrate connects to the database using DATABASE_URL and applies the provided
// schema SQL if the correspondent table does not exist. The operation is
// idempotent because schemaSQL should use CREATE TABLE IF NOT EXISTS.
//
// Call this once at startup before the HTTP server starts listening.
func Migrate(databaseURL, schemaSQL string) error {
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return err
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Check if the correspondent table exists.
	var exists bool
	err = db.QueryRowContext(ctx,
		`SELECT EXISTS (
			SELECT 1 FROM information_schema.tables
			WHERE table_schema = 'public' AND table_name = 'correspondent'
		)`,
	).Scan(&exists)
	if err != nil {
		return err
	}

	if exists {
		slog.Info("database schema already exists, skipping migration")
		return nil
	}

	slog.Info("applying database schema...")
	if _, err := db.ExecContext(ctx, schemaSQL); err != nil {
		return err
	}
	slog.Info("database schema applied successfully")
	return nil
}
