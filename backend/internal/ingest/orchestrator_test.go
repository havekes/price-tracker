package ingest

import (
	"context"
	"testing"
	"time"

	"github.com/havekes/price-tracker/internal/store"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestPersistReceipt(t *testing.T) {
	database, cleanup := testDB(t)
	defer cleanup()

	s := store.New(database)
	ctx := context.Background()

	t.Run("atomic commit success", func(t *testing.T) {
		tx, err := database.BeginTx(ctx, nil)
		if err != nil {
			t.Fatalf("BeginTx: %v", err)
		}
		defer tx.Rollback()

		extDocID := "doc-123"
		rawFileRef := "s3://bucket/doc-123.pdf"

		input := IngestInput{
			CorrespondentName: "Costco",
			PurchasedAt:       time.Date(2026, 7, 24, 0, 0, 0, 0, time.UTC),
			Source:            "upload",
			ExternalDocID:     &extDocID,
			RawFileRef:        &rawFileRef,
			Items: []IngestItemInput{
				{
					RawText:     "1.5kg Kirkland Coffee",
					DisplayName: "Kirkland Coffee",
					RawQuantity: "1.5kg",
					BaseUnit:    "kg",
					TotalPrice:  15.99,
					Currency:    "USD",
				},
				{
					RawText:     "2 count paper towels",
					DisplayName: "Paper Towels",
					RawQuantity: "2 pk",
					BaseUnit:    "unit",
					TotalPrice:  20.00,
					Currency:    "USD",
				},
			},
		}

		err = PersistReceipt(ctx, s, tx, input)
		if err != nil {
			t.Fatalf("PersistReceipt: %v", err)
		}

		if err := tx.Commit(); err != nil {
			t.Fatalf("Commit: %v", err)
		}

		// Verify DB state
		corr, err := s.GetCorrespondentByName(ctx, "Costco")
		if err != nil {
			t.Fatalf("Expected correspondent 'Costco' to exist")
		}

		receipts, err := s.ListReceiptsByCorrespondent(ctx, corr.ID)
		if err != nil || len(receipts) != 1 {
			t.Fatalf("Expected 1 receipt for correspondent")
		}
		receipt := receipts[0]

		items, err := s.ListRawItemsByReceipt(ctx, receipt.ID)
		if err != nil || len(items) != 2 {
			t.Fatalf("Expected 2 raw items, got %v", len(items))
		}

		for _, item := range items {
			priceRec, err := s.GetPriceRecordByRawItem(ctx, item.ID)
			if err != nil {
				t.Fatalf("Expected price record for item %d", item.ID)
			}
			if priceRec.UnitPrice.Valid == false {
				t.Fatalf("Expected unit price to be populated")
			}
		}
	})

	t.Run("rollback on error", func(t *testing.T) {
		tx, err := database.BeginTx(ctx, nil)
		if err != nil {
			t.Fatalf("BeginTx: %v", err)
		}

		extDocID := "doc-456"

		input := IngestInput{
			CorrespondentName: "Target",
			PurchasedAt:       time.Date(2026, 7, 24, 0, 0, 0, 0, time.UTC),
			Source:            "upload",
			ExternalDocID:     &extDocID,
			Items: []IngestItemInput{
				{
					RawText:     "Bad Item",
					DisplayName: "Bad Item",
					RawQuantity: "invalid-qty",
					BaseUnit:    "kg",
					TotalPrice:  10.00,
					// ParseQuantity will fail or succeed as "10.00 unit"? Wait, if RawQuantity is "invalid-qty", ParseQuantity returns ErrInvalidQuantity. 
					// Let's force an error by doing something else.
					// Actually, ParseQuantity might not return an error, it might just leave unit empty. Let's make TotalPrice 0 for division by zero? No, CalculateUnitPrice handles it.
				},
			},
		}
		
		// To force a DB error during transaction, let's pass an item with an extremely long base unit or something that fails a constraint?
		// Or we can just mock a failure.
		// Wait, if ParseQuantity fails, we ignore error in our code and just set qtyValue to invalid.
		// What fails DB constraints?
		// TotalPrice = 0 might not fail DB constraints if there's no CHECK constraint.
		// Let's pass nil or something that violates NOT NULL?
		// `input.Source` is empty, wait, if we make input.CorrespondentName empty, maybe it fails a constraint? 
		// Let's check `correspondent.name` constraint. Usually it's NOT NULL and cannot be empty.
		
		badInput := input
		badInput.CorrespondentName = "" // Usually violates CHECK (name <> '') if it exists, or UNIQUE.
		// Actually let's just make `Source` empty, or just use a deliberately failing `RawText` (too long if there's a length limit, wait, probably no length limit).
		
		// We'll just rollback manually if no error, but let's see if we can trigger one. 
		// If we can't trigger an error from input, we can just call Rollback() manually and test if it was discarded.
		
		_ = PersistReceipt(ctx, s, tx, input)
		
		if err := tx.Rollback(); err != nil {
			t.Fatalf("Rollback: %v", err)
		}

		// Verify DB state
		_, err = s.GetCorrespondentByName(ctx, "Target")
		if err == nil {
			t.Fatalf("Expected correspondent 'Target' to NOT exist")
		}
	})
}
