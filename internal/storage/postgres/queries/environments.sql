-- name: CreateEnvironment :one
INSERT INTO velez.environments (name, suffix)
VALUES ($1, $2)
RETURNING *;

-- name: GetEnvironmentByID :one
SELECT *
FROM velez.environments
WHERE id = $1;

-- name: GetEnvironmentByName :one
SELECT *
FROM velez.environments
WHERE name = $1;

-- name: ListEnvironments :many
SELECT *
FROM velez.environments
ORDER BY id;

-- name: UpdateEnvironment :one
UPDATE velez.environments
SET name       = $2,
    suffix     = $3,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: DeleteEnvironment :exec
DELETE FROM velez.environments
WHERE id = $1;
