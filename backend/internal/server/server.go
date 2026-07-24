package server

import (
	"context"
	"time"

	"github.com/havekes/price-tracker/internal/store"
	"github.com/havekes/price-tracker/internal/vision"
)

// Server holds the dependencies for HTTP handlers.
// Handlers are methods on *Server so they receive injected dependencies
// (e.g., store.Querier for database access).
type ReceiptProcessor interface {
	ProcessDirectUpload(ctx context.Context, imageBytes []byte, mimeType string, correspondentName string, purchaseDate time.Time) (*vision.ExtractedReceipt, error)
}

type Server struct {
	querier   store.Querier
	processor ReceiptProcessor
}

// New creates a new Server with the given dependencies.
func New(querier store.Querier, processor ReceiptProcessor) *Server {
	return &Server{querier: querier, processor: processor}
}
