-- name: UpsertService :exec
INSERT INTO velez.services (name)
VALUES ($1)
ON CONFLICT (name) DO NOTHING;

-- name: GetByName :one
SELECT id,
       name,
       created_at
FROM velez.services
WHERE name = $1
    FETCH FIRST 1 ROWS ONLY;

-- name: GetById :one
SELECT id,
       name,
       created_at
FROM velez.services
WHERE id = $1
    FETCH FIRST 1 ROWS ONLY;