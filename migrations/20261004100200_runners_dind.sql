-- +goose Up
-- +goose StatementBegin

ALTER TABLE velez.runners
    ADD COLUMN dind_service_id INT8 REFERENCES velez.dind_instances (service_id) ON DELETE RESTRICT;

COMMENT ON COLUMN velez.runners.dind_service_id IS 'NULL = legacy runner on the host docker socket';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE velez.runners
    DROP COLUMN dind_service_id;

-- +goose StatementEnd
