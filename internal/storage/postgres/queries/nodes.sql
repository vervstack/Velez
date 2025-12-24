-- name: InitNode :one
INSERT INTO velez.nodes (name)
VALUES (CURRENT_USER)
RETURNING id;

-- name: UpdateOnline :exec
UPDATE velez.nodes
SET last_online = now()
WHERE name = CURRENT_USER;