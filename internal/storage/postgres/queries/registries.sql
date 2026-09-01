-- name: CreateRegistry :one
INSERT INTO velez.registries (name, type, url, username, secret, is_default)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetRegistryByID :one
SELECT *
FROM velez.registries
WHERE id = $1;

-- name: ListRegistries :many
SELECT *
FROM velez.registries
ORDER BY id;

-- name: UpdateRegistry :one
UPDATE velez.registries
SET name       = $2,
    type       = $3,
    url        = $4,
    username   = $5,
    secret     = $6,
    is_default = $7,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: DeleteRegistry :exec
DELETE FROM velez.registries
WHERE id = $1;

-- name: ClearDefaultRegistry :exec
UPDATE velez.registries
SET is_default = FALSE
WHERE is_default = TRUE;
