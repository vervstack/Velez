-- name: UpsertServiceDependency :exec
INSERT INTO velez.service_dependencies (source_service, target_service, proto)
VALUES ($1, $2, $3)
ON CONFLICT (source_service, target_service)
    DO UPDATE SET proto      = EXCLUDED.proto,
                  updated_at = NOW();

-- name: GetServiceDependencies :many
SELECT target_service, proto
FROM velez.service_dependencies
WHERE source_service = $1;

-- name: GetServiceCallers :many
SELECT source_service, proto
FROM velez.service_dependencies
WHERE target_service = $1;
