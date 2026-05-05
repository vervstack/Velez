-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS velez.plugins
(
    plugin_type TEXT   NOT NULL PRIMARY KEY,
    service_id  BIGINT NULL REFERENCES velez.services (id)
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS velez.plugins;

-- +goose StatementEnd
