-- name: UpsertRunner :one
INSERT INTO velez.runners (service_id, provider, scope, target, labels, secret_ref)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (service_id) DO UPDATE
    SET provider   = EXCLUDED.provider,
        scope      = EXCLUDED.scope,
        target     = EXCLUDED.target,
        labels     = EXCLUDED.labels,
        secret_ref = EXCLUDED.secret_ref,
        updated_at = NOW()
RETURNING service_id, provider, scope, target, labels, secret_ref, created_at, updated_at;

-- name: GetRunnerByServiceID :one
SELECT service_id, provider, scope, target, labels, secret_ref, created_at, updated_at
FROM velez.runners
WHERE service_id = $1;

-- name: ListRunners :many
SELECT service_id, provider, scope, target, labels, secret_ref, created_at, updated_at
FROM velez.runners
ORDER BY service_id;

-- name: DeleteRunner :exec
DELETE FROM velez.runners
WHERE service_id = $1;
