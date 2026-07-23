---
id: P1-T01
phase: 1
title: Scaffold Go backend with routing & env config
status: pending
depends_on: []
branch: ao/p1-t01-go-backend-scaffold
pr: null
source: "PROJECT.md → Phase 1 → Task 1.1"
---

## Objective

Establish the monorepo root layout and a runnable Go backend service with HTTP
routing and environment-based configuration. This is the runtime foundation that
every later backend ticket (persistence, ingestion, API) builds on.

## Scope

**In scope:**
- Monorepo root structure: `backend/` directory plus root files (README, comprehensive root `.gitignore` covering Go + Node)
- Go module (`backend/go.mod`) with a lightweight HTTP router (recommend `chi`)
- Configuration package loading from environment variables with sensible defaults (e.g. `PORT`, `PAPERLESS_BASE_URL`, `PAPERLESS_TOKEN`, `VISION_API_BASE_URL`, `VISION_API_KEY`, `DB_PATH`)
- `.env.example` documenting all variables
- Health check route `GET /api/health` returning JSON `{"status":"ok"}`
- Graceful shutdown on SIGTERM/SIGINT
- Structured logging (recommend `log/slog` from stdlib)

**Out of scope:**
- Database connections and migrations (Phase 2, P2-T02)
- Any business-logic routes (Phase 3/4)
- Frontend (P1-T02 creates `frontend/`)

## Acceptance criteria

- [ ] `cd backend && go build ./...` succeeds with no errors
- [ ] `go run ./cmd/server` (or equivalent) starts and listens on configured `PORT` (default 8080)
- [ ] `GET /api/health` returns `200` with JSON body `{"status":"ok"}`
- [ ] Environment variables are loaded; missing required vars produce a clear startup error (required set is empty for now aside from `PORT` which has a default)
- [ ] `.env.example` lists every recognized variable with a comment
- [ ] SIGTERM/SIGINT triggers a clean shutdown (log line confirms server stopped)
- [ ] Root `.gitignore` ignores `backend/bin/`, Go build artifacts, `.env`, and Node dirs
- [ ] `go vet ./...` is clean

## Technical notes

- Router choice: `chi` (github.com/go-chi/chi/v5) — idiomatic, lightweight, middleware-friendly. Stdlib `net/http` is also acceptable if deps must stay minimal; pick one and note it.
- Config: prefer a small `internal/config` package returning a typed struct. `godotenv` is optional for local `.env` loading; env vars already present in the process work without it.
- Do NOT create a `frontend/` directory — P1-T02 owns that and the two tickets are parallelizable.
- This ticket defines the monorepo root; the root `.gitignore` and README land here so P1-T02 doesn't have to.
- Keep the dependency list small. Phase 2 will add the SQLite driver and sqlc.

## Review feedback

