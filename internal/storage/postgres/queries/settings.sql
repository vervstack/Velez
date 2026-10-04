-- name: GetSettings :one
SELECT is_sysbox_enabled, is_sysbox_whitelist_ignored, updated_at
FROM velez.settings
WHERE is_singleton;

-- name: UpsertSettings :one
INSERT INTO velez.settings (is_singleton, is_sysbox_enabled, is_sysbox_whitelist_ignored)
VALUES (TRUE, $1, $2)
ON CONFLICT (is_singleton) DO UPDATE
    SET is_sysbox_enabled           = EXCLUDED.is_sysbox_enabled,
        is_sysbox_whitelist_ignored = EXCLUDED.is_sysbox_whitelist_ignored,
        updated_at                  = NOW()
RETURNING is_sysbox_enabled, is_sysbox_whitelist_ignored, updated_at;
