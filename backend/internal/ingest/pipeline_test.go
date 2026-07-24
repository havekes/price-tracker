package ingest

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	internaldb "github.com/havekes/price-tracker/internal/db"
	"github.com/havekes/price-tracker/internal/paperless"
	"github.com/havekes/price-tracker/internal/store"
	"github.com/havekes/price-tracker/internal/vision"
)

type mockPaperlessClient struct {
	doc     *paperless.Document
	corr    *paperless.Correspondent
	content []byte
	mime    string
	docs    []paperless.Document
	err     error
}

func (m *mockPaperlessClient) GetDocument(ctx context.Context, docID int) (*paperless.Document, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.doc, nil
}
func (m *mockPaperlessClient) GetCorrespondent(ctx context.Context, id int) (*paperless.Correspondent, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.corr, nil
}
func (m *mockPaperlessClient) DownloadDocument(ctx context.Context, docID int) ([]byte, string, error) {
	if m.err != nil {
		return nil, "", m.err
	}
	return m.content, m.mime, nil
}
func (m *mockPaperlessClient) GetDocuments(ctx context.Context, params paperless.GetDocumentsParams) ([]paperless.Document, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.docs, nil
}

type mockVisionClient struct {
	receipt *vision.ExtractedReceipt
	err     error
}

func (m *mockVisionClient) ExtractReceipt(ctx context.Context, imageBytes []byte, mimeType string) (*vision.ExtractedReceipt, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.receipt, nil
}

type mockStore struct {
	store.Querier
	persistCalls int
	err error
}

func (m *mockStore) GetCorrespondentByName(ctx context.Context, name string) (internaldb.Correspondent, error) {
	if m.err != nil {
		return internaldb.Correspondent{}, m.err
	}
	return internaldb.Correspondent{ID: 1, Name: name}, nil
}
func (m *mockStore) CreateReceipt(ctx context.Context, arg internaldb.CreateReceiptParams) (internaldb.Receipt, error) {
	if m.err != nil {
		return internaldb.Receipt{}, m.err
	}
	return internaldb.Receipt{ID: 1}, nil
}
func (m *mockStore) GetProductByName(ctx context.Context, name string) (internaldb.Product, error) {
	if m.err != nil {
		return internaldb.Product{}, m.err
	}
	return internaldb.Product{ID: 1, DisplayName: name}, nil
}
func (m *mockStore) CreateRawItem(ctx context.Context, arg internaldb.CreateRawItemParams) (internaldb.RawItem, error) {
	if m.err != nil {
		return internaldb.RawItem{}, m.err
	}
	return internaldb.RawItem{ID: 1}, nil
}
func (m *mockStore) CreatePriceRecord(ctx context.Context, arg internaldb.CreatePriceRecordParams) (internaldb.PriceRecord, error) {
	if m.err != nil {
		return internaldb.PriceRecord{}, m.err
	}
	m.persistCalls++
	return internaldb.PriceRecord{ID: 1}, nil
}
func (m *mockStore) WithTx(tx *sql.Tx) store.Querier {
	return m
}

func TestProcessPaperlessDocument(t *testing.T) {
	database, cleanup := testDB(t)
	defer cleanup()

	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		corrID := 1
		pClient := &mockPaperlessClient{
			doc: &paperless.Document{
				ID:              123,
				Title:           "Test Receipt",
				Created:         "2026-07-24T00:00:00Z",
				CorrespondentID: &corrID,
			},
			corr: &paperless.Correspondent{
				ID:   1,
				Name: "Test Store",
			},
			content: []byte("fake-pdf-content"),
			mime:    "application/pdf",
		}
		vClient := &mockVisionClient{
			receipt: &vision.ExtractedReceipt{
				Items: []vision.ExtractedItem{
					{
						RawText:              "1 item",
						TotalPrice:           10.5,
						RawQuantity:          1.0,
						SuggestedProductName: "Item",
						SuggestedBaseUnit:    "unit",
					},
				},
			},
		}
		mStore := &mockStore{}

		pipeline := NewPipeline(pClient, vClient, mStore, database)

		err := pipeline.ProcessPaperlessDocument(ctx, 123)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if mStore.persistCalls != 1 {
			t.Errorf("expected 1 persist call, got %d", mStore.persistCalls)
		}
	})

	t.Run("paperless failure", func(t *testing.T) {
		pClient := &mockPaperlessClient{
			err: errors.New("paperless error"),
		}
		pipeline := NewPipeline(pClient, &mockVisionClient{}, &mockStore{}, database)
		err := pipeline.ProcessPaperlessDocument(ctx, 123)
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
	})
}

func TestProcessDirectUpload(t *testing.T) {
	database, cleanup := testDB(t)
	defer cleanup()

	ctx := context.Background()

	vClient := &mockVisionClient{
		receipt: &vision.ExtractedReceipt{
			Items: []vision.ExtractedItem{
				{
					RawText:              "1 item",
					TotalPrice:           10.5,
					RawQuantity:          1.0,
					SuggestedProductName: "Item",
					SuggestedBaseUnit:    "unit",
				},
			},
		},
	}
	mStore := &mockStore{}

	pipeline := NewPipeline(&mockPaperlessClient{}, vClient, mStore, database)

	receipt, err := pipeline.ProcessDirectUpload(ctx, []byte("img"), "image/jpeg", "Store", time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if receipt == nil {
		t.Fatalf("expected receipt, got nil")
	}
	if mStore.persistCalls != 1 {
		t.Errorf("expected 1 persist call, got %d", mStore.persistCalls)
	}
}



type countingVisionClient struct {
	calls int
}

func (c *countingVisionClient) ExtractReceipt(ctx context.Context, imageBytes []byte, mimeType string) (*vision.ExtractedReceipt, error) {
	c.calls++
	if c.calls == 2 {
		return nil, errors.New("simulated vision error on second call")
	}
	return &vision.ExtractedReceipt{
		Items: []vision.ExtractedItem{
			{
				RawText:              "Item 1",
				TotalPrice:           10.0,
				RawQuantity:          1.0,
				SuggestedProductName: "Item 1",
				SuggestedBaseUnit:    "unit",
			},
		},
	}, nil
}

func TestSyncPaperlessReceipts(t *testing.T) {
	database, cleanup := testDB(t)
	defer cleanup()

	ctx := context.Background()

	pClient := &mockPaperlessClient{
		docs: []paperless.Document{
			{ID: 1, Created: "2026-07-24T00:00:00Z"},
			{ID: 2, Created: "2026-07-24T00:00:00Z"},
		},
		doc: &paperless.Document{
			ID:      1,
			Created: "2026-07-24T00:00:00Z",
		},
		content: []byte("fake"),
		mime:    "application/pdf",
	}

	vClient := &countingVisionClient{}
	mStore := &mockStore{}

	pipeline := NewPipeline(pClient, vClient, mStore, database)

	success, failure, err := pipeline.SyncPaperlessReceipts(ctx, 123)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if success != 1 {
		t.Errorf("expected 1 success, got %d", success)
	}
	if failure != 1 {
		t.Errorf("expected 1 failure, got %d", failure)
	}
}
