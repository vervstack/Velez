-- +goose Up
-- +goose StatementBegin

ALTER TABLE velez.runners
    ADD COLUMN is_buildkit_enabled BOOLEAN NOT NULL DEFAULT FALSE;

COMMENT ON COLUMN velez.runners.is_buildkit_enabled IS 'TRUE = a BuildKit daemon runs inside the runner DinD';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE velez.runners
    DROP COLUMN is_buildkit_enabled;

-- +goose StatementEnd
