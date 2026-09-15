-- name: UpsertRegistryInstance :one
INSERT INTO velez.registry_instances (service_id, port, ui_port, username, secret_ref)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (service_id) DO UPDATE
    SET port       = EXCLUDED.port,
        ui_port    = EXCLUDED.ui_port,
        username   = EXCLUDED.username,
        secret_ref = EXCLUDED.secret_ref,
        updated_at = NOW()
RETURNING service_id, port, ui_port, username, secret_ref, created_at, updated_at;

-- name: GetRegistryInstanceByServiceID :one
SELECT service_id, port, ui_port, username, secret_ref, created_at, updated_at
FROM velez.registry_instances
WHERE service_id = $1;

-- name: ListRegistryInstances :many
SELECT service_id, port, ui_port, username, secret_ref, created_at, updated_at
FROM velez.registry_instances
ORDER BY service_id;

-- name: DeleteRegistryInstance :exec
DELETE FROM velez.registry_instances
WHERE service_id = $1;
