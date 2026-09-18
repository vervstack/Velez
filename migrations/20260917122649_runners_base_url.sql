-- +goose Up
-- +goose StatementBegin

ALTER TABLE velez.runners
    ADD COLUMN base_url TEXT NOT NULL DEFAULT '';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE velez.runners
    DROP COLUMN base_url;

-- +goose StatementEnd
