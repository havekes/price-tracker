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

-- name: GetProductByName :one
SELECT * FROM product
WHERE display_name = $1 LIMIT 1;

-- name: ListProductsWithPrices :many
SELECT p.id, p.display_name, p.base_unit, AVG(pr.unit_price)::DOUBLE PRECISION AS latest_avg_price
FROM product p
LEFT JOIN raw_item ri ON p.id = ri.product_id
LEFT JOIN price_record pr ON ri.id = pr.raw_item_id
GROUP BY p.id, p.display_name, p.base_unit
ORDER BY p.display_name;
