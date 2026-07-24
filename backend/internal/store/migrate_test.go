package store

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"

	"github.com/havekes/price-tracker/db"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// databaseURL returns the DATABASE_URL from the environment or the default dev connection string.
func databaseURL() string {
	if v := os.Getenv("DATABASE_URL"); v != "" {
		return v
	}
	return "postgres://price-tracker:price-tracker@localhost:5433/price-tracker?sslmode=disable"
}

func TestMigrateIdempotent(t *testing.T) {
	connStr := databaseURL()

	sqlDB, err := sql.Open("pgx", connStr)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer sqlDB.Close()

	ctx := context.Background()
	if err := sqlDB.PingContext(ctx); err != nil {
		t.Skipf("postgres not reachable at %s: %v (docker compose up -d required)", connStr, err)
	}

	// Pass 1: apply schema.
	if err := Migrate(sqlDB, db.Schema); err != nil {
		t.Fatalf("migration pass 1 failed: %v", err)
	}

	// Pass 2: apply again — should be idempotent.
	if err := Migrate(sqlDB, db.Schema); err != nil {
		t.Fatalf("migration pass 2 (idempotent) failed: %v", err)
	}
}

func TestMigrateAppliesSchemaOnEmptyTarget(t *testing.T) {
	// Admin connection to the 'postgres' metadata database for managing throwaway databases.
	adminConnStr := "postgres://price-tracker:price-tracker@localhost:5433/postgres?sslmode=disable"

	adminDB, err := sql.Open("pgx", adminConnStr)
	if err != nil {
		t.Fatalf("failed to open admin connection: %v", err)
	}
	defer adminDB.Close()

	ctx := context.Background()
	if err := adminDB.PingContext(ctx); err != nil {
		t.Skipf("postgres not reachable at %s: %v (docker compose up -d required)", adminConnStr, err)
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

	// Connection string pointing to the throwaway database (same host:port:credentials).
	testConnStr := fmt.Sprintf("postgres://price-tracker:price-tracker@localhost:5433/%s?sslmode=disable", dbName)

	testDB, err := sql.Open("pgx", testConnStr)
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}
	defer testDB.Close()

	if err := testDB.PingContext(ctx); err != nil {
		t.Fatalf("test database not reachable: %v", err)
	}

	// Pass 1 — should execute the schema-apply branch.
	if err := Migrate(testDB, db.Schema); err != nil {
		t.Fatalf("first migration (apply) failed: %v", err)
	}

	// Pass 2 — should be idempotent (table now exists).
	if err := Migrate(testDB, db.Schema); err != nil {
		t.Fatalf("second migration (idempotent) failed: %v", err)
	}
}
