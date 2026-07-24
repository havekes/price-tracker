-- Receipt CRUD.
-- A single receipt document belonging to a correspondent.

-- name: CreateReceipt :one
INSERT INTO receipt (
    correspondent_id,
    purchased_at,
    source,
    external_doc_id,
    raw_file_ref
) VALUES (
    $1, $2, $3, $4, $5
)
RETURNING *;

-- name: GetReceipt :one
SELECT * FROM receipt
WHERE id = $1;

-- name: ListReceipts :many
SELECT * FROM receipt
ORDER BY purchased_at DESC;

-- name: ListReceiptsByCorrespondent :many
SELECT * FROM receipt
WHERE correspondent_id = $1
ORDER BY purchased_at DESC;

-- name: UpdateReceipt :one
UPDATE receipt
SET
    correspondent_id = $2,
    purchased_at = $3,
    source = $4,
    external_doc_id = $5,
    raw_file_ref = $6
WHERE id = $1
RETURNING *;

-- name: DeleteReceipt :exec
DELETE FROM receipt
WHERE id = $1;
