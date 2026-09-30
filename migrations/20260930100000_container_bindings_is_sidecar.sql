-- +goose Up
-- +goose StatementBegin

ALTER TABLE velez.container_bindings
    ADD COLUMN is_sidecar BOOLEAN NOT NULL DEFAULT FALSE;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE velez.container_bindings
    DROP COLUMN is_sidecar;

-- +goose StatementEnd
