# Price Tracker — Backend

Go backend service for the price-tracker application.

## Prerequisites

- Go 1.26+
- Docker Compose with a running Postgres 16 container

## Quick Start

```bash
# Start Postgres
docker compose up -d

# Run the server
go run ./cmd/server/

# The health endpoint
curl http://localhost:8080/api/health
```

## Database

The project uses **PostgreSQL 16** with the schema defined in `db/schema.sql`.
On startup the server automatically applies the schema if the `correspondent`
table does not exist (idempotent — safe to run on every restart).

See [db/README.md](./db/README.md) for schema details and entity relationships.

## Running Tests

Tests connect to a real Postgres instance defined by the `DATABASE_URL`
environment variable (default: `postgres://price-tracker:price-tracker@localhost:5433/price-tracker?sslmode=disable`).

**The Postgres container must be running** for tests to work:

```bash
# From the repo root
docker compose up -d

# Run all backend tests
go test ./...

# Run specific package tests
go test ./internal/store/... -v
```

Each test creates a throwaway Postgres schema (prefixed with `test_`) and drops
it on cleanup, so tests are isolated and leave no state behind.

## Project Layout

```
cmd/server/main.go        — Application entrypoint
db/                       — Schema and SQL query files
  schema.sql              — DDL (6 tables)
  queries/                — sqlc query files (one per entity)
internal/
  config/                 — Environment-based configuration
  db/                     — sqlc-generated Go code (type-safe queries)
  server/                 — HTTP handlers and middleware
  store/                  — Database interface (Querier) and migration
```

## Code Generation

SQL queries are code-generated with [sqlc](https://sqlc.dev/):

```bash
sqlc generate
# Output: internal/db/*.go
```

The generated Go code is committed to the repository.
