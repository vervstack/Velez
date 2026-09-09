-- name: UpsertPgInstance :one
INSERT INTO velez.pg_instances (service_id, db_name, username, secret_ref, port)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (service_id) DO UPDATE
    SET db_name    = EXCLUDED.db_name,
        username   = EXCLUDED.username,
        secret_ref = EXCLUDED.secret_ref,
        port       = EXCLUDED.port,
        updated_at = NOW()
RETURNING service_id, db_name, username, secret_ref, port, created_at, updated_at;

-- name: GetPgInstanceByServiceID :one
SELECT service_id, db_name, username, secret_ref, port, created_at, updated_at
FROM velez.pg_instances
WHERE service_id = $1;

-- name: ListPgInstances :many
SELECT service_id, db_name, username, secret_ref, port, created_at, updated_at
FROM velez.pg_instances
ORDER BY service_id;

-- name: DeletePgInstance :exec
DELETE FROM velez.pg_instances
WHERE service_id = $1;
