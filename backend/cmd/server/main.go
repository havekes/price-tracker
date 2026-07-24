package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/jackc/pgx/v5/stdlib"

	dbembed "github.com/havekes/price-tracker/db"
	"github.com/havekes/price-tracker/internal/config"
	"github.com/havekes/price-tracker/internal/server"
	"github.com/havekes/price-tracker/internal/store"
)

func main() {
	// Set up structured JSON logging before config.Load so any warnings
	// from config parsing are already formatted as JSON.
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	cfg := config.Load()

	// Open a shared *sql.DB pool for the application.
	db, err := sql.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to open database", "error", err)
		os.Exit(1)
	}

	// Sensible pool defaults for a CLI-adjacent HTTP server.
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Apply database migration before starting the server.
	// Reuses the shared pool — does not open its own connection.
	if err := store.Migrate(db, dbembed.Schema); err != nil {
		slog.Error("database migration failed", "error", err)
		os.Exit(1)
	}

	// Build the dependency graph: *sql.DB → store.Querier → server.Server.
	querier := store.New(db)
	srv := server.New(querier)

	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(server.RequestLogging)
	r.Use(middleware.Recoverer)

	// NOTE: No global chi middleware.Timeout is used. HTTP timeouts are managed
	// at the http.Server level (ReadTimeout/WriteTimeout below). Routes that call
	// the Vision LLM (Phase 3.2+, e.g., /api/sync, /api/receipts/upload) should
	// use context.WithTimeout for per-call deadlines rather than a global middleware
	// timeout, which can misbehave with large/streaming request bodies.

	// Routes
	r.Get("/api/health", srv.HealthHandler)
	r.Get("/api/products", srv.ListProductsHandler)

	addr := fmt.Sprintf(":%d", cfg.Port)
	httpSrv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Graceful shutdown — os.Interrupt covers SIGINT on Unix; no need for syscall.SIGINT.
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)

	go func() {
		slog.Info("server starting", "addr", addr)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	<-done
	slog.Info("shutting down gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := httpSrv.Shutdown(ctx); err != nil {
		slog.Error("shutdown error", "error", err)
		os.Exit(1)
	}

	// Close the database pool after the HTTP server has stopped.
	if err := db.Close(); err != nil {
		slog.Error("database close error", "error", err)
		os.Exit(1)
	}
	slog.Info("server stopped")
}
