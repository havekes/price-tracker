package server

import (
	"encoding/json"
	"net/http"
)

// HealthHandler responds with {"status":"ok"} for health checks.
func (s *Server) HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
