package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/havekes/price-tracker/internal/db"
	"github.com/havekes/price-tracker/internal/store"
)

type mockProductQuerier struct {
	db.Querier
}

func (m mockProductQuerier) WithTx(tx *sql.Tx) store.Querier {
	return m
}

func (m mockProductQuerier) GetProduct(ctx context.Context, id int64) (db.Product, error) {
	if id == 1 {
		return db.Product{ID: 1, DisplayName: "Mock Product", BaseUnit: "g"}, nil
	}
	return db.Product{}, sql.ErrNoRows
}

func (m mockProductQuerier) ListLinkedProducts(ctx context.Context, productAID int64) ([]db.Product, error) {
	return []db.Product{}, nil
}

func (m mockProductQuerier) ListPriceHistoryByProduct(ctx context.Context, productID sql.NullInt64) ([]db.ListPriceHistoryByProductRow, error) {
	if productID.Int64 == 1 {
		return []db.ListPriceHistoryByProductRow{
			{
				ID:                10,
				TotalPrice:        1.99,
				UnitPrice:         sql.NullFloat64{Float64: 0.0199, Valid: true},
				RawText:           "Mock item",
				RawQuantity:       sql.NullString{String: "100g", Valid: true},
				CorrespondentName: "Store A",
				PurchasedAt:       time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			},
		}, nil
	}
	return []db.ListPriceHistoryByProductRow{}, nil
}

func TestGetProductHandler(t *testing.T) {
	s := New(mockProductQuerier{})

	r := chi.NewRouter()
	r.Get("/api/products/{id}", s.GetProduct)

	tests := []struct {
		name           string
		productID      string
		expectedStatus int
	}{
		{"Valid Product", "1", http.StatusOK},
		{"Non-existent Product", "999", http.StatusNotFound},
		{"Invalid Product ID", "abc", http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/api/products/"+tt.productID, nil)
			rr := httptest.NewRecorder()

			r.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tt.expectedStatus)
			}

			if tt.expectedStatus == http.StatusOK {
				var resp ProductDetailResponse
				if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
					t.Errorf("failed to decode response: %v", err)
				}
				if resp.Product.ID != 1 {
					t.Errorf("expected product ID 1, got %v", resp.Product.ID)
				}
				if len(resp.PriceHistory) != 1 {
					t.Errorf("expected 1 store in price history, got %v", len(resp.PriceHistory))
				}
			} else {
				if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
					t.Errorf("expected Content-Type application/json, got %v", ct)
				}
				var errResp map[string]string
				if err := json.NewDecoder(rr.Body).Decode(&errResp); err != nil {
					t.Errorf("failed to decode error json: %v", err)
				}
				if errResp["error"] == "" {
					t.Errorf("expected error message in response body")
				}
			}
		})
	}
}
