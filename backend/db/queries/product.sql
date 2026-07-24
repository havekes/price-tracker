-- Product CRUD.
-- A canonical product identified after normalizing raw item text.

-- name: CreateProduct :one
INSERT INTO product (display_name, base_unit)
VALUES ($1, $2)
RETURNING *;

-- name: GetProduct :one
SELECT * FROM product
WHERE id = $1;

-- name: ListProducts :many
SELECT * FROM product
ORDER BY display_name;

-- name: UpdateProduct :one
UPDATE product
SET
    display_name = $2,
    base_unit = $3
WHERE id = $1
RETURNING *;

-- name: DeleteProduct :exec
DELETE FROM product
WHERE id = $1;
