-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS velez.runners
(
    service_id INT8        NOT NULL PRIMARY KEY REFERENCES velez.services (id) ON DELETE CASCADE,
    provider   TEXT        NOT NULL,
    scope      TEXT        NOT NULL,
    target     TEXT        NOT NULL,
    labels     TEXT[]      NOT NULL DEFAULT '{}',
    secret_ref TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS velez.runners;

-- +goose StatementEnd
