-- +goose Up
CREATE TABLE IF NOT EXISTS velez.service_dependencies
(
    source_service TEXT        NOT NULL,
    target_service TEXT        NOT NULL,
    proto          TEXT        NOT NULL DEFAULT '',
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (source_service, target_service)
);

-- +goose Down
DROP TABLE IF EXISTS velez.service_dependencies;
