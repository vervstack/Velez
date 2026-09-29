-- +goose Up
-- +goose StatementBegin

ALTER TABLE velez.runners
    ADD COLUMN concurrent INT NOT NULL DEFAULT 1;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE velez.runners
    DROP COLUMN concurrent;

-- +goose StatementEnd
