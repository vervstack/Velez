-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS velez.registry_instances
(
    service_id INT8        NOT NULL PRIMARY KEY REFERENCES velez.services (id) ON DELETE CASCADE,
    port       INT4        NOT NULL DEFAULT 5000,
    ui_port    INT4        NOT NULL,
    username   TEXT        NOT NULL,
    secret_ref TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS velez.registry_instances;

-- +goose StatementEnd
