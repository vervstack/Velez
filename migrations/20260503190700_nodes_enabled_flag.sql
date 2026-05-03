-- +goose Up
-- +goose StatementBegin
ALTER TABLE velez.nodes
    ADD COLUMN is_enabled BOOLEAN NOT NULL DEFAULT true,
    ADD COLUMN addr       TEXT    NOT NULL DEFAULT '';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- +goose StatementEnd
