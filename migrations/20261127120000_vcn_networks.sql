-- +goose Up
-- +goose StatementBegin

CREATE TABLE velez.networks
(
    id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    name         TEXT        NOT NULL UNIQUE,
    provider     TEXT        NOT NULL,
    source       TEXT        NOT NULL CHECK (source IN ('deployed', 'external')),
    url          TEXT        NOT NULL,
    token        TEXT        NOT NULL DEFAULT '',
    service_name TEXT        NOT NULL DEFAULT '',
    container_id TEXT        NOT NULL DEFAULT '',
    node_name    TEXT        NOT NULL DEFAULT '',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMENT ON COLUMN velez.networks.token IS 'Administrative provider credential used for cluster management.';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE velez.networks;

-- +goose StatementEnd
