package server

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/havekes/price-tracker/internal/db"
)

type UpdateProductRequest struct {
	DisplayName string `json:"display_name"`
	BaseUnit    string `json:"base_unit"`
}

type LinkProductsRequest struct {
	ProductAID int64 `json:"product_a_id"`
	ProductBID int64 `json:"product_b_id"`
}

type ProductResponse struct {
	ID          int64     `json:"id"`
	DisplayName string    `json:"display_name"`
	BaseUnit    string    `json:"base_unit"`
	CreatedAt   time.Time `json:"created_at"`
}

type LinkProductsResponse struct {
	ID         int64     `json:"id"`
	ProductAID int64     `json:"product_a_id"`
	ProductBID int64     `json:"product_b_id"`
	CreatedAt  time.Time `json:"created_at"`
}

func (s *Server) UpdateProductHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid product id", http.StatusBadRequest)
		return
	}

	var req UpdateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.DisplayName == "" || req.BaseUnit == "" {
		http.Error(w, "display_name and base_unit are required", http.StatusBadRequest)
		return
	}

	product, err := s.querier.UpdateProduct(r.Context(), db.UpdateProductParams{
		ID:          id,
		DisplayName: req.DisplayName,
		BaseUnit:    req.BaseUnit,
	})

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "product not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ProductResponse{
		ID:          product.ID,
		DisplayName: product.DisplayName,
		BaseUnit:    product.BaseUnit,
		CreatedAt:   product.CreatedAt,
	})
}

func (s *Server) LinkProductsHandler(w http.ResponseWriter, r *http.Request) {
	var req LinkProductsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.ProductAID == 0 || req.ProductBID == 0 {
		http.Error(w, "missing product_a_id or product_b_id", http.StatusBadRequest)
		return
	}

	if req.ProductAID == req.ProductBID {
		http.Error(w, "products cannot be linked to themselves", http.StatusBadRequest)
		return
	}

	if req.ProductAID > req.ProductBID {
		req.ProductAID, req.ProductBID = req.ProductBID, req.ProductAID
	}

	link, err := s.querier.CreateMarketplaceLink(r.Context(), db.CreateMarketplaceLinkParams{
		ProductAID: req.ProductAID,
		ProductBID: req.ProductBID,
	})

	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(LinkProductsResponse{
		ID:         link.ID,
		ProductAID: link.ProductAID,
		ProductBID: link.ProductBID,
		CreatedAt:  link.CreatedAt,
	})
}