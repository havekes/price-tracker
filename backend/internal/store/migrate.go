package store

import (
	"context"
	"database/sql"
	"log/slog"
	"time"
)

// Migrate checks if the correspondent table exists and applies the provided
// schema SQL if it doesn't. The operation is idempotent because schemaSQL
// should use CREATE TABLE IF NOT EXISTS.
//
// It reuses the provided *sql.DB pool — does not open its own connection.
// Call this once at startup before the HTTP server starts listening.
func Migrate(db *sql.DB, schemaSQL string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Check if the correspondent table exists.
	var exists bool
	err := db.QueryRowContext(ctx,
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
