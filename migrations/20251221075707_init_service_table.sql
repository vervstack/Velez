-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS velez.services
(
    name TEXT PRIMARY KEY
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS velez.services;
-- +goose StatementEnd
