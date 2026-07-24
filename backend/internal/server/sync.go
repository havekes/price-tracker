package server

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
)

type SyncRequest struct {
	Tag int `json:"tag"`
}

func (s *Server) HandleSync(w http.ResponseWriter, r *http.Request) {
	var req SyncRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json payload", http.StatusBadRequest)
		return
	}

	if req.Tag == 0 {
		http.Error(w, "missing or invalid tag", http.StatusBadRequest)
		return
	}

	bgCtx := context.WithoutCancel(r.Context())

	go func() {
		defer func() {
			if r := recover(); r != nil {
				slog.Error("panic in sync", "err", r)
			}
		}()

		if s.pipeline != nil {
			_, _, err := s.pipeline.SyncPaperlessReceipts(bgCtx, req.Tag)
			if err != nil {
				slog.Error("sync pipeline error", "err", err)
			}
		}
	}()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "sync triggered asynchronously",
	})
}
