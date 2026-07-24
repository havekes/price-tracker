-- Raw_item CRUD.
-- A line item as it appears verbatim on a receipt, before normalization.

-- name: CreateRawItem :one
INSERT INTO raw_item (
    receipt_id,
    product_id,
    raw_text,
    raw_quantity,
    quantity_value,
    quantity_unit
) VALUES (
    $1, $2, $3, $4, $5, $6
)
RETURNING *;

-- name: GetRawItem :one
SELECT * FROM raw_item
WHERE id = $1;

-- name: ListRawItemsByReceipt :many
SELECT * FROM raw_item
WHERE receipt_id = $1
ORDER BY id;

-- name: UpdateRawItem :one
UPDATE raw_item
SET
    product_id = $2,
    raw_text = $3,
    raw_quantity = $4,
    quantity_value = $5,
    quantity_unit = $6
WHERE id = $1
RETURNING *;

-- name: DeleteRawItem :exec
DELETE FROM raw_item
WHERE id = $1;
