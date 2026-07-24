-- Price_record CRUD.
-- A price observation tied to a raw item (1:1).

-- name: CreatePriceRecord :one
INSERT INTO price_record (
    raw_item_id,
    total_price,
    unit_price,
    currency
) VALUES (
    $1, $2, $3, $4
)
RETURNING *;

-- name: GetPriceRecord :one
SELECT * FROM price_record
WHERE id = $1;

-- name: GetPriceRecordByRawItem :one
SELECT * FROM price_record
WHERE raw_item_id = $1;

-- name: ListPriceRecords :many
SELECT * FROM price_record
ORDER BY id;

-- name: UpdatePriceRecord :one
UPDATE price_record
SET
    total_price = $2,
    unit_price = $3,
    currency = $4
WHERE id = $1
RETURNING *;

-- name: DeletePriceRecord :exec
DELETE FROM price_record
WHERE id = $1;
