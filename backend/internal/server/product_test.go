package server_test

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/havekes/price-tracker/internal/db"
	"github.com/havekes/price-tracker/internal/server"
	"github.com/havekes/price-tracker/internal/store"
)

// mockQuerier struct implements db.Querier
type mockQuerier struct {
	store.Querier
	products []db.ListProductsWithPricesRow
}

func (m *mockQuerier) ListProductsWithPrices(ctx context.Context) ([]db.ListProductsWithPricesRow, error) {
	return m.products, nil
}

func TestListProductsHandler(t *testing.T) {
	tests := []struct {
		name     string
		products []db.ListProductsWithPricesRow
		wantBody string
	}{
		{
			name:     "empty database",
			products: []db.ListProductsWithPricesRow{},
			wantBody: `[]`,
		},
		{
			name: "with products and prices",
			products: []db.ListProductsWithPricesRow{
				{ID: 1, DisplayName: "Apple", BaseUnit: "kg", LatestAvgPrice: sql.NullFloat64{Float64: 2.5, Valid: true}},
				{ID: 2, DisplayName: "Banana", BaseUnit: "kg", LatestAvgPrice: sql.NullFloat64{Valid: false}},
			},
			wantBody: `[{"id":1,"display_name":"Apple","base_unit":"kg","latest_avg_price":2.5},{"id":2,"display_name":"Banana","base_unit":"kg","latest_avg_price":null}]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mq := &mockQuerier{products: tt.products}
			srv := server.New(mq)

			req := httptest.NewRequest(http.MethodGet, "/api/products", nil)
			w := httptest.NewRecorder()

			srv.ListProductsHandler(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
			}

			if ct := w.Header().Get("Content-Type"); ct != "application/json" {
				t.Errorf("expected content type application/json, got %s", ct)
			}
            
            // Trim newline from encode
            got := w.Body.String()
            if len(got) > 0 && got[len(got)-1] == '\n' {
                got = got[:len(got)-1]
            }

			if got != tt.wantBody {
				t.Errorf("expected body %s, got %s", tt.wantBody, got)
			}
		})
	}
}
