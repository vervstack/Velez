-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS velez.services
(
    name       TEXT PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS velez.services;
-- +goose StatementEnd
