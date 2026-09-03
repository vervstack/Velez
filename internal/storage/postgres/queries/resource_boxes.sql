-- name: GetResourceBoxByName :one
SELECT name, cpu, ram_mb, disk_mb, is_builtin, created_at, updated_at
FROM velez.resource_boxes
WHERE name = $1;

-- name: ListResourceBoxes :many
SELECT name, cpu, ram_mb, disk_mb, is_builtin, created_at, updated_at
FROM velez.resource_boxes
ORDER BY name;

-- name: UpsertResourceBox :one
INSERT INTO velez.resource_boxes (name, cpu, ram_mb, disk_mb, is_builtin)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (name) DO UPDATE
    SET cpu        = $2,
        ram_mb     = $3,
        disk_mb    = $4,
        is_builtin = $5,
        updated_at = NOW()
RETURNING name, cpu, ram_mb, disk_mb, is_builtin, created_at, updated_at;

-- name: DeleteResourceBoxByName :exec
DELETE FROM velez.resource_boxes
WHERE name = $1
  AND is_builtin = FALSE;
