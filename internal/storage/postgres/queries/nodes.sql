-- name: InitNode :exec
INSERT INTO velez.nodes (name, region) VALUES (CURRENT_USER, $1)
ON CONFLICT (name) DO NOTHING;

-- name: GetOwnRegion :one
SELECT region FROM velez.nodes WHERE name = CURRENT_USER;

-- name: UpdateOnline :exec
UPDATE velez.nodes
SET last_online = now(), cpu_percent = $1, mem_percent = $2
WHERE name = CURRENT_USER;