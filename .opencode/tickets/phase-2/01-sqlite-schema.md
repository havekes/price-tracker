---
id: P2-T01
phase: 2
title: Design relational SQLite schema for receipts & prices
status: pending
depends_on: [P1-T01]
branch: feat/p2-t01-sqlite-schema
pr: null
source: "PROJECT.md → Phase 2 → Task 2.1"
---

## Objective

Define the normalized SQLite schema that models correspondents, receipts,
canonical products, verbatim receipt items, price records, and cross-marketplace
links — including unit-normalization columns. This `schema.sql` is the canonical
contract that P2-T02 generates type-safe code from and that Phase 3 writes into.

## Scope

**In scope:**
- `backend/db/schema.sql` — fully annotated (SQL comments per table/column)
- Tables: `correspondent`, `receipt`, `product`, `raw_item`, `price_record`, `marketplace_link`
- Primary keys, foreign keys (with `ON DELETE` semantics chosen deliberately), `NOT NULL`, `UNIQUE` constraints
- Indexes on hot lookup paths (e.g. product by display_name, price_record by raw_item, receipt by correspondent/purchased_at)
- Unit-normalization columns: `product.base_unit` (e.g. `g`, `kg`, `ml`, `l`, `unit`), `price_record.unit_price` (precomputed normalized price per base unit), `raw_item.raw_quantity` (verbatim string) and a parsed numeric quantity column
- Timestamps (`created_at`, relevant `updated_at`) with sensible defaults
- `PRAGMA foreign_keys = ON;` at top of file
- A `backend/db/README.md` documenting the schema and the unit-normalization convention

**Out of scope:**
- sqlc query code / generated Go (P2-T02)
- Migration runner / embedding (P2-T02)
- Seed data

## Acceptance criteria

- [ ] `sqlite3 /tmp/test.db < backend/db/schema.sql` executes with no errors
- [ ] All six tables exist with correct relationships (verifiable via `.schema`)
- [ ] Foreign keys enforce: inserting a child with a nonexistent parent fails when `PRAGMA foreign_keys = ON`
- [ ] `marketplace_link` enforces bidirectional uniqueness (no duplicate pair; consider a `CHECK` to keep `product_a_id < product_b_id`)
- [ ] Unit normalization columns present on `product` and `price_record`
- [ ] Every table and non-obvious column has a SQL comment explaining intent
- [ ] `backend/db/README.md` documents the entity relationships and base-unit convention
- [ ] Re-running `schema.sql` against a fresh DB is idempotent (use `CREATE TABLE IF NOT EXISTS`)

## Technical notes

- Depends on P1-T01 for repo location (`backend/db/`).
- Suggested column-level shape (implementer may refine, but keep these contracts for downstream tickets):
  - `correspondent(id PK, name UNIQUE NOT NULL, created_at)`
  - `receipt(id PK, correspondent_id FK→correspondent, purchased_at DATE NOT NULL, source TEXT NOT NULL CHECK(source IN ('paperless','upload')), external_doc_id TEXT, raw_file_ref TEXT, created_at)`
  - `product(id PK, display_name TEXT NOT NULL, base_unit TEXT NOT NULL, created_at)`
  - `raw_item(id PK, receipt_id FK→receipt, product_id FK→product NULL, raw_text TEXT NOT NULL, raw_quantity TEXT, quantity_value REAL, quantity_unit TEXT, created_at)` — `product_id` nullable until an item is matched/linked to a canonical product
  - `price_record(id PK, raw_item_id FK→raw_item UNIQUE, total_price REAL NOT NULL, unit_price REAL, currency TEXT NOT NULL DEFAULT 'USD', created_at)`
  - `marketplace_link(id PK, product_a_id FK→product, product_b_id FK→product, created_at, UNIQUE(product_a_id, product_b_id), CHECK(product_a_id < product_b_id))`
- The `unit_price` precomputation is written by the Phase 3 orchestrator; schema just stores it.
- This file is the seam — Phase 3 (ingestion) and Phase 4 (API) depend on these table/column names, so name them carefully and document any deviation in the README.

## Review feedback

