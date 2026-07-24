package store

import (
	"context"
	"testing"

	"github.com/havekes/price-tracker/internal/db"
)

func TestProductCRUD(t *testing.T) {
	t.Parallel()

	database, cleanup := testDB(t)
	defer cleanup()

	q := New(database)
	ctx := context.Background()

	// Create.
	p1, err := q.CreateProduct(ctx, db.CreateProductParams{
		DisplayName: "Organic Milk",
		BaseUnit:    "g",
	})
	if err != nil {
		t.Fatalf("CreateProduct: %v", err)
	}
	if p1.ID == 0 {
		t.Fatal("expected non-zero ID")
	}
	if p1.DisplayName != "Organic Milk" {
		t.Fatalf("expected Organic Milk, got %s", p1.DisplayName)
	}

	p2, err := q.CreateProduct(ctx, db.CreateProductParams{
		DisplayName: "Sourdough Bread",
		BaseUnit:    "g",
	})
	if err != nil {
		t.Fatalf("CreateProduct: %v", err)
	}

	// Get.
	got, err := q.GetProduct(ctx, p1.ID)
	if err != nil {
		t.Fatalf("GetProduct: %v", err)
	}
	if got.DisplayName != "Organic Milk" {
		t.Fatalf("expected Organic Milk, got %s", got.DisplayName)
	}

	// List.
	list, err := q.ListProducts(ctx)
	if err != nil {
		t.Fatalf("ListProducts: %v", err)
	}
	if len(list) < 2 {
		t.Fatalf("expected at least 2 products, got %d", len(list))
	}

	// Update.
	updated, err := q.UpdateProduct(ctx, db.UpdateProductParams{
		ID:          p1.ID,
		DisplayName: "Organic Whole Milk",
		BaseUnit:    "ml",
	})
	if err != nil {
		t.Fatalf("UpdateProduct: %v", err)
	}
	if updated.DisplayName != "Organic Whole Milk" {
		t.Fatalf("expected Organic Whole Milk, got %s", updated.DisplayName)
	}
	if updated.BaseUnit != "ml" {
		t.Fatalf("expected base_unit ml, got %s", updated.BaseUnit)
	}

	// Delete.
	if err := q.DeleteProduct(ctx, p2.ID); err != nil {
		t.Fatalf("DeleteProduct: %v", err)
	}

	_, err = q.GetProduct(ctx, p2.ID)
	if err == nil {
		t.Fatal("expected error getting deleted product")
	}
}
