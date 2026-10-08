-- +goose Up
-- +goose StatementBegin

ALTER TABLE velez.runners
    ADD COLUMN gitlab_runner_id BIGINT NOT NULL DEFAULT 0;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE velez.runners
    DROP COLUMN gitlab_runner_id;

-- +goose StatementEnd
