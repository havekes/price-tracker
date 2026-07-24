---
id: P2-T02
phase: 2
title: Implement sqlc persistence & migration layer
status: pending
depends_on: [P1-T01, P2-T03]
branch: feat/p2-t02-sqlc-persistence
pr: null
source: "PROJECT.md → Phase 2 → Task 2.2"
---

## Objective

Provide type-safe database access for all six entities via `sqlc`, and auto-apply
the schema on backend startup so the service is self-initializing. This is the
data-access contract that Phase 3 (ingestion orchestrator) and Phase 4 (REST API)
build on.

## Scope

**In scope:**
- `sqlc.yaml` config targeting SQLite
- SQL query files under `backend/db/queries/*.sql` covering full CRUD for: `correspondent`, `receipt`, `product`, `raw_item`, `price_record`, `marketplace_link`
- Generated Go code committed under `backend/internal/db/` (or `backend/gen/db/`)
- SQLite driver wired into the backend: pure-Go `modernc.org/sqlite` (no CGO) recommended
- Migration auto-apply on startup: `schema.sql` embedded via `go:embed` and executed against the configured `DB_PATH` when tables are missing (idempotent)
- `internal/store` (or similar) package exposing a `Querier` interface + concrete implementation
- Unit tests for each entity's CRUD using an in-memory/temp-file SQLite DB
- Transaction helper (`WithTx`) exposed for the Phase 3 orchestrator

**Out of scope:**
- Business logic / ingestion orchestration (Phase 3)
- REST handlers (Phase 4)
- Multi-database support (SQLite only)

## Acceptance criteria

- [ ] `cd backend && sqlc generate` succeeds and produces type-safe Go (committed)
- [ ] Backend on startup creates the schema in `DB_PATH` if not present (idempotent across restarts)
- [ ] CRUD functions exist and pass tests for every entity (create, get, list, update where applicable, delete where applicable)
- [ ] `go test ./...` passes
- [ ] `go build ./...` succeeds; no CGO required (builds on darwin/arm64 without CGO flags)
- [ ] A `Querier` interface is defined so handlers/orchestrators can depend on the interface, not the concrete type
- [ ] A `WithTx(tx) Querier` (or equivalent) helper exists, enabling atomic multi-step writes for Phase 3
- [ ] Tests use a throwaway temp SQLite file (or `:memory:`) and clean up

## Technical notes

- Depends on P1-T01 (backend runtime + config, now `DATABASE_URL` instead of `DB_PATH`) and P2-T03 (`backend/db/schema.sql` in Postgres dialect + running Postgres via docker-compose).
- Driver: `github.com/jackc/pgx/v5` (pure Go, CGO-free, the standard Postgres driver) — NOT modernc.org/sqlite. The datastore pivoted from SQLite to Postgres in P2-T03.
- sqlc config: use the `postgres` engine in `sqlc.yaml` (NOT `sqlite`). Verify the installed `sqlc` version supports the postgres engine (v1.25+).
- Embed schema: `//go:embed db/schema.sql` in the store package and run it on startup against the configured `DATABASE_URL` (Postgres) when tables are missing (check `information_schema.tables`). P2-T03's docker-compose auto-loads schema.sql on first init for dev; this Go-side runner handles non-dev/prod starts.
- Keep query files one-entity-per-file for reviewability.
- The transaction helper is the explicit seam for Phase 3's "atomic ingestion orchestrator" — make sure it actually wraps a single `*sql.Tx`.

## Review feedback

