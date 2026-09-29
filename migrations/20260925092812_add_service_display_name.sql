-- +goose Up
-- +goose StatementBegin

ALTER TABLE velez.services ADD COLUMN display_name TEXT NOT NULL DEFAULT '';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE velez.services DROP COLUMN display_name;

-- +goose StatementEnd
