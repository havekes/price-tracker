package ingest

import (
	"context"
	"crypto/rand"
	"database/sql"
	"fmt"
	"log"
	"math/big"
	"os"
	"testing"

	"github.com/havekes/price-tracker/db"
)

// testDB returns a *sql.DB connected to a dedicated test schema.
// It creates the test schema and applies schema.sql, and registers a cleanup
// that drops the test schema.
//
// Requires a running Postgres at DATABASE_URL (set via env or the default
// dev config). Tests will be skipped if the connection fails.
func testDB(t *testing.T) (*sql.DB, func()) {
	t.Helper()

	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "postgres://price-tracker:price-tracker@localhost:5433/price-tracker?sslmode=disable"
	}

	rootDB, err := sql.Open("pgx", connStr)
	if err != nil {
		t.Fatalf("failed to open root connection: %v", err)
	}

	ctx := context.Background()
	if err := rootDB.PingContext(ctx); err != nil {
		rootDB.Close()
		t.Skipf("postgres not reachable at %s: %v (docker compose up -d required)", connStr, err)
	}

	schemaName := fmt.Sprintf("test_%s", randString(8))

	_, err = rootDB.ExecContext(ctx, fmt.Sprintf(`CREATE SCHEMA IF NOT EXISTS %s`, schemaName))
	if err != nil {
		rootDB.Close()
		t.Fatalf("failed to create test schema %s: %v", schemaName, err)
	}

	// Switch to the test schema via search_path.
	testConnStr := connStr + fmt.Sprintf("&search_path=%s,public", schemaName)
	testDB, err := sql.Open("pgx", testConnStr)
	if err != nil {
		rootDB.ExecContext(ctx, fmt.Sprintf(`DROP SCHEMA %s CASCADE`, schemaName))
		rootDB.Close()
		t.Fatalf("failed to open test connection: %v", err)
	}

	// Apply schema in the test schema.
	if _, err := testDB.ExecContext(ctx, db.Schema); err != nil {
		testDB.Close()
		rootDB.ExecContext(ctx, fmt.Sprintf(`DROP SCHEMA %s CASCADE`, schemaName))
		rootDB.Close()
		t.Fatalf("failed to apply schema: %v", err)
	}

	cleanup := func() {
		testDB.Close()
		if _, err := rootDB.ExecContext(ctx, fmt.Sprintf(`DROP SCHEMA %s CASCADE`, schemaName)); err != nil {
			log.Printf("WARNING: failed to drop test schema %s: %v", schemaName, err)
		}
		rootDB.Close()
	}

	return testDB, cleanup
}

// randString generates a random alphanumeric string of length n using crypto/rand.
func randString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			// Fallback — should never happen.
			b[i] = 'x'
			continue
		}
		b[i] = letters[idx.Int64()]
	}
	return string(b)
}
