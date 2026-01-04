-- name: CreateSpecification :one
INSERT INTO velez.deployment_specifications (name, verv_payload)
VALUES ($1, $2)
RETURNING id;

-- name: CreateDeployment :one
INSERT
INTO velez.deployments
    (service_id, node_id, status, spec_id)
VALUES ($1, $2, $3, $4)
RETURNING (id, service_id, node_id, created_at, updated_at, status, spec_id);

-- name: GetSpecificationById :one
SELECT id,
       name,
       created_at,
       verv_payload
FROM velez.deployment_specifications spec
WHERE spec.id = $1;

-- name: UpdateDeploymentStatus :exec
UPDATE velez.deployments
SET status = $1
WHERE id = $2;