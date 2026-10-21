-- +goose Up
-- +goose StatementBegin

CREATE TABLE velez.whitelisted_images
(
    id         UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    repository TEXT        NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE velez.whitelisted_images;

-- +goose StatementEnd
