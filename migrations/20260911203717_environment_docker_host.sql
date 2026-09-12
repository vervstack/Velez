-- +goose Up
-- +goose StatementBegin

ALTER TABLE velez.environments
    ADD COLUMN docker_host TEXT NOT NULL DEFAULT '';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE velez.environments
    DROP COLUMN docker_host;

-- +goose StatementEnd
