-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS velez.pg_instances
(
    service_id INT8        NOT NULL PRIMARY KEY REFERENCES velez.services (id) ON DELETE CASCADE,
    db_name    TEXT        NOT NULL,
    username   TEXT        NOT NULL,
    secret_ref TEXT        NOT NULL,
    port       INT4        NOT NULL DEFAULT 5432,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS velez.pg_instances;

-- +goose StatementEnd
