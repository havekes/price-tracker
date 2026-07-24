package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const maxUploadSize = 32 << 20 // 32MB

type UploadReceiptResponse struct {
	Status  string      `json:"status"`
	Receipt interface{} `json:"receipt"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func (s *Server) UploadReceiptHandler(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		sendError(w, http.StatusBadRequest, fmt.Sprintf("file too large or invalid multipart form: %v", err))
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		sendError(w, http.StatusBadRequest, "missing \"file\" in form")
		return
	}
	defer file.Close()

	mimeType := header.Header.Get("Content-Type")
	if mimeType != "image/jpeg" && mimeType != "image/png" && mimeType != "image/webp" {
		sendError(w, http.StatusBadRequest, "unsupported file format. Allowed: image/jpeg, image/png, image/webp")
		return
	}

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to read file")
		return
	}

	correspondent := r.FormValue("correspondent")
	if correspondent == "" {
		correspondent = "Unknown"
	}

	purchasedAtStr := r.FormValue("purchased_at")
	var purchaseDate time.Time
	if purchasedAtStr != "" {
		parsed, err := time.Parse(time.RFC3339, purchasedAtStr)
		if err != nil {
			parsed, err = time.Parse(time.DateOnly, purchasedAtStr)
			if err != nil {
				sendError(w, http.StatusBadRequest, "invalid purchased_at format. Use RFC3339 or YYYY-MM-DD")
				return
			}
		}
		purchaseDate = parsed
	} else {
		purchaseDate = time.Now()
	}

	receipt, err := s.processor.ProcessDirectUpload(r.Context(), fileBytes, mimeType, correspondent, purchaseDate)
	if err != nil {
		sendError(w, http.StatusBadRequest, fmt.Sprintf("failed to process receipt: %v", err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(UploadReceiptResponse{
		Status:  "success",
		Receipt: receipt,
	})
}

func sendError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(ErrorResponse{Error: message})
}
