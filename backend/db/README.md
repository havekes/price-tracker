# Database Schema

The price-tracker project uses **PostgreSQL 16** running in docker-compose for local
development. The schema is loaded automatically on first container init via
`/docker-entrypoint-initdb.d/schema.sql`.

## Dev Database (docker-compose)

```bash
# Start Postgres
docker compose up -d

# Verify health
docker compose ps
# → db service shows "healthy"

# Connect via psql
docker compose exec db psql -U price-tracker -d price-tracker

# List tables
docker compose exec db psql -U price-tracker -d price-tracker -c "\dt"

# Stop and remove data volume
docker compose down -v

# Restart with fresh schema
docker compose up -d
```

> **Note:** `schema.sql` runs only on first volume init (empty data directory).
> For schema changes during active development, use `docker compose down -v && docker compose up -d`
> to re-initialize.

## Entity-Relationship Diagram

```
┌──────────────┐       ┌──────────────┐
│  correspondent│       │   product    │
│──────────────│       │──────────────│
│ id           │──┐    │ id           │──┐
│ name         │  │    │ display_name │  │
│ created_at   │  │    │ base_unit    │  │
└──────────────┘  │    │ created_at   │  │
                  │    └──────────────┘  │
                  │         │            │
                  │         │ 1          │ 1
                  │         ▼            ▼
                  │    ┌──────────────┐  │  ┌──────────────────┐
                  │    │   raw_item   │  │  │ marketplace_link │
                  │    │──────────────│  │  │──────────────────│
                  │    │ id           │  │  │ id               │
                  └────│ receipt_id   │  │  │ product_a_id     │◄──┘
                       │ product_id   │──┘  │ product_b_id     │◄────┘
                       │ raw_text     │      │ created_at       │
                       │ raw_quantity │      │ UNIQUE(a_id,b_id)│
                       │ quantity_val │      │ CHECK(a_id<b_id) │
                       │ quantity_unit│      └──────────────────┘
                       │ created_at   │
                       └──────┬───────┘
                              │ 1
                              ▼
                    ┌────────────────┐
                    │  price_record  │
                    │────────────────│
                    │ id             │
                    │ raw_item_id    │ (1:1 UNIQUE)
                    │ total_price    │
                    │ unit_price     │
                    │ currency       │
                    │ created_at     │
                    └────────────────┘
```

## Foreign-Key Semantics

| FK Column(s)                        | Parent         | ON DELETE  | Rationale                                                                 |
|-------------------------------------|----------------|------------|---------------------------------------------------------------------------|
| `receipt.correspondent_id`          | `correspondent`| RESTRICT   | Prevent deleting a correspondent who still has receipts                   |
| `raw_item.receipt_id`               | `receipt`      | CASCADE    | Deleting a receipt removes its line items                                 |
| `raw_item.product_id`               | `product`      | SET NULL   | Removing a product leaves the raw item intact but unlinked                |
| `price_record.raw_item_id`          | `raw_item`     | CASCADE    | Deleting a raw item removes its price record                              |
| `marketplace_link.product_a_id`     | `product`      | CASCADE    | Removing a product cleans up its marketplace links                        |
| `marketplace_link.product_b_id`     | `product`      | CASCADE    | Removing a product cleans up its marketplace links                        |

> **Postgres enforces foreign keys always** — no per-connection PRAGMA or DSN
> parameter needed.

## Unit-Normalization Convention

Price normalization enables meaningful comparisons across different package sizes
and units (e.g., $/100g vs $/oz). The schema stores both the verbatim text and
the parsed numeric values so downstream consumers can compare or re-derive.

| Column                        | Table          | Purpose                                                      |
|-------------------------------|----------------|--------------------------------------------------------------|
| `raw_item.raw_quantity`       | `raw_item`     | Verbatim quantity text from the receipt (e.g., `"1 lb"`)     |
| `raw_item.quantity_value`     | `raw_item`     | Parsed numeric quantity (e.g., `1.0`)                        |
| `raw_item.quantity_unit`      | `raw_item`     | Normalised unit string (e.g., `lb`)                          |
| `product.base_unit`           | `product`      | Canonical unit for price comparison (e.g., `g`, `kg`, `ml`)  |
| `price_record.unit_price`     | `price_record` | Precomputed price per base_unit (populated by ingestion)      |

## Timestamp Format

All `created_at` columns use `TIMESTAMPTZ` (timestamp with time zone) with a
default of `now()`. `receipt.purchased_at` is `DATE` (date-only — no time
component).

| Column                        | Type         | Purpose                              |
|-------------------------------|--------------|--------------------------------------|
| `*.created_at`                | `TIMESTAMPTZ`| Row creation timestamp (millisecond) |
| `receipt.purchased_at`        | `DATE`       | Purchase date from the receipt       |

## Notes

- Table/column names are the contract for Phase 3 (ingestion) and Phase 4 (API).
  Any changes must be coordinated across phases.
- The `CHECK(product_a_id < product_b_id)` constraint on `marketplace_link`
  prevents duplicate pair orderings and self-links.
