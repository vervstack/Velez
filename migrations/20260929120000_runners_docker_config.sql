-- +goose Up
-- +goose StatementBegin

ALTER TABLE velez.runners
    ADD COLUMN docker_image TEXT NOT NULL DEFAULT '',
    ADD COLUMN docker_socket_address TEXT NOT NULL DEFAULT '';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE velez.runners
    DROP COLUMN docker_image,
    DROP COLUMN docker_socket_address;

-- +goose StatementEnd
