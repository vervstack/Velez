-- name: CreateSpecification :one
INSERT INTO velez.deployment_specifications (name, service_id, verv_payload)
VALUES ($1, $2, $3)
RETURNING id;

-- name: CreateDeployment :one
INSERT
INTO velez.deployments
    (node_id, status, spec_id)
VALUES ($1, $2, $3)
RETURNING (id, node_id, created_at, updated_at, status, spec_id);

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