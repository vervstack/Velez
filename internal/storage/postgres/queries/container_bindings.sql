-- name: UpsertContainerBinding :exec
INSERT INTO velez.container_bindings (service_id, node_id, environment, container_name)
VALUES ($1, $2, $3, $4)
ON CONFLICT (node_id, environment, container_name)
    DO UPDATE SET service_id = EXCLUDED.service_id;

-- name: ListContainerBindingsByNode :many
SELECT b.id,
       b.service_id,
       s.name AS service_name,
       b.node_id,
       b.environment,
       b.container_name
FROM velez.container_bindings b
         JOIN velez.services s ON s.id = b.service_id
WHERE b.node_id = $1
  AND b.environment = $2;
