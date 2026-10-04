-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS velez.dind_instances
(
    service_id        INT8        NOT NULL PRIMARY KEY REFERENCES velez.services (id) ON DELETE CASCADE,
    is_sysbox_enabled BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS velez.dind_instances;

-- +goose StatementEnd
