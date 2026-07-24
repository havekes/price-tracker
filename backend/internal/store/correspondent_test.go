package store

import (
	"context"
	"testing"

	"github.com/havekes/price-tracker/internal/db"
)

func TestCorrespondentCRUD(t *testing.T) {
	t.Parallel()

	database, cleanup := testDB(t)
	defer cleanup()

	q := New(database)
	ctx := context.Background()

	// Create.
	c1, err := q.CreateCorrespondent(ctx, "Alice")
	if err != nil {
		t.Fatalf("CreateCorrespondent: %v", err)
	}
	if c1.ID == 0 {
		t.Fatal("expected non-zero ID")
	}
	if c1.Name != "Alice" {
		t.Fatalf("expected name Alice, got %s", c1.Name)
	}

	c2, err := q.CreateCorrespondent(ctx, "Bob")
	if err != nil {
		t.Fatalf("CreateCorrespondent: %v", err)
	}
	if c2.Name != "Bob" {
		t.Fatalf("expected name Bob, got %s", c2.Name)
	}

	// Get.
	got, err := q.GetCorrespondent(ctx, c1.ID)
	if err != nil {
		t.Fatalf("GetCorrespondent: %v", err)
	}
	if got.Name != "Alice" {
		t.Fatalf("expected Alice, got %s", got.Name)
	}

	// List.
	list, err := q.ListCorrespondents(ctx)
	if err != nil {
		t.Fatalf("ListCorrespondents: %v", err)
	}
	if len(list) < 2 {
		t.Fatalf("expected at least 2 correspondents, got %d", len(list))
	}

	// Update.
	updated, err := q.UpdateCorrespondent(ctx, db.UpdateCorrespondentParams{
		ID:   c1.ID,
		Name: "Alice Updated",
	})
	if err != nil {
		t.Fatalf("UpdateCorrespondent: %v", err)
	}
	if updated.Name != "Alice Updated" {
		t.Fatalf("expected Alice Updated, got %s", updated.Name)
	}

	// Delete.
	if err := q.DeleteCorrespondent(ctx, c2.ID); err != nil {
		t.Fatalf("DeleteCorrespondent: %v", err)
	}

	// Verify deleted.
	_, err = q.GetCorrespondent(ctx, c2.ID)
	if err == nil {
		t.Fatal("expected error getting deleted correspondent")
	}
}

func TestCorrespondentUniqueName(t *testing.T) {
	t.Parallel()

	database, cleanup := testDB(t)
	defer cleanup()

	q := New(database)
	ctx := context.Background()

	_, err := q.CreateCorrespondent(ctx, "Unique")
	if err != nil {
		t.Fatalf("CreateCorrespondent: %v", err)
	}

	// Duplicate name should fail.
	_, err = q.CreateCorrespondent(ctx, "Unique")
	if err == nil {
		t.Fatal("expected unique constraint violation")
	}
}
