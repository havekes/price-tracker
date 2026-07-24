package store

import (
	"context"
	"testing"

	"github.com/havekes/price-tracker/internal/db"
)

func TestMarketplaceLinkCRUD(t *testing.T) {
	t.Parallel()

	database, cleanup := testDB(t)
	defer cleanup()

	q := New(database)
	ctx := context.Background()

	// Create two products.
	pA, err := q.CreateProduct(ctx, db.CreateProductParams{
		DisplayName: "Product A",
		BaseUnit:    "g",
	})
	if err != nil {
		t.Fatalf("CreateProduct A: %v", err)
	}

	pB, err := q.CreateProduct(ctx, db.CreateProductParams{
		DisplayName: "Product B",
		BaseUnit:    "g",
	})
	if err != nil {
		t.Fatalf("CreateProduct B: %v", err)
	}

	// Ensure pA.ID < pB.ID for the CHECK constraint.
	a, b := pA.ID, pB.ID
	if a > b {
		a, b = b, a
	}

	// Create link.
	link, err := q.CreateMarketplaceLink(ctx, db.CreateMarketplaceLinkParams{
		ProductAID: a,
		ProductBID: b,
	})
	if err != nil {
		t.Fatalf("CreateMarketplaceLink: %v", err)
	}
	if link.ID == 0 {
		t.Fatal("expected non-zero ID")
	}

	// Get.
	got, err := q.GetMarketplaceLink(ctx, link.ID)
	if err != nil {
		t.Fatalf("GetMarketplaceLink: %v", err)
	}
	if got.ProductAID != a || got.ProductBID != b {
		t.Fatalf("expected (%d,%d), got (%d,%d)", a, b, got.ProductAID, got.ProductBID)
	}

	// List.
	list, err := q.ListMarketplaceLinks(ctx)
	if err != nil {
		t.Fatalf("ListMarketplaceLinks: %v", err)
	}
	if len(list) < 1 {
		t.Fatal("expected at least 1 marketplace link")
	}

	// List by product.
	byProduct, err := q.ListMarketplaceLinksByProduct(ctx, a)
	if err != nil {
		t.Fatalf("ListMarketplaceLinksByProduct: %v", err)
	}
	if len(byProduct) != 1 {
		t.Fatalf("expected 1 link, got %d", len(byProduct))
	}

	// Delete.
	if err := q.DeleteMarketplaceLink(ctx, link.ID); err != nil {
		t.Fatalf("DeleteMarketplaceLink: %v", err)
	}

	_, err = q.GetMarketplaceLink(ctx, link.ID)
	if err == nil {
		t.Fatal("expected error getting deleted marketplace link")
	}
}

func TestMarketplaceLinkCheckConstraint(t *testing.T) {
	t.Parallel()

	database, cleanup := testDB(t)
	defer cleanup()

	q := New(database)
	ctx := context.Background()

	pA, _ := q.CreateProduct(ctx, db.CreateProductParams{DisplayName: "P1", BaseUnit: "unit"})
	pB, _ := q.CreateProduct(ctx, db.CreateProductParams{DisplayName: "P2", BaseUnit: "unit"})

	// Swap to test the CHECk constraint: product_a_id must be < product_b_id.
	small, big := pA.ID, pB.ID
	if small > big {
		small, big = big, small
	}

	// This should work.
	_, err := q.CreateMarketplaceLink(ctx, db.CreateMarketplaceLinkParams{
		ProductAID: small,
		ProductBID: big,
	})
	if err != nil {
		t.Fatalf("expected link to succeed: %v", err)
	}

	// Duplicate (same pair) should fail — UNIQUE constraint.
	_, err = q.CreateMarketplaceLink(ctx, db.CreateMarketplaceLinkParams{
		ProductAID: small,
		ProductBID: big,
	})
	if err == nil {
		t.Fatal("expected unique constraint violation for duplicate link")
	}
}

func TestMarketplaceLinkCascade(t *testing.T) {
	t.Parallel()

	database, cleanup := testDB(t)
	defer cleanup()

	q := New(database)
	ctx := context.Background()

	pA, _ := q.CreateProduct(ctx, db.CreateProductParams{DisplayName: "Cascade A", BaseUnit: "unit"})
	pB, _ := q.CreateProduct(ctx, db.CreateProductParams{DisplayName: "Cascade B", BaseUnit: "unit"})

	small, big := pA.ID, pB.ID
	if small > big {
		small, big = big, small
	}

	link, err := q.CreateMarketplaceLink(ctx, db.CreateMarketplaceLinkParams{
		ProductAID: small,
		ProductBID: big,
	})
	if err != nil {
		t.Fatalf("CreateMarketplaceLink: %v", err)
	}

	// Deleting one product should cascade-delete the link.
	if err := q.DeleteProduct(ctx, small); err != nil {
		t.Fatalf("DeleteProduct: %v", err)
	}

	_, err = q.GetMarketplaceLink(ctx, link.ID)
	if err == nil {
		t.Fatal("expected error — marketplace link should be cascade-deleted with product")
	}
}
