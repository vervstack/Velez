-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS velez.plugins
(
    plugin_type INT         NOT NULL PRIMARY KEY,
    state       INT         NOT NULL DEFAULT 0,
    port        INT,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS velez.plugins;

-- +goose StatementEnd
