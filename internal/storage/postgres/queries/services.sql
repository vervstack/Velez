-- name: UpsertService :exec
INSERT INTO velez.services (name, display_name)
VALUES ($1, $2)
ON CONFLICT (name) DO UPDATE
    SET display_name = EXCLUDED.display_name;

-- name: GetByName :one
SELECT *
FROM velez.services
WHERE name = $1
    FETCH FIRST 1 ROWS ONLY;

-- name: DeleteByName :exec
DELETE FROM velez.services
WHERE name = $1;

