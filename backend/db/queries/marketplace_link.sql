-- Marketplace_link CRUD.
-- A bidirectional pair link between two products across marketplaces.

-- name: CreateMarketplaceLink :one
INSERT INTO marketplace_link (product_a_id, product_b_id)
VALUES ($1, $2)
RETURNING *;

-- name: GetMarketplaceLink :one
SELECT * FROM marketplace_link
WHERE id = $1;

-- name: ListMarketplaceLinks :many
SELECT * FROM marketplace_link
ORDER BY id;

-- name: ListMarketplaceLinksByProduct :many
SELECT * FROM marketplace_link
WHERE product_a_id = $1 OR product_b_id = $1
ORDER BY id;

-- name: DeleteMarketplaceLink :exec
DELETE FROM marketplace_link
WHERE id = $1;
