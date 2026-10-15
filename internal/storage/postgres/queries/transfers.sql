-- name: InsertTransferToken :exec
INSERT INTO velez.transfer_tokens (token_hash, expires_at)
VALUES ($1, $2);

-- name: ConsumeTransferToken :one
UPDATE velez.transfer_tokens
SET used_at = NOW()
WHERE token_hash = $1
  AND used_at IS NULL
  AND expires_at > NOW()
RETURNING id;

-- name: InsertTransfer :one
INSERT INTO velez.transfers (is_source, container_name, peer_address, status)
VALUES ($1, $2, $3, $4)
RETURNING id, is_source, container_name, peer_address, status, error, total_bytes, done_bytes, created_at, updated_at;

-- name: UpdateTransferProgress :exec
UPDATE velez.transfers
SET total_bytes = $2,
    done_bytes  = $3,
    updated_at  = NOW()
WHERE id = $1;

-- name: FinishTransfer :exec
UPDATE velez.transfers
SET status     = $2,
    error      = $3,
    updated_at = NOW()
WHERE id = $1;

-- name: GetTransfer :one
SELECT id, is_source, container_name, peer_address, status, error, total_bytes, done_bytes, created_at, updated_at
FROM velez.transfers
WHERE id = $1;

-- name: ListTransfers :many
SELECT id, is_source, container_name, peer_address, status, error, total_bytes, done_bytes, created_at, updated_at
FROM velez.transfers
ORDER BY created_at DESC;

-- name: FailRunningTransfers :exec
UPDATE velez.transfers
SET status     = 'FAILED',
    error      = 'interrupted by restart',
    updated_at = NOW()
WHERE status IN ('PENDING', 'RUNNING');
