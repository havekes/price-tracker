-- Correspondent CRUD.
-- A person or entity whose documents/receipts are tracked.

-- name: CreateCorrespondent :one
INSERT INTO correspondent (name)
VALUES ($1)
RETURNING *;

-- name: GetCorrespondent :one
SELECT * FROM correspondent
WHERE id = $1;

-- name: ListCorrespondents :many
SELECT * FROM correspondent
ORDER BY name;

-- name: UpdateCorrespondent :one
UPDATE correspondent
SET name = $2
WHERE id = $1
RETURNING *;

-- name: DeleteCorrespondent :exec
DELETE FROM correspondent
WHERE id = $1;

-- name: GetCorrespondentByName :one
SELECT * FROM correspondent
WHERE name = $1 LIMIT 1;
