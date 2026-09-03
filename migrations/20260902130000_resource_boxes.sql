-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS velez.resource_boxes
(
    name       TEXT PRIMARY KEY,
    cpu        NUMERIC     NOT NULL,
    ram_mb     BIGINT      NOT NULL,
    disk_mb    BIGINT      NOT NULL,
    is_builtin BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO velez.resource_boxes (name, cpu, ram_mb, disk_mb, is_builtin)
VALUES ('small', 0.5, 512, 2048, TRUE),
       ('medium', 1.0, 1024, 8192, TRUE),
       ('large', 2.0, 4096, 32768, TRUE)
ON CONFLICT (name) DO NOTHING;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS velez.resource_boxes;

-- +goose StatementEnd
