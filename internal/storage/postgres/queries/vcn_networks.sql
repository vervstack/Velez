-- name: CreateVcnNetwork :exec
INSERT INTO velez.networks (
    id,
    name,
    provider,
    source,
    url,
    token,
    service_name,
    container_id,
    node_name,
    created_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10);

-- name: GetVcnNetworkById :one
SELECT id,
       name,
       provider,
       source,
       url,
       token,
       service_name,
       container_id,
       node_name,
       created_at
FROM velez.networks
WHERE id = $1;

-- name: ListVcnNetworks :many
SELECT id,
       name,
       provider,
       source,
       url,
       token,
       service_name,
       container_id,
       node_name,
       created_at
FROM velez.networks
ORDER BY created_at, id;

-- name: DropVcnNetworkById :execrows
DELETE FROM velez.networks
WHERE id = $1;

-- name: GetCurrentNode :one
SELECT CURRENT_USER::TEXT AS node_name;
