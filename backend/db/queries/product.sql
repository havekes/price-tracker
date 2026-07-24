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

-- name: ListLinkedProducts :many
SELECT p.* 
FROM product p
JOIN marketplace_link ml ON (ml.product_a_id = p.id OR ml.product_b_id = p.id)
WHERE (ml.product_a_id = $1 OR ml.product_b_id = $1) 
  AND p.id != $1
ORDER BY p.display_name;

-- name: ListPriceHistoryByProduct :many
SELECT 
    pr.id,
    pr.total_price,
    pr.unit_price,
    ri.raw_text,
    ri.raw_quantity,
    c.name AS correspondent_name,
    r.purchased_at
FROM price_record pr
JOIN raw_item ri ON pr.raw_item_id = ri.id
JOIN receipt r ON ri.receipt_id = r.id
JOIN correspondent c ON r.correspondent_id = c.id
WHERE ri.product_id = $1
ORDER BY r.purchased_at DESC;
