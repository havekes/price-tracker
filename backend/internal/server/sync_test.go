package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type mockPipeline struct {
	called bool
	tag    int
}

func (m *mockPipeline) SyncPaperlessReceipts(ctx context.Context, tag int) (int, int, error) {
	m.called = true
	m.tag = tag
	return 0, 0, nil
}

func TestHandleSync(t *testing.T) {
	mock := &mockPipeline{}
	srv := New(nil, mock)

	t.Run("valid payload", func(t *testing.T) {
		mock.called = false
		body := bytes.NewReader([]byte(`{"tag": 123}`))
		req := httptest.NewRequest(http.MethodPost, "/api/sync", body)
		rec := httptest.NewRecorder()

		srv.HandleSync(rec, req)

		if rec.Code != http.StatusAccepted {
			t.Errorf("expected status 202, got %d", rec.Code)
		}

		var resp map[string]string
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if resp["status"] != "sync triggered asynchronously" {
			t.Errorf("unexpected status message: %s", resp["status"])
		}

		time.Sleep(50 * time.Millisecond) // wait for goroutine

		if !mock.called {
			t.Error("expected pipeline to be called")
		}
		if mock.tag != 123 {
			t.Errorf("expected tag 123, got %d", mock.tag)
		}
	})

	t.Run("invalid payload", func(t *testing.T) {
		mock.called = false
		body := bytes.NewReader([]byte(`{"tag": "not-an-int"}`))
		req := httptest.NewRequest(http.MethodPost, "/api/sync", body)
		rec := httptest.NewRecorder()

		srv.HandleSync(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", rec.Code)
		}
	})

	t.Run("missing tag", func(t *testing.T) {
		mock.called = false
		body := bytes.NewReader([]byte(`{}`))
		req := httptest.NewRequest(http.MethodPost, "/api/sync", body)
		rec := httptest.NewRecorder()

		srv.HandleSync(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", rec.Code)
		}
	})
}
