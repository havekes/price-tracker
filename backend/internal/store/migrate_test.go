package store

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/havekes/price-tracker/db"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestMigrateIdempotent(t *testing.T) {
	connStr := "postgres://price-tracker:price-tracker@localhost:5433/price-tracker?sslmode=disable"

	// Pass 1: apply schema.
	if err := Migrate(connStr, db.Schema); err != nil {
		t.Fatalf("migration pass 1 failed: %v", err)
	}

	// Pass 2: apply again — should be idempotent.
	if err := Migrate(connStr, db.Schema); err != nil {
		t.Fatalf("migration pass 2 (idempotent) failed: %v", err)
	}
}

func TestMigrateAppliesSchemaOnEmptyTarget(t *testing.T) {
	// Connect to the default 'postgres' database to manage throwaway databases.
	adminURL := "postgres://price-tracker:price-tracker@localhost:5433/postgres?sslmode=disable"

	adminDB, err := sql.Open("pgx", adminURL)
	if err != nil {
		t.Fatalf("failed to open admin connection: %v", err)
	}
	defer adminDB.Close()

	ctx := context.Background()
	if err := adminDB.PingContext(ctx); err != nil {
		t.Skipf("postgres not reachable at %s: %v (docker compose up -d required)", adminURL, err)
	}

	// Create a throwaway database so none of the tables exist yet.
	dbName := fmt.Sprintf("test_migrate_%s", randString(8))
	if _, err := adminDB.ExecContext(ctx, fmt.Sprintf("CREATE DATABASE %s", dbName)); err != nil {
		t.Fatalf("failed to create throwaway database: %v", err)
	}
	defer func() {
		if _, err := adminDB.ExecContext(ctx, fmt.Sprintf("DROP DATABASE IF EXISTS %s", dbName)); err != nil {
			t.Logf("WARNING: failed to drop database %s: %v", dbName, err)
		}
	}()

	// Connection string pointing to the throwaway database.
	testConnStr := fmt.Sprintf("postgres://price-tracker:price-tracker@localhost:5433/%s?sslmode=disable", dbName)

	// Pass 1 — should execute the schema-apply branch.
	if err := Migrate(testConnStr, db.Schema); err != nil {
		t.Fatalf("first migration (apply) failed: %v", err)
	}

	// Pass 2 — should be idempotent (table now exists).
	if err := Migrate(testConnStr, db.Schema); err != nil {
		t.Fatalf("second migration (idempotent) failed: %v", err)
	}
}
