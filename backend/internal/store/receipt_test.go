package store

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/havekes/price-tracker/internal/db"
)

func TestReceiptCRUD(t *testing.T) {
	t.Parallel()

	database, cleanup := testDB(t)
	defer cleanup()

	q := New(database)
	ctx := context.Background()

	// Create a correspondent first (required FK).
	corr, err := q.CreateCorrespondent(ctx, "Receipt Tester")
	if err != nil {
		t.Fatalf("CreateCorrespondent: %v", err)
	}

	purchasedAt := time.Date(2026, 7, 20, 0, 0, 0, 0, time.UTC)
	externalDocID := sql.NullString{String: "doc-001", Valid: true}

	// Create.
	r1, err := q.CreateReceipt(ctx, db.CreateReceiptParams{
		CorrespondentID: corr.ID,
		PurchasedAt:     purchasedAt,
		Source:          "upload",
		ExternalDocID:   externalDocID,
		RawFileRef:      sql.NullString{},
	})
	if err != nil {
		t.Fatalf("CreateReceipt: %v", err)
	}
	if r1.ID == 0 {
		t.Fatal("expected non-zero ID")
	}
	if r1.Source != "upload" {
		t.Fatalf("expected source upload, got %s", r1.Source)
	}

	// Get.
	got, err := q.GetReceipt(ctx, r1.ID)
	if err != nil {
		t.Fatalf("GetReceipt: %v", err)
	}
	if got.CorrespondentID != corr.ID {
		t.Fatalf("expected correspondent_id %d, got %d", corr.ID, got.CorrespondentID)
	}

	// List.
	list, err := q.ListReceipts(ctx)
	if err != nil {
		t.Fatalf("ListReceipts: %v", err)
	}
	if len(list) < 1 {
		t.Fatal("expected at least 1 receipt")
	}

	// List by correspondent.
	byCorr, err := q.ListReceiptsByCorrespondent(ctx, corr.ID)
	if err != nil {
		t.Fatalf("ListReceiptsByCorrespondent: %v", err)
	}
	if len(byCorr) != 1 {
		t.Fatalf("expected 1 receipt, got %d", len(byCorr))
	}

	// Update.
	purchasedAt2 := time.Date(2026, 7, 21, 0, 0, 0, 0, time.UTC)
	updated, err := q.UpdateReceipt(ctx, db.UpdateReceiptParams{
		ID:              r1.ID,
		CorrespondentID: corr.ID,
		PurchasedAt:     purchasedAt2,
		Source:          "paperless",
		ExternalDocID:   externalDocID,
		RawFileRef:      sql.NullString{},
	})
	if err != nil {
		t.Fatalf("UpdateReceipt: %v", err)
	}
	if updated.Source != "paperless" {
		t.Fatalf("expected source paperless, got %s", updated.Source)
	}

	// Delete.
	if err := q.DeleteReceipt(ctx, r1.ID); err != nil {
		t.Fatalf("DeleteReceipt: %v", err)
	}

	// Verify deleted.
	_, err = q.GetReceipt(ctx, r1.ID)
	if err == nil {
		t.Fatal("expected error getting deleted receipt")
	}
}

func TestReceiptFKRestrict(t *testing.T) {
	t.Parallel()

	database, cleanup := testDB(t)
	defer cleanup()

	q := New(database)
	ctx := context.Background()

	corr, err := q.CreateCorrespondent(ctx, "FK Restrict Tester")
	if err != nil {
		t.Fatalf("CreateCorrespondent: %v", err)
	}

	_, err = q.CreateReceipt(ctx, db.CreateReceiptParams{
		CorrespondentID: corr.ID,
		PurchasedAt:     time.Date(2026, 7, 20, 0, 0, 0, 0, time.UTC),
		Source:          "upload",
		ExternalDocID:   sql.NullString{},
		RawFileRef:      sql.NullString{},
	})
	if err != nil {
		t.Fatalf("CreateReceipt: %v", err)
	}

	// Deleting correspondent with receipts should fail (ON DELETE RESTRICT).
	err = q.DeleteCorrespondent(ctx, corr.ID)
	if err == nil {
		t.Fatal("expected foreign key restrict error")
	}
}
