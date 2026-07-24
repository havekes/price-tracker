package store

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/havekes/price-tracker/internal/db"
)

func TestPriceRecordCRUD(t *testing.T) {
	t.Parallel()

	database, cleanup := testDB(t)
	defer cleanup()

	q := New(database)
	ctx := context.Background()

	// Create prerequisites.
	corr, _ := q.CreateCorrespondent(ctx, "Price Tester")
	receipt, _ := q.CreateReceipt(ctx, db.CreateReceiptParams{
		CorrespondentID: corr.ID,
		PurchasedAt:     time.Date(2026, 7, 20, 0, 0, 0, 0, time.UTC),
		Source:          "upload",
		ExternalDocID:   sql.NullString{},
		RawFileRef:      sql.NullString{},
	})

	ri, err := q.CreateRawItem(ctx, db.CreateRawItemParams{
		ReceiptID:     receipt.ID,
		ProductID:     sql.NullInt64{},
		RawText:       "Milk",
		RawQuantity:   sql.NullString{},
		QuantityValue: sql.NullFloat64{},
		QuantityUnit:  sql.NullString{},
	})
	if err != nil {
		t.Fatalf("CreateRawItem: %v", err)
	}

	// Create price record.
	pr1, err := q.CreatePriceRecord(ctx, db.CreatePriceRecordParams{
		RawItemID:  ri.ID,
		TotalPrice: 4.99,
		UnitPrice:  sql.NullFloat64{Float64: 0.31, Valid: true}, // per 100g
		Currency:   "USD",
	})
	if err != nil {
		t.Fatalf("CreatePriceRecord: %v", err)
	}
	if pr1.ID == 0 {
		t.Fatal("expected non-zero ID")
	}
	if pr1.TotalPrice != 4.99 {
		t.Fatalf("expected 4.99, got %f", pr1.TotalPrice)
	}
	if pr1.Currency != "USD" {
		t.Fatalf("expected USD, got %s", pr1.Currency)
	}

	// Get.
	got, err := q.GetPriceRecord(ctx, pr1.ID)
	if err != nil {
		t.Fatalf("GetPriceRecord: %v", err)
	}
	if got.TotalPrice != 4.99 {
		t.Fatalf("expected 4.99, got %f", got.TotalPrice)
	}

	// Get by raw item.
	byRI, err := q.GetPriceRecordByRawItem(ctx, ri.ID)
	if err != nil {
		t.Fatalf("GetPriceRecordByRawItem: %v", err)
	}
	if byRI.ID != pr1.ID {
		t.Fatalf("expected price record id %d, got %d", pr1.ID, byRI.ID)
	}

	// List.
	list, err := q.ListPriceRecords(ctx)
	if err != nil {
		t.Fatalf("ListPriceRecords: %v", err)
	}
	if len(list) < 1 {
		t.Fatal("expected at least 1 price record")
	}

	// Update.
	updated, err := q.UpdatePriceRecord(ctx, db.UpdatePriceRecordParams{
		ID:         pr1.ID,
		TotalPrice: 3.99,
		UnitPrice:  sql.NullFloat64{Float64: 0.25, Valid: true},
		Currency:   "EUR",
	})
	if err != nil {
		t.Fatalf("UpdatePriceRecord: %v", err)
	}
	if updated.TotalPrice != 3.99 {
		t.Fatalf("expected 3.99, got %f", updated.TotalPrice)
	}
	if updated.Currency != "EUR" {
		t.Fatalf("expected EUR, got %s", updated.Currency)
	}

	// Delete.
	if err := q.DeletePriceRecord(ctx, pr1.ID); err != nil {
		t.Fatalf("DeletePriceRecord: %v", err)
	}

	_, err = q.GetPriceRecord(ctx, pr1.ID)
	if err == nil {
		t.Fatal("expected error getting deleted price record")
	}
}

func TestPriceRecordUniqueRawItem(t *testing.T) {
	t.Parallel()

	database, cleanup := testDB(t)
	defer cleanup()

	q := New(database)
	ctx := context.Background()

	corr, _ := q.CreateCorrespondent(ctx, "Unique Price Tester")
	receipt, _ := q.CreateReceipt(ctx, db.CreateReceiptParams{
		CorrespondentID: corr.ID,
		PurchasedAt:     time.Date(2026, 7, 20, 0, 0, 0, 0, time.UTC),
		Source:          "upload",
		ExternalDocID:   sql.NullString{},
		RawFileRef:      sql.NullString{},
	})

	ri, _ := q.CreateRawItem(ctx, db.CreateRawItemParams{
		ReceiptID:     receipt.ID,
		ProductID:     sql.NullInt64{},
		RawText:       "Unique Price Item",
		RawQuantity:   sql.NullString{},
		QuantityValue: sql.NullFloat64{},
		QuantityUnit:  sql.NullString{},
	})

	_, err := q.CreatePriceRecord(ctx, db.CreatePriceRecordParams{
		RawItemID:  ri.ID,
		TotalPrice: 1.99,
		UnitPrice:  sql.NullFloat64{},
		Currency:   "USD",
	})
	if err != nil {
		t.Fatalf("CreatePriceRecord: %v", err)
	}

	// Second price record for the same raw_item should fail (UNIQUE constraint).
	_, err = q.CreatePriceRecord(ctx, db.CreatePriceRecordParams{
		RawItemID:  ri.ID,
		TotalPrice: 2.99,
		UnitPrice:  sql.NullFloat64{},
		Currency:   "USD",
	})
	if err == nil {
		t.Fatal("expected unique constraint violation for duplicate raw_item_id")
	}
}

func TestPriceRecordCascade(t *testing.T) {
	t.Parallel()

	database, cleanup := testDB(t)
	defer cleanup()

	q := New(database)
	ctx := context.Background()

	corr, _ := q.CreateCorrespondent(ctx, "Cascade Price Tester")
	receipt, _ := q.CreateReceipt(ctx, db.CreateReceiptParams{
		CorrespondentID: corr.ID,
		PurchasedAt:     time.Date(2026, 7, 20, 0, 0, 0, 0, time.UTC),
		Source:          "upload",
		ExternalDocID:   sql.NullString{},
		RawFileRef:      sql.NullString{},
	})

	ri, _ := q.CreateRawItem(ctx, db.CreateRawItemParams{
		ReceiptID:     receipt.ID,
		ProductID:     sql.NullInt64{},
		RawText:       "Cascade Price Item",
		RawQuantity:   sql.NullString{},
		QuantityValue: sql.NullFloat64{},
		QuantityUnit:  sql.NullString{},
	})

	pr, err := q.CreatePriceRecord(ctx, db.CreatePriceRecordParams{
		RawItemID:  ri.ID,
		TotalPrice: 5.99,
		UnitPrice:  sql.NullFloat64{},
		Currency:   "USD",
	})
	if err != nil {
		t.Fatalf("CreatePriceRecord: %v", err)
	}

	// Delete raw item — should cascade to price record.
	if err := q.DeleteRawItem(ctx, ri.ID); err != nil {
		t.Fatalf("DeleteRawItem: %v", err)
	}

	_, err = q.GetPriceRecord(ctx, pr.ID)
	if err == nil {
		t.Fatal("expected error — price record should be cascade-deleted with raw item")
	}
}
