// Package store provides a Querier interface over sqlc-generated database code.
//
// Phase 3 (ingestion) and Phase 4 (API) depend on the Querier interface for
// testability without requiring a live Postgres instance.
package store

import (
	"database/sql"

	"github.com/havekes/price-tracker/internal/db"
)

// Querier is the interface for all database CRUD operations.
// Embeds the sqlc-generated db.Querier and adds a WithTx helper.
//
// Phase 3/4 should accept Querier in their constructors so they can be tested
// with a mock or a real database interchangeably.
type Querier interface {
	db.Querier

	// WithTx returns a Querier that wraps a single *sql.Tx.
	// All operations on the returned Querier run in the same transaction,
	// enabling atomic multi-step writes.
	WithTx(tx *sql.Tx) Querier
}

// Store implements Querier by wrapping sqlc-generated *db.Queries.
type Store struct {
	*db.Queries
}

// New creates a Store from a DBTX (either *sql.DB or *sql.Tx).
func New(dbtx db.DBTX) *Store {
	return &Store{Queries: db.New(dbtx)}
}

// WithTx returns a Querier that wraps the given transaction.
// This shadows the embedded db.Queries.WithTx to return the store.Querier
// interface instead of *db.Queries.
func (s *Store) WithTx(tx *sql.Tx) Querier {
	return &Store{Queries: s.Queries.WithTx(tx)}
}

// Compile-time check: *Store satisfies Querier.
var _ Querier = (*Store)(nil)
