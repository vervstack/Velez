-- name: UpsertDindInstance :one
INSERT INTO velez.dind_instances (service_id, is_sysbox_enabled)
VALUES ($1, $2)
ON CONFLICT (service_id) DO UPDATE
    SET is_sysbox_enabled = EXCLUDED.is_sysbox_enabled,
        updated_at        = NOW()
RETURNING service_id, is_sysbox_enabled, created_at, updated_at;

-- name: GetDindInstanceByServiceID :one
SELECT service_id, is_sysbox_enabled, created_at, updated_at
FROM velez.dind_instances
WHERE service_id = $1;

-- name: ListDindInstances :many
SELECT service_id, is_sysbox_enabled, created_at, updated_at
FROM velez.dind_instances
ORDER BY service_id;

-- name: DeleteDindInstance :exec
DELETE FROM velez.dind_instances
WHERE service_id = $1;
