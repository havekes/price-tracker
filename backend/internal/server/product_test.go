package server

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/havekes/price-tracker/internal/db"
	"github.com/havekes/price-tracker/internal/store"
)

// mockQuerier implements store.Querier for testing.
type mockQuerier struct {
	store.Querier // Embed to satisfy interface for methods not mocked here
	updateProductFunc         func(ctx context.Context, arg db.UpdateProductParams) (db.Product, error)
	createMarketplaceLinkFunc func(ctx context.Context, arg db.CreateMarketplaceLinkParams) (db.MarketplaceLink, error)
}

func (m *mockQuerier) UpdateProduct(ctx context.Context, arg db.UpdateProductParams) (db.Product, error) {
	if m.updateProductFunc != nil {
		return m.updateProductFunc(ctx, arg)
	}
	return db.Product{}, nil
}

func (m *mockQuerier) CreateMarketplaceLink(ctx context.Context, arg db.CreateMarketplaceLinkParams) (db.MarketplaceLink, error) {
	if m.createMarketplaceLinkFunc != nil {
		return m.createMarketplaceLinkFunc(ctx, arg)
	}
	return db.MarketplaceLink{}, nil
}

func TestUpdateProductHandler(t *testing.T) {
	tests := []struct {
		name           string
		productID      string
		reqBody        UpdateProductRequest
		mockUpdateFunc func(ctx context.Context, arg db.UpdateProductParams) (db.Product, error)
		expectedStatus int
	}{
		{
			name:      "success",
			productID: "1",
			reqBody:   UpdateProductRequest{DisplayName: "Apple", BaseUnit: "kg"},
			mockUpdateFunc: func(ctx context.Context, arg db.UpdateProductParams) (db.Product, error) {
				return db.Product{ID: 1, DisplayName: "Apple", BaseUnit: "kg"}, nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "empty display name",
			productID:      "1",
			reqBody:        UpdateProductRequest{DisplayName: "", BaseUnit: "kg"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:      "product not found",
			productID: "999",
			reqBody:   UpdateProductRequest{DisplayName: "Orange", BaseUnit: "kg"},
			mockUpdateFunc: func(ctx context.Context, arg db.UpdateProductParams) (db.Product, error) {
				return db.Product{}, sql.ErrNoRows
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "invalid product id",
			productID:      "abc",
			reqBody:        UpdateProductRequest{DisplayName: "Apple", BaseUnit: "kg"},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &mockQuerier{updateProductFunc: tt.mockUpdateFunc}
			srv := New(m)

			bodyBytes, _ := json.Marshal(tt.reqBody)
			req := httptest.NewRequest(http.MethodPut, "/api/products/"+tt.productID, bytes.NewBuffer(bodyBytes))

			// Setup chi context to simulate path params
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.productID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rec := httptest.NewRecorder()
			srv.UpdateProductHandler(rec, req)

			if rec.Result().StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Result().StatusCode)
			}

			if tt.expectedStatus == http.StatusOK {
				var resp ProductResponse
				if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				if resp.ID != 1 || resp.DisplayName != "Apple" || resp.BaseUnit != "kg" {
					t.Errorf("unexpected response body: %+v", resp)
				}
			}
		})
	}
}

func TestLinkProductsHandler(t *testing.T) {
	tests := []struct {
		name           string
		reqBody        LinkProductsRequest
		mockLinkFunc   func(ctx context.Context, arg db.CreateMarketplaceLinkParams) (db.MarketplaceLink, error)
		expectedStatus int
	}{
		{
			name:    "success sorted",
			reqBody: LinkProductsRequest{ProductAID: 1, ProductBID: 2},
			mockLinkFunc: func(ctx context.Context, arg db.CreateMarketplaceLinkParams) (db.MarketplaceLink, error) {
				if arg.ProductAID != 1 || arg.ProductBID != 2 {
					t.Errorf("expected sorted IDs 1,2 got %d,%d", arg.ProductAID, arg.ProductBID)
				}
				return db.MarketplaceLink{ID: 1, ProductAID: 1, ProductBID: 2}, nil
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:    "success unsorted",
			reqBody: LinkProductsRequest{ProductAID: 2, ProductBID: 1},
			mockLinkFunc: func(ctx context.Context, arg db.CreateMarketplaceLinkParams) (db.MarketplaceLink, error) {
				if arg.ProductAID != 1 || arg.ProductBID != 2 {
					t.Errorf("expected sorted IDs 1,2 got %d,%d", arg.ProductAID, arg.ProductBID)
				}
				return db.MarketplaceLink{ID: 1, ProductAID: 1, ProductBID: 2}, nil
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "identical ids",
			reqBody:        LinkProductsRequest{ProductAID: 1, ProductBID: 1},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing id",
			reqBody:        LinkProductsRequest{ProductAID: 1}, // ProductBID is 0
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &mockQuerier{createMarketplaceLinkFunc: tt.mockLinkFunc}
			srv := New(m)

			bodyBytes, _ := json.Marshal(tt.reqBody)
			req := httptest.NewRequest(http.MethodPost, "/api/products/link", bytes.NewBuffer(bodyBytes))

			rec := httptest.NewRecorder()
			srv.LinkProductsHandler(rec, req)

			if rec.Result().StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Result().StatusCode)
			}

			if tt.expectedStatus == http.StatusCreated {
				var resp LinkProductsResponse
				if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				if resp.ID != 1 || resp.ProductAID != 1 || resp.ProductBID != 2 {
					t.Errorf("unexpected response body: %+v", resp)
				}
			}
		})
	}
}
