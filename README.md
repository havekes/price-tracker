# Price Tracker

A receipt parsing and price tracking application — ingest receipts from Paperless-ngx, extract line items via a Vision LLM, and track price history across stores.

## Project Structure

```
.
├── backend/        # Go API server (cmd/server, internal/store, internal/server, etc.)
├── frontend/       # SvelteKit web app (Tailwind CSS, shadcn-svelte)
├── docker-compose.yml  # PostgreSQL 16 dev database
├── .env.example    # Environment variable reference
└── PROJECT.md      # Full implementation plan
```

## Getting Started

### Prerequisites

- Go 1.26+
- Node.js 22+ (for frontend)

### Backend

```bash
cd backend
go mod tidy
go run ./cmd/server
```

The server starts on `http://localhost:8080` (configurable via `PORT`).

### Configuration

Copy `.env.example` to `.env` and fill in your values:

```bash
cp .env.example .env
```

See `.env.example` for all supported environment variables.

## License

MIT
