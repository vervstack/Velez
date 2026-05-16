-- name: UpsertServiceResource :exec
INSERT INTO velez.service_resources (service_name, resource_name, resource_type)
VALUES ($1, $2, $3)
ON CONFLICT (service_name, resource_name)
    DO UPDATE SET resource_type = EXCLUDED.resource_type,
                  updated_at    = NOW();

-- name: GetServiceResources :many
SELECT resource_name, resource_type, status
FROM velez.service_resources
WHERE service_name = $1;
