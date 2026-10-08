-- name: CreateTask :one
INSERT INTO velez.tasks (entity_id, action, context)
VALUES ($1, $2, $3)
ON CONFLICT (entity_id, action) WHERE status IN ('PENDING', 'RUNNING') DO NOTHING
RETURNING *;

-- name: GetTaskByEntityAction :one
SELECT *
FROM velez.tasks
WHERE entity_id = $1
  AND action = $2
ORDER BY id DESC
LIMIT 1;

-- name: GetTaskById :one
SELECT *
FROM velez.tasks
WHERE id = $1;

-- name: ClaimTask :one
UPDATE velez.tasks
SET status     = 'RUNNING',
    claimed_at = now(),
    claimed_by = $1,
    updated_at = now()
WHERE id = (
    SELECT t.id
    FROM velez.tasks t
    WHERE t.status = 'PENDING'
       OR (t.status = 'RUNNING' AND t.claimed_at < $2)
    ORDER BY t.created_at
    FOR UPDATE SKIP LOCKED
    LIMIT 1
    )
RETURNING *;

-- name: UpdateTaskContext :exec
UPDATE velez.tasks
SET context    = $1,
    updated_at = now()
WHERE id = $2;

-- name: RenewTaskClaim :exec
UPDATE velez.tasks
SET claimed_at = now()
WHERE id = $1
  AND status = 'RUNNING'
  AND claimed_by = $2;

-- name: FinishTask :exec
UPDATE velez.tasks
SET status     = $1,
    error      = $2,
    updated_at = now()
WHERE id = $3;

-- name: ListProvisioningTasks :many
SELECT id,
       entity_id,
       action,
       status,
       context,
       error,
       claimed_at,
       claimed_by,
       created_at,
       updated_at,
       environment_id
FROM velez.tasks
WHERE action = ANY (sqlc.arg(actions)::text[])
  AND (status IN ('PENDING', 'RUNNING')
    OR (status = 'FAILED' AND updated_at > sqlc.arg(failed_since)))
ORDER BY created_at;

-- name: DeleteTask :exec
DELETE
FROM velez.tasks
WHERE id = $1;
