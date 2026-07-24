package store

import (
	"testing"

	"github.com/havekes/price-tracker/db"
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
