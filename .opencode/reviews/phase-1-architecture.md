---
phase: 1
date: 2026-07-23
verdict: sound-with-concerns
---

# Phase 1 Architecture Review

## Summary
Phase 1 landed the monorepo split (`backend/` Go service + `frontend/` SvelteKit app), a runnable HTTP server with env config and `/api/health`, and a SvelteKit + Tailwind v4 + shadcn-svelte UI with a Vite `/api` dev proxy to `:8080`. The phase goal — "base repository structure, runtime environments, and communication boundaries between services" — is met cleanly, and every non-blocking carry-over from the two PR reviews was addressed (JSON `slog` request logging, `internal/server` test pattern, dedup'd SIGINT, `getEnvInt` warn, `go 1.26` pin). Verdict is **sound-with-concerns**: the foundation is solid, but two server defaults (HTTP timeouts, stateless handler signature) will need small refactors when Phase 3/4 introduce long-running LLM calls and dependency injection — both are ticket-sized, neither blocks Phase 2 grooming.

## Findings

### 1. HTTP server timeouts will cancel Phase 3/4 long-running pipelines [severity: risk]
**Observation:** `backend/cmd/server/main.go:51-58` sets `ReadTimeout: 10s`, `WriteTimeout: 15s`, and the chi `middleware.Timeout(30 * time.Second)` wraps every route (`main.go:30`). The Vision LLM endpoint (`VISION_API_BASE_URL=https://ai.havek.es/api`) is documented in PROJECT.md Phase 3 as taking receipt images and returning structured JSON; such calls routinely run 10–30s+. `POST /api/receipts/upload` (Phase 4) runs the same pipeline synchronously after a multipart read.
**Impact:** When Phase 3 wires the vision client and Phase 4 the upload route, the first end-to-end run will fail with `context deadline exceeded` / write-timeout resets that are hard to attribute. The global `middleware.Timeout` is also known to misbehave with streaming/large response bodies in chi.
**Recommendation:** When grooming Phase 3 and Phase 4 tickets, replace the global 30s chi timeout with per-route timeouts, raise `http.Server` `ReadTimeout`/`WriteTimeout` (or move uploads/LLM work off the request path — PROJECT.md already specifies `/api/sync` runs "asynchronously"). Do not carry the current defaults into the ingestion phase.

### 2. Stateless package-level handlers won't carry dependency injection [severity: concern]
**Observation:** `backend/internal/server/health.go` defines `func HealthHandler(w http.ResponseWriter, r *http.Request)` as a package-level function; `main.go:34` registers it directly. There is no `server.Server` struct or wiring point holding `*sql.DB` / sqlc queries / service clients.
**Impact:** Phase 2 (P2-T02 persistence) introduces the DB and Phase 4 adds service-backed handlers; the natural DI pattern (a `Server`/`Handlers` struct with fields) is not present, so handlers will either grow globals or force a refactor mid-phase. The current `func(w,r)` signature is fine for health but is not the shape later handlers want.
**Recommendation:** In Phase 2, introduce a small `server.Server` (or `Handlers`) struct that receives wired dependencies via constructor and convert `HealthHandler` to a method as the first instance — establishes the DI pattern before Phase 3/4 multiply it. Fits inside P2-T02 or a tiny follow-up ticket.

### 3. `.gitignore` does not cover the SQLite database path [severity: concern]
**Observation:** `.env.example` sets `DB_PATH=data/price-tracker.db` and `backend/db/` is already committed (empty, a Phase 2 placeholder). `.gitignore` covers `backend/bin/`, `.env`, and frontend dirs but not `backend/data/`, `*.db`, or the SQLite sidecars (`*.db-wal`, `*.db-shm`).
**Impact:** Phase 2's first migration run will create `backend/data/price-tracker.db` (plus `-wal`/`-shm` in WAL mode) inside the repo tree; without an ignore rule it is one `git add .` away from being committed.
**Recommendation:** Fold a `.gitignore` update into P2-T01 (schema) or P2-T02 (persistence) before the first migration is applied: ignore `backend/data/`, `*.db`, `*.db-wal`, `*.db-shm`.

### 4. `adapter-static` SPA mode locked in; prerendering deliberately deferred [severity: debt]
**Observation:** `frontend/vite.config.ts` uses `adapter-static` with `fallback: 'index.html'` (pure SPA, no SSR/prerender). P1-T02 review noted this is fine unless a route needs prerendering.
**Impact:** Inert for Phases 2–5 (dashboard + charts work client-side). Phase 6 (PWA) actually *prefers* `adapter-static` SPA, so this choice is forward-aligned, but it means no SSR fallback for SEO/share previews and every route is a client render.
**Recommendation:** No action now; record the decision in the Phase 6 PWA ticket so the prerender question is explicitly closed rather than re-litigated.

### 5. Doc/scaffold drift: root README + committed `.vscode` [severity: debt]
**Observation:** Root `README.md` structure block still lists `frontend/  # SvelteKit web app (coming in Phase 1)` — frontend now exists. `frontend/.vscode/{settings.json,extensions.json}` were committed by the SvelteKit scaffold (P1-T02 nit, not addressed).
**Impact:** Pure low-interest cosmetic debt; no functional cost.
**Recommendation:** Update the README structure block during the next doc-touching ticket; optionally gitignore `frontend/.vscode/`. Do not spin a dedicated ticket.

## What went well
- **Clean monorepo boundaries.** `backend/` and `frontend/` are fully independent at the repo root; dependencies are one-directional (`cmd/server` → `internal/{config,server}`; `server` does not import `config`). The two Phase 1 tickets were genuinely parallelizable with no merge conflict.
- **Every PR-review carry-over was addressed, not silently dropped.** JSON-only request logging via a custom `server.RequestLogging` middleware emitting `slog.LogAttrs` (`handler.go`); handler extracted to `internal/server/health.go` with `health_test.go` using `httptest`; redundant `syscall.SIGINT` removed with an explanatory comment (`main.go:51`); `getEnvInt` now `slog.Warn`s on parse failure (`config.go:42-49`); `go.mod` pins `go 1.26`.
- **Forward-looking config pre-staging.** `config.Config` and `.env.example` already declare `PaperlessBaseURL`, `PaperlessToken`, `VisionAPIBaseURL`, `VisionAPIKey`, and `DBPath` — Phase 2/3 tickets can wire clients without touching config plumbing or its tests.
- **Symmetric dev-proxy contract.** `/api` prefix → `http://localhost:8080` is agreed on both sides: `frontend/vite.config.ts` server.proxy, `frontend/README.md` "Dev proxy" section, the P1-T02 ticket notes, and `backend`'s `PORT` default 8080. The inter-service boundary the phase goal asked for is documented and tested via `/api/health`.
- **Green baseline + extensible test pattern.** `go vet ./...`, `go build ./...`, and `go test ./...` pass; `internal/server/health_test.go` sets the `httptest` pattern later handler tickets can copy.
- **Stack choices are current and coherent.** Svelte 5 runes mode + Tailwind v4 + shadcn-svelte "nova" style + `adapter-static` is a consistent, PWA-ready frontend baseline.

## Carry-over for next phase
- [ ] **P2-T01 or P2-T02 (persistence):** before running the first migration, add `backend/data/`, `*.db`, `*.db-wal`, `*.db-shm` to `.gitignore`; confirm migrations live under `backend/db/` (placeholder dir already exists).
- [ ] **P2-T02 (persistence) or a small Phase 2 refactor ticket:** introduce a `server.Server`/`Handlers` struct receiving wired deps (`*sql.DB`, sqlc queries) via constructor; convert `HealthHandler` to a method as the first instance so Phase 4 handlers inherit the DI shape.
- [ ] **Phase 3 vision-LLM ticket:** replace the global `chi middleware.Timeout(30s)` with per-route timeouts and raise/relax `http.Server` `ReadTimeout`/`WriteTimeout` (or move LLM work off the request path) — current defaults will cancel vision calls.
- [ ] **Phase 4 upload ticket:** same timeout review for `POST /api/receipts/upload`; prefer the async pattern PROJECT.md already specifies for `/api/sync`.
- [ ] **Phase 2:** add a `config_test.go` covering `Load` defaults + env overrides to extend the established test pattern beyond handlers.
- [ ] **Grooming nit:** update root `README.md` structure block (frontend exists); optionally gitignore `frontend/.vscode/`. Fold into the next ticket that touches docs — no standalone ticket.

## Prior carry-over disposition
*Omitted — Phase 1 is the first phase; no prior architecture review report exists. The intra-phase carry-overs from PR reviews #5 and #6 are tracked under "What went well" above (all addressed except the optional `.vscode` nit).*
