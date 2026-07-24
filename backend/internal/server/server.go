package server

import (
	"github.com/havekes/price-tracker/internal/store"
)

// Server holds the dependencies for HTTP handlers.
// Handlers are methods on *Server so they receive injected dependencies
// (e.g., store.Querier for database access).
type Server struct {
	querier store.Querier
}

// New creates a new Server with the given dependencies.
func New(querier store.Querier) *Server {
	return &Server{querier: querier}
}
