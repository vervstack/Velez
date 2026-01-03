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