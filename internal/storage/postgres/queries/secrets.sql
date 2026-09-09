-- name: PutSecret :exec
INSERT INTO velez.secrets (scope, owner, key, value)
VALUES ($1, $2, $3, $4)
ON CONFLICT (scope, owner, key) DO UPDATE
    SET value      = EXCLUDED.value,
        updated_at = NOW();

-- name: GetSecretValue :one
SELECT value
FROM velez.secrets
WHERE scope = $1
  AND owner = $2
  AND key = $3;

-- name: DeleteSecret :exec
DELETE FROM velez.secrets
WHERE scope = $1
  AND owner = $2
  AND key = $3;

-- name: ListSecretRefs :many
SELECT scope, owner, key
FROM velez.secrets
WHERE scope = $1
  AND owner = $2
ORDER BY key;
