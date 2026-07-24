-- PostgreSQL schema for price-tracker
-- Auto-loaded by docker-compose on first init via /docker-entrypoint-initdb.d/schema.sql.
-- For schema changes during development, run:
--   docker compose down -v && docker compose up -d

-- Correspondent: a person or entity whose documents/receipts are tracked.
CREATE TABLE IF NOT EXISTS correspondent (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

COMMENT ON TABLE correspondent IS 'A person or entity whose documents/receipts are tracked';

-- Receipt: a single receipt document belonging to a correspondent.
CREATE TABLE IF NOT EXISTS receipt (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    correspondent_id BIGINT NOT NULL REFERENCES correspondent(id) ON DELETE RESTRICT,
    purchased_at DATE NOT NULL,
    source TEXT NOT NULL CHECK (source IN ('paperless', 'upload')),
    external_doc_id TEXT,
    raw_file_ref TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

COMMENT ON TABLE receipt IS 'A single receipt document belonging to a correspondent';
COMMENT ON COLUMN receipt.correspondent_id IS 'FK to correspondent — RESTRICT prevents deleting a correspondent with existing receipts';
COMMENT ON COLUMN receipt.purchased_at IS 'Date-only: the purchase date printed on the receipt';
COMMENT ON COLUMN receipt.source IS 'Ingestion source: paperless-ngx import or direct upload';
COMMENT ON COLUMN receipt.external_doc_id IS 'ID in the external source system (e.g., paperless-ngx document ID)';
COMMENT ON COLUMN receipt.raw_file_ref IS 'Path or reference to the original receipt file';

-- Product: a canonical product identified after normalizing raw item text.
CREATE TABLE IF NOT EXISTS product (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    display_name TEXT NOT NULL,
    base_unit TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

COMMENT ON TABLE product IS 'A canonical product after normalization';
COMMENT ON COLUMN product.base_unit IS 'Normalized unit for price comparison (e.g., g, kg, ml, l, unit)';

-- Raw_item: a line item as it appears verbatim on a receipt, before normalization.
CREATE TABLE IF NOT EXISTS raw_item (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    receipt_id BIGINT NOT NULL REFERENCES receipt(id) ON DELETE CASCADE,
    product_id BIGINT REFERENCES product(id) ON DELETE SET NULL,
    raw_text TEXT NOT NULL,
    raw_quantity TEXT,
    quantity_value DOUBLE PRECISION,
    quantity_unit TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

COMMENT ON TABLE raw_item IS 'A line item as it appears verbatim on a receipt';
COMMENT ON COLUMN raw_item.product_id IS 'FK to product — nullable until the item is linked to a canonical product';
COMMENT ON COLUMN raw_item.raw_quantity IS 'Quantity text as it appears on the receipt (e.g., "1 lb", "500g")';
COMMENT ON COLUMN raw_item.quantity_value IS 'Parsed numeric quantity extracted from raw_quantity';
COMMENT ON COLUMN raw_item.quantity_unit IS 'Normalized unit extracted from raw_quantity (e.g., g, lb, ml)';

-- Price_record: a price observation tied to a raw item (1:1).
CREATE TABLE IF NOT EXISTS price_record (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    raw_item_id BIGINT NOT NULL UNIQUE REFERENCES raw_item(id) ON DELETE CASCADE,
    total_price DOUBLE PRECISION NOT NULL,
    unit_price DOUBLE PRECISION,
    currency TEXT NOT NULL DEFAULT 'USD',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

COMMENT ON TABLE price_record IS 'A price observation — one per raw item (1:1 relationship)';
COMMENT ON COLUMN price_record.raw_item_id IS 'FK to raw_item — UNIQUE enforces one price record per item';
COMMENT ON COLUMN price_record.total_price IS 'Total price as printed on the receipt';
COMMENT ON COLUMN price_record.unit_price IS 'Precomputed price per base_unit (populated by the ingestion pipeline)';

-- Marketplace_link: a bidirectional pair link between two products across marketplaces.
CREATE TABLE IF NOT EXISTS marketplace_link (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    product_a_id BIGINT NOT NULL REFERENCES product(id) ON DELETE CASCADE,
    product_b_id BIGINT NOT NULL REFERENCES product(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (product_a_id, product_b_id),
    CONSTRAINT marketplace_link_product_a_id_product_b_id_check CHECK (product_a_id < product_b_id)
);

COMMENT ON TABLE marketplace_link IS 'Bidirectional link between two products across different marketplaces';
COMMENT ON COLUMN marketplace_link.product_a_id IS 'FK to product — lower ID side of the pair (enforced by CHECK)';
COMMENT ON COLUMN marketplace_link.product_b_id IS 'FK to product — higher ID side of the pair (enforced by CHECK)';
COMMENT ON CONSTRAINT marketplace_link_product_a_id_product_b_id_check ON marketplace_link IS 'Enforces product_a_id < product_b_id to prevent duplicates and self-links';

-- Indexes for hot lookup paths
CREATE INDEX IF NOT EXISTS idx_receipt_correspondent_id ON receipt(correspondent_id);
CREATE INDEX IF NOT EXISTS idx_receipt_purchased_at ON receipt(purchased_at);
CREATE INDEX IF NOT EXISTS idx_raw_item_receipt_id ON raw_item(receipt_id);
CREATE INDEX IF NOT EXISTS idx_raw_item_product_id ON raw_item(product_id);
CREATE INDEX IF NOT EXISTS idx_price_record_raw_item_id ON price_record(raw_item_id);
CREATE INDEX IF NOT EXISTS idx_product_display_name ON product(display_name);
CREATE INDEX IF NOT EXISTS idx_marketplace_link_product_a_id ON marketplace_link(product_a_id);
CREATE INDEX IF NOT EXISTS idx_marketplace_link_product_b_id ON marketplace_link(product_b_id);
