package paperless

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewClient(t *testing.T) {
	baseURL := "http://example.com"
	token := "secret-token"

	client := NewClient(baseURL, token, nil)
	if client.baseURL != baseURL {
		t.Errorf("expected baseURL %q, got %q", baseURL, client.baseURL)
	}
	if client.authToken != token {
		t.Errorf("expected token %q, got %q", token, client.authToken)
	}
	if client.httpClient != http.DefaultClient {
		t.Errorf("expected default http client when nil is provided")
	}

	customClient := &http.Client{}
	clientWithCustom := NewClient(baseURL, token, customClient)
	if clientWithCustom.httpClient != customClient {
		t.Errorf("expected custom http client to be set")
	}
}

func setupTestServer(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *Client) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Verify auth header on all requests
		if r.Header.Get("Authorization") != "Token test-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		handler(w, r)
	})

	server := httptest.NewServer(mux)
	client := NewClient(server.URL, "test-token", server.Client())
	return server, client
}

func TestGetDocuments(t *testing.T) {
	server, client := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/documents/" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		
		corrID := 2
		resp := DocumentListResponse{
			Count: 1,
			Results: []Document{
				{
					ID:              1,
					Title:           "Test Doc",
					Created:         "2023-01-01T00:00:00Z",
					CorrespondentID: &corrID,
					Tags:            []int{10, 20},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	corrID := 2
	params := GetDocumentsParams{
		Tags:            []int{10},
		CorrespondentID: &corrID,
	}

	docs, err := client.GetDocuments(context.Background(), params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(docs) != 1 {
		t.Fatalf("expected 1 document, got %d", len(docs))
	}
	if docs[0].ID != 1 {
		t.Errorf("expected document ID 1, got %d", docs[0].ID)
	}
	if docs[0].Title != "Test Doc" {
		t.Errorf("expected title 'Test Doc', got %q", docs[0].Title)
	}
}

func TestGetDocument(t *testing.T) {
	server, client := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/documents/1/" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		
		doc := Document{
			ID:    1,
			Title: "Test Doc",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(doc)
	})
	defer server.Close()

	doc, err := client.GetDocument(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if doc == nil {
		t.Fatal("expected non-nil document")
	}
	if doc.ID != 1 {
		t.Errorf("expected document ID 1, got %d", doc.ID)
	}
}

func TestGetCorrespondent(t *testing.T) {
	server, client := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/correspondents/2/" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		
		corr := Correspondent{
			ID:   2,
			Name: "ACME Corp",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(corr)
	})
	defer server.Close()

	corr, err := client.GetCorrespondent(context.Background(), 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if corr == nil {
		t.Fatal("expected non-nil correspondent")
	}
	if corr.Name != "ACME Corp" {
		t.Errorf("expected correspondent name 'ACME Corp', got %q", corr.Name)
	}
}

func TestDownloadDocument(t *testing.T) {
	server, client := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/documents/1/download/" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		
		w.Header().Set("Content-Type", "application/pdf")
		w.Write([]byte("fake-pdf-content"))
	})
	defer server.Close()

	content, contentType, err := client.DownloadDocument(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if contentType != "application/pdf" {
		t.Errorf("expected content type 'application/pdf', got %q", contentType)
	}
	if string(content) != "fake-pdf-content" {
		t.Errorf("expected content 'fake-pdf-content', got %q", string(content))
	}
}

func TestErrors(t *testing.T) {
	server, client := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/401" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if r.URL.Path == "/404" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("server error"))
	})
	defer server.Close()

	// Test Unauthorized
	_, err := client.doRequest(context.Background(), "GET", "/401", nil)
	if err != ErrUnauthorized {
		t.Errorf("expected ErrUnauthorized, got %v", err)
	}

	// Test Not Found
	_, err = client.doRequest(context.Background(), "GET", "/404", nil)
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}

	// Test Other Error
	_, err = client.doRequest(context.Background(), "GET", "/500", nil)
	if err == nil {
		t.Error("expected error for 500 status, got nil")
	} else if err == ErrUnauthorized || err == ErrNotFound {
		t.Errorf("expected generic error, got %v", err)
	}
}
