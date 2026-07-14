-- name: CreateTask :one
INSERT INTO velez.tasks (entity_id, action, context)
VALUES ($1, $2, $3)
ON CONFLICT (entity_id, action) DO NOTHING
RETURNING *;

-- name: GetTaskByEntityAction :one
SELECT *
FROM velez.tasks
WHERE entity_id = $1
  AND action = $2;

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

-- name: FinishTask :exec
UPDATE velez.tasks
SET status     = $1,
    error      = $2,
    updated_at = now()
WHERE id = $3;
