-- +goose Up
CREATE TABLE IF NOT EXISTS velez.service_resources
(
    service_name  TEXT        NOT NULL,
    resource_name TEXT        NOT NULL,
    resource_type TEXT        NOT NULL,
    status        TEXT        NOT NULL DEFAULT 'unknown',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (service_name, resource_name)
);

-- +goose Down
DROP TABLE IF EXISTS velez.service_resources;
