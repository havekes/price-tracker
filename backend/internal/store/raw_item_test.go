package store

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/havekes/price-tracker/internal/db"
)

func TestRawItemCRUD(t *testing.T) {
	t.Parallel()

	database, cleanup := testDB(t)
	defer cleanup()

	q := New(database)
	ctx := context.Background()

	// Create prerequisites.
	corr, err := q.CreateCorrespondent(ctx, "RawItem Tester")
	if err != nil {
		t.Fatalf("CreateCorrespondent: %v", err)
	}

	receipt, err := q.CreateReceipt(ctx, db.CreateReceiptParams{
		CorrespondentID: corr.ID,
		PurchasedAt:     time.Date(2026, 7, 20, 0, 0, 0, 0, time.UTC),
		Source:          "upload",
		ExternalDocID:   sql.NullString{},
		RawFileRef:      sql.NullString{},
	})
	if err != nil {
		t.Fatalf("CreateReceipt: %v", err)
	}

	product, err := q.CreateProduct(ctx, db.CreateProductParams{
		DisplayName: "Test Product",
		BaseUnit:    "g",
	})
	if err != nil {
		t.Fatalf("CreateProduct: %v", err)
	}

	// Create raw item with product linked.
	ri1, err := q.CreateRawItem(ctx, db.CreateRawItemParams{
		ReceiptID:     receipt.ID,
		ProductID:     sql.NullInt64{Int64: product.ID, Valid: true},
		RawText:       "Organic Milk 1 gal",
		RawQuantity:   sql.NullString{String: "1 gal", Valid: true},
		QuantityValue: sql.NullFloat64{Float64: 1.0, Valid: true},
		QuantityUnit:  sql.NullString{String: "gal", Valid: true},
	})
	if err != nil {
		t.Fatalf("CreateRawItem: %v", err)
	}
	if ri1.ID == 0 {
		t.Fatal("expected non-zero ID")
	}
	if ri1.RawText != "Organic Milk 1 gal" {
		t.Fatalf("expected Organic Milk 1 gal, got %s", ri1.RawText)
	}

	// Create raw item without product (nullable).
	ri2, err := q.CreateRawItem(ctx, db.CreateRawItemParams{
		ReceiptID:     receipt.ID,
		ProductID:     sql.NullInt64{},
		RawText:       "Generic Item",
		RawQuantity:   sql.NullString{},
		QuantityValue: sql.NullFloat64{},
		QuantityUnit:  sql.NullString{},
	})
	if err != nil {
		t.Fatalf("CreateRawItem: %v", err)
	}
	if ri2.ProductID.Valid {
		t.Fatal("expected nil product_id")
	}

	// Get.
	got, err := q.GetRawItem(ctx, ri1.ID)
	if err != nil {
		t.Fatalf("GetRawItem: %v", err)
	}
	if got.RawText != "Organic Milk 1 gal" {
		t.Fatalf("expected Organic Milk 1 gal, got %s", got.RawText)
	}

	// List by receipt.
	items, err := q.ListRawItemsByReceipt(ctx, receipt.ID)
	if err != nil {
		t.Fatalf("ListRawItemsByReceipt: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}

	// Update.
	updated, err := q.UpdateRawItem(ctx, db.UpdateRawItemParams{
		ID:            ri1.ID,
		ProductID:     sql.NullInt64{Int64: product.ID, Valid: true},
		RawText:       "Organic Milk 1/2 gal",
		RawQuantity:   sql.NullString{String: "0.5 gal", Valid: true},
		QuantityValue: sql.NullFloat64{Float64: 0.5, Valid: true},
		QuantityUnit:  sql.NullString{String: "gal", Valid: true},
	})
	if err != nil {
		t.Fatalf("UpdateRawItem: %v", err)
	}
	if updated.RawText != "Organic Milk 1/2 gal" {
		t.Fatalf("expected Organic Milk 1/2 gal, got %s", updated.RawText)
	}

	// Delete.
	if err := q.DeleteRawItem(ctx, ri2.ID); err != nil {
		t.Fatalf("DeleteRawItem: %v", err)
	}

	_, err = q.GetRawItem(ctx, ri2.ID)
	if err == nil {
		t.Fatal("expected error getting deleted raw item")
	}
}

func TestRawItemCascade(t *testing.T) {
	t.Parallel()

	database, cleanup := testDB(t)
	defer cleanup()

	q := New(database)
	ctx := context.Background()

	corr, _ := q.CreateCorrespondent(ctx, "Cascade Tester")
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
		RawText:       "Cascade test",
		RawQuantity:   sql.NullString{},
		QuantityValue: sql.NullFloat64{},
		QuantityUnit:  sql.NullString{},
	})
	if err != nil {
		t.Fatalf("CreateRawItem: %v", err)
	}

	// Delete receipt — should cascade to raw items.
	if err := q.DeleteReceipt(ctx, receipt.ID); err != nil {
		t.Fatalf("DeleteReceipt: %v", err)
	}

	_, err = q.GetRawItem(ctx, ri.ID)
	if err == nil {
		t.Fatal("expected error — raw item should be cascade-deleted with receipt")
	}
}
