# Database Schema

The Price Tracker uses SQLite as its embedded database. The canonical schema is
defined in [`schema.sql`](./schema.sql), which is designed to be re-run
idempotently (`CREATE TABLE IF NOT EXISTS …`).

## Entity-Relationship Overview

```
correspondent 1──N receipt 1──N raw_item 1──1 price_record
                                    │
                                    N──1 product  N──N product
                                                  (via marketplace_link)
```

- **correspondent** — A merchant or store. Each receipt belongs to one
  correspondent. Deleting a correspondent is **restricted** when receipts
  reference it.
- **receipt** — A single uploaded or imported receipt. Deleting a receipt
  **cascades** to its raw items and their price records.
- **raw_item** — A single line extracted from a receipt. Contains the verbatim
  text and parsed quantity fields. May be linked to a canonical product;
  deleting a product **sets `product_id` to NULL** on linked items.
- **product** — A canonical product the system knows about. Multiple raw items
  from different receipts can be linked to the same product.
- **price_record** — The price info for one raw item (one-to-one). Stores both
  the raw `total_price` and `unit_price` (precomputed normalised price).
- **marketplace_link** — A symmetrical link between two products that represent
  the same item on different marketplaces. The `CHECK` constraint
  `product_a_id < product_b_id` prevents duplicate pairs in either direction.

## Unit-Normalisation Convention

To enable meaningful price comparison across products sold in different
quantities, the schema includes a normalisation convention at two levels:

### 1. Product-level: `base_unit`

Every `product` record defines a **`base_unit`** — the unit the system uses
when normalising prices for that product. Examples:

| `base_unit` | Meaning         |
|-------------|-----------------|
| `unit`      | Per item        |
| `g`         | Per gram        |
| `kg`        | Per kilogram    |
| `ml`        | Per millilitre  |
| `l`         | Per litre       |
| `oz`        | Per ounce       |
| `lb`        | Per pound       |

### 2. Price-level: `unit_price`

`price_record.unit_price` is a precomputed value representing the price
normalised to the linked product's `base_unit`. It is **not** computed by
the schema — the Phase 3 ingestion pipeline populates it by dividing
`total_price` by `quantity_value` (with appropriate unit conversion).

For example, given:
- Product: `base_unit = 'g'`
- Raw item: `raw_quantity = '2 kg'`, `quantity_value = 2000`, `quantity_unit = 'g'`
- Price record: `total_price = 10.00`

Then `unit_price = 10.00 / 2000 = 0.005` (price per gram).

### 3. Item-level: parsed quantity

`raw_item` stores both:
- `raw_quantity` — the verbatim string from the receipt (e.g. `"2 lbs"`)
- `quantity_value` (REAL) and `quantity_unit` (TEXT) — the parsed numeric
  value and unit, ready for the `unit_price` computation.

## Foreign-Key Semantics

| FK | ON DELETE | Rationale |
|----|-----------|-----------|
| `receipt.correspondent_id` → `correspondent.id` | `RESTRICT` | Prevent accidental deletion of a correspondent that has receipts |
| `raw_item.receipt_id` → `receipt.id` | `CASCADE` | Deleting a receipt should remove its items |
| `raw_item.product_id` → `product.id` | `SET NULL` | Removing a product leaves items unlinked rather than destroying data |
| `price_record.raw_item_id` → `raw_item.id` | `CASCADE` | Price record is an attribute of the item |
| `marketplace_link.product_a_id` / `product_b_id` → `product.id` | `CASCADE` | Removing a product cleans up its marketplace links |

## Timestamp Format

All timestamps use ISO-8601 with millisecond precision:
`strftime('%Y-%m-%dT%H:%M:%fZ', 'now')`.

Tables that support mutable data (`product`, `price_record`) also carry an
`updated_at` column, updated on creation (the ingestion pipeline may update it
explicitly on mutation).

## Running the Schema

```bash
sqlite3 /tmp/price_tracker.db < backend/db/schema.sql
```

Re-running is safe — all `CREATE` statements use `IF NOT EXISTS` and indexes
use `CREATE INDEX IF NOT EXISTS`.
