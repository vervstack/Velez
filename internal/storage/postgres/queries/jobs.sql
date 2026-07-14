-- name: GetJob :one
SELECT *
FROM velez.jobs
WHERE task_id = $1
  AND job_name = $2;

-- name: CreateRunningJob :one
INSERT INTO velez.jobs (task_id, job_name, status)
VALUES ($1, $2, 'RUNNING')
ON CONFLICT (task_id, job_name) DO NOTHING
RETURNING *;

-- name: FinishJob :exec
UPDATE velez.jobs
SET status     = $1,
    error      = $2,
    updated_at = now()
WHERE task_id = $3
  AND job_name = $4;

-- name: ListJobsByTask :many
SELECT *
FROM velez.jobs
WHERE task_id = $1
ORDER BY created_at;
