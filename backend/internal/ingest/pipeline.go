package ingest

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/havekes/price-tracker/internal/paperless"
	"github.com/havekes/price-tracker/internal/store"
	"github.com/havekes/price-tracker/internal/vision"
)

type PaperlessClient interface {
	GetDocument(ctx context.Context, docID int) (*paperless.Document, error)
	GetCorrespondent(ctx context.Context, correspondentID int) (*paperless.Correspondent, error)
	DownloadDocument(ctx context.Context, docID int) ([]byte, string, error)
	GetDocuments(ctx context.Context, params paperless.GetDocumentsParams) ([]paperless.Document, error)
}

type VisionClient interface {
	ExtractReceipt(ctx context.Context, imageBytes []byte, mimeType string) (*vision.ExtractedReceipt, error)
}

type Pipeline struct {
	paperless PaperlessClient
	vision    VisionClient
	store     store.Querier
	db        *sql.DB
}

func NewPipeline(p PaperlessClient, v VisionClient, s store.Querier, db *sql.DB) *Pipeline {
	return &Pipeline{
		paperless: p,
		vision:    v,
		store:     s,
		db:        db,
	}
}

func (p *Pipeline) ProcessPaperlessDocument(ctx context.Context, docID int) error {
	doc, err := p.paperless.GetDocument(ctx, docID)
	if err != nil {
		return fmt.Errorf("failed to get document %d: %w", docID, err)
	}

	correspondentName := "Unknown"
	if doc.CorrespondentID != nil {
		corr, err := p.paperless.GetCorrespondent(ctx, *doc.CorrespondentID)
		if err == nil {
			correspondentName = corr.Name
		} else {
			log.Printf("Warning: failed to get correspondent %d for document %d: %v", *doc.CorrespondentID, docID, err)
		}
	}

	docBytes, mimeType, err := p.paperless.DownloadDocument(ctx, docID)
	if err != nil {
		return fmt.Errorf("failed to download document %d: %w", docID, err)
	}

	receipt, err := p.vision.ExtractReceipt(ctx, docBytes, mimeType)
	if err != nil {
		return fmt.Errorf("failed to extract receipt for document %d: %w", docID, err)
	}

	purchasedAt, err := time.Parse(time.RFC3339, doc.Created)
	if err != nil {
		// Fallback to now if missing or unparseable
		purchasedAt = time.Now()
	}

	extDocID := strconv.Itoa(docID)

	items := make([]IngestItemInput, len(receipt.Items))
	for i, item := range receipt.Items {
		items[i] = IngestItemInput{
			RawText:     item.RawText,
			DisplayName: item.SuggestedProductName,
			RawQuantity: fmt.Sprintf("%g", item.RawQuantity),
			BaseUnit:    item.SuggestedBaseUnit,
			TotalPrice:  item.TotalPrice,
			Currency:    "USD",
		}
	}

	input := IngestInput{
		CorrespondentName: correspondentName,
		PurchasedAt:       purchasedAt,
		Source:            "paperless",
		ExternalDocID:     &extDocID,
		RawFileRef:        nil, // We don't have a raw file ref stored anywhere yet
		Items:             items,
	}

	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	if err := PersistReceipt(ctx, p.store, tx, input); err != nil {
		return fmt.Errorf("failed to persist receipt for document %d: %w", docID, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (p *Pipeline) ProcessDirectUpload(ctx context.Context, imageBytes []byte, mimeType string, correspondentName string, purchaseDate time.Time) (*vision.ExtractedReceipt, error) {
	receipt, err := p.vision.ExtractReceipt(ctx, imageBytes, mimeType)
	if err != nil {
		return nil, fmt.Errorf("failed to extract receipt from upload: %w", err)
	}

	items := make([]IngestItemInput, len(receipt.Items))
	for i, item := range receipt.Items {
		items[i] = IngestItemInput{
			RawText:     item.RawText,
			DisplayName: item.SuggestedProductName,
			RawQuantity: fmt.Sprintf("%g", item.RawQuantity),
			BaseUnit:    item.SuggestedBaseUnit,
			TotalPrice:  item.TotalPrice,
			Currency:    "USD",
		}
	}

	input := IngestInput{
		CorrespondentName: correspondentName,
		PurchasedAt:       purchaseDate,
		Source:            "direct_upload",
		ExternalDocID:     nil,
		RawFileRef:        nil,
		Items:             items,
	}

	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	if err := PersistReceipt(ctx, p.store, tx, input); err != nil {
		return nil, fmt.Errorf("failed to persist receipt for upload: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return receipt, nil
}

func (p *Pipeline) SyncPaperlessReceipts(ctx context.Context, tag int) (successCount int, failureCount int, err error) {
	docs, err := p.paperless.GetDocuments(ctx, paperless.GetDocumentsParams{
		Tags: []int{tag},
	})
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get documents: %w", err)
	}

	for _, doc := range docs {
		if err := p.ProcessPaperlessDocument(ctx, doc.ID); err != nil {
			log.Printf("Error processing document %d: %v", doc.ID, err)
			failureCount++
		} else {
			successCount++
		}
	}

	return successCount, failureCount, nil
}
