package server

import (
	"context"

	"github.com/havekes/price-tracker/internal/store"
)

type SyncPipeline interface {
	SyncPaperlessReceipts(context.Context, int) (int, int, error)
}

// Server holds the dependencies for HTTP handlers.
// Handlers are methods on *Server so they receive injected dependencies
// (e.g., store.Querier for database access).
type Server struct {
	querier  store.Querier
	pipeline SyncPipeline
}

// New creates a new Server with the given dependencies.
func New(querier store.Querier, pipeline SyncPipeline) *Server {
	return &Server{querier: querier, pipeline: pipeline}
}
