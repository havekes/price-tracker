package store

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/havekes/price-tracker/db"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// adminConnStr parses the databaseURL and returns a connection string targeting the
// 'postgres' administrative database (for CREATE/DROP DATABASE).
func adminConnStr(dsn string) (string, error) {
	u, err := url.Parse(dsn)
	if err != nil {
		return "", err
	}
	// Strip leading slash from path, replace with "postgres".
	parts := strings.SplitN(u.Path, "/", 2)
	parts[1] = "postgres"
	u.Path = strings.Join(parts, "/")
	return u.String(), nil
}

// testConnStrDBName extracts the database-name segment from a DSN.
func testConnStrDBName(dsn string) (string, error) {
	u, err := url.Parse(dsn)
	if err != nil {
		return "", err
	}
	parts := strings.SplitN(u.Path, "/", 2)
	if len(parts) < 2 || parts[1] == "" {
		return "", fmt.Errorf("dsn %q has no database name segment", dsn)
	}
	return parts[1], nil
}

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
	connStr := databaseURL()

	// Admin connection to the 'postgres' metadata database for managing throwaway databases.
	adminDSN, err := adminConnStr(connStr)
	if err != nil {
		t.Fatalf("failed to derive admin DSN: %v", err)
	}

	adminDB, err := sql.Open("pgx", adminDSN)
	if err != nil {
		t.Fatalf("failed to open admin connection: %v", err)
	}
	defer adminDB.Close()

	ctx := context.Background()
	if err := adminDB.PingContext(ctx); err != nil {
		t.Skipf("postgres not reachable at %s: %v (docker compose up -d required)", adminDSN, err)
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

	// Connection string pointing to the throwaway database (same host:port:credentials, different db name).
	origDBName, err := testConnStrDBName(connStr)
	if err != nil {
		t.Fatalf("failed to parse database name from DSN: %v", err)
	}
	testConnStr := strings.Replace(connStr, "/"+origDBName+"?", "/"+dbName+"?", 1)

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
