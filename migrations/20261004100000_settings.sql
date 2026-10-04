-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS velez.settings
(
    is_singleton                BOOLEAN     NOT NULL DEFAULT TRUE UNIQUE CHECK (is_singleton),
    is_sysbox_enabled           BOOLEAN     NOT NULL DEFAULT FALSE,
    is_sysbox_whitelist_ignored BOOLEAN     NOT NULL DEFAULT FALSE,
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO velez.settings (is_singleton)
VALUES (TRUE)
ON CONFLICT DO NOTHING;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS velez.settings;

-- +goose StatementEnd
