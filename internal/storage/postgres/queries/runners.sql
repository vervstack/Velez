-- name: UpsertRunner :one
INSERT INTO velez.runners (service_id, provider, scope, target, labels, secret_ref, base_url,
                           docker_image, docker_socket_address)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
ON CONFLICT (service_id) DO UPDATE
    SET provider              = EXCLUDED.provider,
        scope                 = EXCLUDED.scope,
        target                = EXCLUDED.target,
        labels                = EXCLUDED.labels,
        secret_ref            = EXCLUDED.secret_ref,
        base_url              = EXCLUDED.base_url,
        docker_image          = EXCLUDED.docker_image,
        docker_socket_address = EXCLUDED.docker_socket_address,
        updated_at            = NOW()
RETURNING service_id, provider, scope, target, labels, secret_ref, created_at, updated_at, base_url,
    docker_image, docker_socket_address;

-- name: GetRunnerByServiceID :one
SELECT service_id, provider, scope, target, labels, secret_ref, created_at, updated_at, base_url,
       docker_image, docker_socket_address
FROM velez.runners
WHERE service_id = $1;

-- name: ListRunners :many
SELECT service_id, provider, scope, target, labels, secret_ref, created_at, updated_at, base_url,
       docker_image, docker_socket_address
FROM velez.runners
ORDER BY service_id;

-- name: DeleteRunner :exec
DELETE FROM velez.runners
WHERE service_id = $1;
