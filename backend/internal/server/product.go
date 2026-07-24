package server

import (
	"encoding/json"
	"net/http"
)

type ProductResponse struct {
	ID             int64    `json:"id"`
	DisplayName    string   `json:"display_name"`
	BaseUnit       string   `json:"base_unit"`
	LatestAvgPrice *float64 `json:"latest_avg_price"`
}

func (s *Server) ListProductsHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := s.querier.ListProductsWithPrices(r.Context())
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	resp := []ProductResponse{}
	for _, row := range rows {
		var price *float64
		if row.LatestAvgPrice.Valid {
			v := row.LatestAvgPrice.Float64
			price = &v
		}
		resp = append(resp, ProductResponse{
			ID:             row.ID,
			DisplayName:    row.DisplayName,
			BaseUnit:       row.BaseUnit,
			LatestAvgPrice: price,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		// Log error, but headers are already sent
	}
}
