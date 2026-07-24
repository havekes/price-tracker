package store

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/havekes/price-tracker/internal/db"
)

// TestWithTx verifies that the WithTx helper enables atomic multi-step writes
// within a single transaction (commit case) and that rollback discards changes.
func TestWithTx(t *testing.T) {
	t.Parallel()

	database, cleanup := testDB(t)
	defer cleanup()

	q := New(database)
	ctx := context.Background()

	t.Run("commit persists changes", func(t *testing.T) {
		tx, err := database.BeginTx(ctx, nil)
		if err != nil {
			t.Fatalf("BeginTx: %v", err)
		}
		defer tx.Rollback()

		txQ := q.WithTx(tx)

		// Create a correspondent within the transaction.
		c, err := txQ.CreateCorrespondent(ctx, "Tx Alice")
		if err != nil {
			t.Fatalf("CreateCorrespondent in tx: %v", err)
		}

		if err := tx.Commit(); err != nil {
			t.Fatalf("Commit: %v", err)
		}

		// Verify the correspondent exists after commit.
		got, err := q.GetCorrespondent(ctx, c.ID)
		if err != nil {
			t.Fatalf("GetCorrespondent after commit: %v", err)
		}
		if got.Name != "Tx Alice" {
			t.Fatalf("expected Tx Alice, got %s", got.Name)
		}
	})

	t.Run("rollback discards changes", func(t *testing.T) {
		tx, err := database.BeginTx(ctx, nil)
		if err != nil {
			t.Fatalf("BeginTx: %v", err)
		}

		txQ := q.WithTx(tx)

		c, err := txQ.CreateCorrespondent(ctx, "Tx Bob")
		if err != nil {
			t.Fatalf("CreateCorrespondent in tx: %v", err)
		}

		if err := tx.Rollback(); err != nil {
			t.Fatalf("Rollback: %v", err)
		}

		// Verify the correspondent does NOT exist after rollback.
		_, err = q.GetCorrespondent(ctx, c.ID)
		if err == nil {
			t.Fatal("expected error — correspondent should not exist after rollback")
		}
	})

	t.Run("multi-step atomic write", func(t *testing.T) {
		// Simulate Phase 3 atomic ingestion: create correspondant + receipt + raw_item
		// in a single transaction.
		tx, err := database.BeginTx(ctx, nil)
		if err != nil {
			t.Fatalf("BeginTx: %v", err)
		}
		defer tx.Rollback()

		txQ := q.WithTx(tx)

		corr, err := txQ.CreateCorrespondent(ctx, "Atomic Test")
		if err != nil {
			t.Fatalf("CreateCorrespondent: %v", err)
		}

		receipt, err := txQ.CreateReceipt(ctx, db.CreateReceiptParams{
			CorrespondentID: corr.ID,
			PurchasedAt:     time.Date(2026, 7, 22, 0, 0, 0, 0, time.UTC),
			Source:          "upload",
			ExternalDocID:   sql.NullString{},
			RawFileRef:      sql.NullString{},
		})
		if err != nil {
			t.Fatalf("CreateReceipt: %v", err)
		}

		item, err := txQ.CreateRawItem(ctx, db.CreateRawItemParams{
			ReceiptID:     receipt.ID,
			ProductID:     sql.NullInt64{},
			RawText:       "Atomic Item",
			RawQuantity:   sql.NullString{},
			QuantityValue: sql.NullFloat64{},
			QuantityUnit:  sql.NullString{},
		})
		if err != nil {
			t.Fatalf("CreateRawItem: %v", err)
		}

		_, err = txQ.CreatePriceRecord(ctx, db.CreatePriceRecordParams{
			RawItemID:  item.ID,
			TotalPrice: 9.99,
			UnitPrice:  sql.NullFloat64{},
			Currency:   "USD",
		})
		if err != nil {
			t.Fatalf("CreatePriceRecord: %v", err)
		}

		if err := tx.Commit(); err != nil {
			t.Fatalf("Commit: %v", err)
		}

		// Verify all entities exist after commit.
		if _, err := q.GetCorrespondent(ctx, corr.ID); err != nil {
			t.Fatalf("correspondent should exist: %v", err)
		}
		if _, err := q.GetReceipt(ctx, receipt.ID); err != nil {
			t.Fatalf("receipt should exist: %v", err)
		}
		if _, err := q.GetRawItem(ctx, item.ID); err != nil {
			t.Fatalf("raw_item should exist: %v", err)
		}
	})
}
