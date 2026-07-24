package server

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/havekes/price-tracker/internal/db"
)

type ProductDetailResponse struct {
	Product        db.Product                    `json:"product"`
	LinkedProducts []db.Product                  `json:"linked_products"`
	PriceHistory   map[string][]PriceHistoryItem `json:"price_history"`
}

type PriceHistoryItem struct {
	ID                int64    `json:"id"`
	TotalPrice        float64  `json:"total_price"`
	UnitPrice         *float64 `json:"unit_price"`
	RawText           string   `json:"raw_text"`
	RawQuantity       *string  `json:"raw_quantity"`
	CorrespondentName string   `json:"correspondent_name"`
	PurchasedAt       string   `json:"purchased_at"`
}

func jsonError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func (s *Server) GetProduct(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		jsonError(w, "invalid product id", http.StatusBadRequest)
		return
	}

	product, err := s.querier.GetProduct(r.Context(), id)
	if err != nil {
		if err == sql.ErrNoRows {
			jsonError(w, "product not found", http.StatusNotFound)
			return
		}
		jsonError(w, "database error", http.StatusInternalServerError)
		return
	}

	linked, err := s.querier.ListLinkedProducts(r.Context(), id)
	if err != nil {
		jsonError(w, "database error", http.StatusInternalServerError)
		return
	}
	if linked == nil {
		linked = []db.Product{}
	}

	history, err := s.querier.ListPriceHistoryByProduct(r.Context(), sql.NullInt64{Int64: id, Valid: true})
	if err != nil {
		jsonError(w, "database error", http.StatusInternalServerError)
		return
	}

	groupedHistory := make(map[string][]PriceHistoryItem)
	for _, h := range history {
		var uPrice *float64
		if h.UnitPrice.Valid {
			uPrice = &h.UnitPrice.Float64
		}
		var rQuant *string
		if h.RawQuantity.Valid {
			rQuant = &h.RawQuantity.String
		}

		item := PriceHistoryItem{
			ID:                h.ID,
			TotalPrice:        h.TotalPrice,
			UnitPrice:         uPrice,
			RawText:           h.RawText,
			RawQuantity:       rQuant,
			CorrespondentName: h.CorrespondentName,
			PurchasedAt:       h.PurchasedAt.Format("2006-01-02"),
		}
		groupedHistory[h.CorrespondentName] = append(groupedHistory[h.CorrespondentName], item)
	}

	resp := ProductDetailResponse{
		Product:        product,
		LinkedProducts: linked,
		PriceHistory:   groupedHistory,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
