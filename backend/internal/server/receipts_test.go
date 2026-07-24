package server

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/havekes/price-tracker/internal/vision"
)

type mockProcessor struct {
	receipt *vision.ExtractedReceipt
	err     error
}

func (m *mockProcessor) ProcessDirectUpload(ctx context.Context, imageBytes []byte, mimeType string, correspondentName string, purchaseDate time.Time) (*vision.ExtractedReceipt, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.receipt, nil
}

func TestUploadReceiptHandler(t *testing.T) {
	tests := []struct {
		name           string
		fileName       string
		fileContent    string
		fileMimeType   string
		correspondent  string
		purchasedAt    string
		mockErr        error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "success",
			fileName:       "test.jpg",
			fileContent:    "fake-image",
			fileMimeType:   "image/jpeg",
			correspondent:  "Test Store",
			purchasedAt:    "2024-01-01",
			expectedStatus: http.StatusOK,
			expectedBody:   "\"status\":\"success\"",
		},
		{
			name:           "invalid mime type",
			fileName:       "test.pdf",
			fileContent:    "fake-pdf",
			fileMimeType:   "application/pdf",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "unsupported file format",
		},
		{
			name:           "invalid date format",
			fileName:       "test.jpg",
			fileContent:    "fake-image",
			fileMimeType:   "image/jpeg",
			purchasedAt:    "invalid-date",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "invalid purchased_at format",
		},
		{
			name:           "processor error",
			fileName:       "test.jpg",
			fileContent:    "fake-image",
			fileMimeType:   "image/jpeg",
			mockErr:        errors.New("vision failed"),
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "failed to process receipt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := new(bytes.Buffer)
			writer := multipart.NewWriter(body)

			if tt.fileMimeType != "" {
				h := make(map[string][]string)
				h["Content-Disposition"] = []string{fmt.Sprintf("form-data; name=\"file\"; filename=\"%s\"", tt.fileName)}
				h["Content-Type"] = []string{tt.fileMimeType}
				part, err := writer.CreatePart(h)
				if err != nil {
					t.Fatal(err)
				}
				part.Write([]byte(tt.fileContent))
			} else {
				part, err := writer.CreateFormFile("file", tt.fileName)
				if err != nil {
					t.Fatal(err)
				}
				part.Write([]byte(tt.fileContent))
			}

			if tt.correspondent != "" {
				writer.WriteField("correspondent", tt.correspondent)
			}
			if tt.purchasedAt != "" {
				writer.WriteField("purchased_at", tt.purchasedAt)
			}
			writer.Close()

			req := httptest.NewRequest(http.MethodPost, "/api/receipts/upload", body)
			req.Header.Set("Content-Type", writer.FormDataContentType())

			w := httptest.NewRecorder()

			processor := &mockProcessor{
				receipt: &vision.ExtractedReceipt{},
				err:     tt.mockErr,
			}
			srv := New(nil, processor)
			srv.UploadReceiptHandler(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if !strings.Contains(w.Body.String(), tt.expectedBody) {
				t.Errorf("expected body to contain %q, got %q", tt.expectedBody, w.Body.String())
			}
		})
	}
}
