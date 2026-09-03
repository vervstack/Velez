-- +goose Up
-- +goose StatementBegin

ALTER TABLE velez.deployment_specifications
    ADD COLUMN IF NOT EXISTS verv_descriptor JSON;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE velez.deployment_specifications
    DROP COLUMN IF EXISTS verv_descriptor;

-- +goose StatementEnd
