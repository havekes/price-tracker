-- =============================================================================
-- Price Tracker — SQLite Schema
-- =============================================================================
-- Run against a fresh DB to create all tables, indexes, and constraints.
--   sqlite3 /tmp/price_tracker.db < backend/db/schema.sql
--
-- Uses CREATE TABLE IF NOT EXISTS so re-running is idempotent.
-- =============================================================================

PRAGMA foreign_keys = ON;

-- ---------------------------------------------------------------------------
-- correspondent — A merchant, store, or vendor from which receipts are
-- collected. Each correspondent must have a unique display name.
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS correspondent (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    name        TEXT    NOT NULL UNIQUE,
    created_at  TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

-- ---------------------------------------------------------------------------
-- receipt — A single shopping receipt or invoice uploaded by the user
-- or imported from Paperless-ngx. Belongs to exactly one correspondent.
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS receipt (
    id                INTEGER PRIMARY KEY AUTOINCREMENT,
    correspondent_id  INTEGER NOT NULL REFERENCES correspondent(id) ON DELETE RESTRICT,
    purchased_at      TEXT    NOT NULL,  -- ISO-8601 date (DATE semantics)
    source            TEXT    NOT NULL CHECK (source IN ('paperless', 'upload')),
    external_doc_id   TEXT,              -- ID in the external system (Paperless document PK)
    raw_file_ref      TEXT,              -- Path or object-storage key for the original file
    created_at        TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    FOREIGN KEY (correspondent_id) REFERENCES correspondent(id) ON DELETE RESTRICT
);

-- ---------------------------------------------------------------------------
-- product — A canonical product known to the system. Multiple raw_items
-- (from different receipts) can be linked to the same product. The base_unit
-- field records the unit the system normalises prices against for this
-- product (e.g. 'g', 'kg', 'ml', 'l', 'unit').
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS product (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    display_name  TEXT    NOT NULL,
    base_unit     TEXT    NOT NULL,  -- Normalisation unit: g, kg, ml, l, unit, …
    updated_at    TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    created_at    TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

-- ---------------------------------------------------------------------------
-- raw_item — A single line extracted from a receipt (one product row).
-- Contains the verbatim raw_text as it appeared on the receipt, plus
-- parsed numeric quantity fields. product_id is NULL until the item is
-- matched to a canonical product.
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS raw_item (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    receipt_id      INTEGER NOT NULL REFERENCES receipt(id) ON DELETE CASCADE,
    product_id      INTEGER REFERENCES product(id) ON DELETE SET NULL,
    raw_text        TEXT    NOT NULL,  -- Verbatim line text from receipt OCR/import
    raw_quantity    TEXT,              -- Verbatim quantity string, e.g. "2 lbs" or "500 g"
    quantity_value  REAL,              -- Parsed numeric quantity, e.g. 500.0
    quantity_unit   TEXT,              -- Parsed unit, e.g. "g" or "lbs"
    created_at      TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    FOREIGN KEY (receipt_id) REFERENCES receipt(id) ON DELETE CASCADE,
    FOREIGN KEY (product_id) REFERENCES product(id) ON DELETE SET NULL
);

-- ---------------------------------------------------------------------------
-- price_record — The price information for one raw_item. One-to-one with
-- raw_item (each item has at most one price record). The unit_price column
-- stores a precomputed price normalised to the product's base_unit; it is
-- populated by the Phase 3 ingestion pipeline.
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS price_record (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    raw_item_id   INTEGER NOT NULL UNIQUE REFERENCES raw_item(id) ON DELETE CASCADE,
    total_price   REAL    NOT NULL,  -- Total price as printed on receipt
    unit_price    REAL,              -- Precomputed: total_price normalised to the product's base_unit
    currency      TEXT    NOT NULL DEFAULT 'USD',
    updated_at    TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    created_at    TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    FOREIGN KEY (raw_item_id) REFERENCES raw_item(id) ON DELETE CASCADE
);

-- ---------------------------------------------------------------------------
-- marketplace_link — A directed-acyclic link between two products that are
-- known to be the same item sold on different marketplaces. The CHECK
-- constraint guarantees product_a_id < product_b_id so that each unordered
-- pair can only appear once; the UNIQUE index enforces that no duplicate
-- pair (in either order) is inserted.
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS marketplace_link (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    product_a_id    INTEGER NOT NULL REFERENCES product(id) ON DELETE CASCADE,
    product_b_id    INTEGER NOT NULL REFERENCES product(id) ON DELETE CASCADE,
    created_at      TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    UNIQUE (product_a_id, product_b_id),
    CHECK (product_a_id < product_b_id),
    FOREIGN KEY (product_a_id) REFERENCES product(id) ON DELETE CASCADE,
    FOREIGN KEY (product_b_id) REFERENCES product(id) ON DELETE CASCADE
);

-- =============================================================================
-- Indexes
-- =============================================================================

-- Speed up product lookups by display name (used for matching raw items).
CREATE INDEX IF NOT EXISTS idx_product_display_name ON product (display_name);

-- Speed up receipt queries filtered by correspondent or date range.
CREATE INDEX IF NOT EXISTS idx_receipt_correspondent_id ON receipt (correspondent_id);
CREATE INDEX IF NOT EXISTS idx_receipt_purchased_at     ON receipt (purchased_at);

-- Speed up finding all items on a receipt.
CREATE INDEX IF NOT EXISTS idx_raw_item_receipt_id ON raw_item (receipt_id);

-- Support lookups of price records by raw item (covering the UNIQUE FK).
CREATE INDEX IF NOT EXISTS idx_price_record_raw_item_id ON price_record (raw_item_id);

-- Support bidirectional marketplace-link lookups.
CREATE INDEX IF NOT EXISTS idx_marketplace_link_product_a_id ON marketplace_link (product_a_id);
CREATE INDEX IF NOT EXISTS idx_marketplace_link_product_b_id ON marketplace_link (product_b_id);
