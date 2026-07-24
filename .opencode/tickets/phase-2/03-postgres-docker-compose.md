---
id: P2-T03
phase: 2
title: Migrate persistence from SQLite to Postgres with docker-compose dev environment
status: pending
depends_on: [P1-T01]
branch: feat/p2-t03-postgres-docker-compose
pr: null
source: "Architecture pivot — supersedes P2-T01 (SQLite schema, PR #7 closed)"
---

## Objective

Replace the embedded-SQLite persistence approach with PostgreSQL running in a
docker-compose dev environment, and convert the relational schema to Postgres
dialect. This unblocks P2-T02 (sqlc + pgx persistence) and gives every later
phase a real database service to develop against. Supersedes P2-T01, whose
SQLite schema design (table shapes, FK semantics, unit-normalization columns,
marketplace_link bidirectional uniqueness) carries forward — only the dialect
and runtime change.

## Scope

**In scope:**
- `docker-compose.yml` at repo root running Postgres 16 (alpine) with:
  - env-based credentials (`POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB`)
  - a named volume for data persistence
  - a `pg_isready` healthcheck
  - port `5432` exposed for dev
  - `schema.sql` mounted into `/docker-entrypoint-initdb.d/` so Postgres auto-loads it on first init
- `backend/db/schema.sql` converted from SQLite to **Postgres dialect**:
  - Replace `INTEGER PRIMARY KEY` with `BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY` (modern Postgres)
  - Replace `REAL` with `DOUBLE PRECISION`
  - Use `TIMESTAMPTZ NOT NULL DEFAULT now()` for `created_at`/`updated_at`
  - Keep `DATE` for `receipt.purchased_at`
  - Keep `TEXT` (valid in Postgres) for string columns
  - **Remove** `PRAGMA foreign_keys = ON;` (Postgres enforces FKs by default)
  - Keep `CREATE TABLE IF NOT EXISTS` (valid in Postgres, idempotent)
  - Keep all table/column names identical to P2-T01's design (downstream contract for Phase 3/4)
  - Keep all FK `ON DELETE` semantics, `UNIQUE`, `CHECK(product_a_id < product_b_id)`, indexes, and SQL comments
- Config changes in `backend/internal/config`:
  - Replace `DB_PATH` with `DATABASE_URL` (DSN), default `postgres://price-tracker:price-tracker@localhost:5432/price-tracker?sslmode=disable`
- `.env.example` updated: replace `DB_PATH` with `DATABASE_URL` (document the docker-compose defaults)
- `backend/db/README.md` updated:
  - Remove the SQLite "Foreign-Key Enforcement (Per-Connection)" section (Postgres enforces FKs always)
  - Add a "Dev Database (docker-compose)" section documenting `docker compose up -d`, `docker compose down -v`, and how to connect via `psql`
  - Keep the ER diagram, FK semantics table, unit-normalization convention, timestamp format table
- Root `.gitignore`: remove any SQLite-specific entries (none added yet, but ensure no `*.db`/`data/` cruft); the Postgres data lives in the docker named volume (not committed)

**Out of scope:**
- Go database driver / connection pool wiring (P2-T02 — pgx)
- sqlc codegen / CRUD / migrations runner in Go (P2-T02)
- Production deployment config (dev-only docker-compose here)

## Acceptance criteria

- [ ] `docker compose up -d` starts Postgres; `docker compose ps` shows the db service healthy
- [ ] `docker compose down -v` cleans up the volume
- [ ] Postgres auto-loads `schema.sql` on first init (tables exist without manual `psql -f`)
- [ ] `docker compose exec db psql -U price-tracker -d price-tracker -c "\dt"` lists all six tables
- [ ] FK enforcement verified: `INSERT INTO receipt (correspondent_id, ...) VALUES (999, ...)` fails with FK violation (no PRAGMA needed)
- [ ] `marketplace_link` bidirectional uniqueness + CHECK enforced (same tests as P2-T01)
- [ ] Unit-normalization columns present on `product` and `price_record`
- [ ] Every table and non-obvious column has a SQL comment
- [ ] `backend/db/schema.sql` loads cleanly via `psql -f` against a fresh Postgres DB (idempotent on re-run)
- [ ] `DATABASE_URL` is the config key; `.env.example` documents it with docker-compose defaults
- [ ] `backend/db/README.md` documents docker-compose usage and connection; SQLite PRAGMA section removed
- [ ] `cd backend && go build ./...` still succeeds (config package compiles with new `DATABASE_URL`)
- [ ] `cd backend && go vet ./...` clean; `go test ./...` passes

## Technical notes

- Supersedes P2-T01 (PR #7, closed without merge). The schema DESIGN from P2-T01 is the contract — only the dialect changes. Keep table/column names identical so Phase 3/4 contracts hold.
- Postgres 16-alpine is the recommended image (current stable, small).
- Identity columns: `BIGINT GENERATED ALWAYS AS IDENTITY` is the modern Postgres recommendation (SQL-standard, supersedes `SERIAL`/`BIGSERIAL`).
- docker-compose init: mounting `schema.sql` into `/docker-entrypoint-initdb.d/` runs it only on first volume init (empty data dir). For schema changes mid-dev, `docker compose down -v && docker compose up -d` re-initializes. Document this.
- DSN default uses `sslmode=disable` for local dev; production would change this.
- This ticket is the seam for P2-T02: the running Postgres + Postgres-dialect schema.sql is what sqlc (postgres engine) and pgx will target. P2-T02 will add the Go connection pool, codegen, CRUD, and a Go-side migration runner (embed schema.sql + exec, or golang-migrate).
- Phase 1 arch review carry-over resolved here: no more `DB_PATH`/`*.db` gitignore concerns; Postgres data lives in a docker volume.
- Do NOT wire up the Go database connection or sqlc here — that's P2-T02. This ticket is infrastructure + schema dialect + config.

## Review feedback
