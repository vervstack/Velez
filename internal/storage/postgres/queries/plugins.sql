-- name: ListPlugins :many
SELECT plugin_type,
       state,
       port
FROM velez.plugins;

-- name: UpsertPlugin :exec
INSERT INTO velez.plugins (plugin_type, state, port, updated_at)
VALUES ($1, $2, $3, now())
ON CONFLICT (plugin_type) DO UPDATE
    SET state      = EXCLUDED.state,
        port       = EXCLUDED.port,
        updated_at = now();
