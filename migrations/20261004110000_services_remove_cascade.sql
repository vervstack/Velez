-- +goose Up
-- +goose StatementBegin

ALTER TABLE velez.deployment_specifications
    DROP CONSTRAINT deployment_specifications_service_id_fkey,
    ADD CONSTRAINT deployment_specifications_service_id_fkey
        FOREIGN KEY (service_id) REFERENCES velez.services (id) ON DELETE CASCADE;

ALTER TABLE velez.deployments
    DROP CONSTRAINT deployments_spec_id_fkey,
    ADD CONSTRAINT deployments_spec_id_fkey
        FOREIGN KEY (spec_id) REFERENCES velez.deployment_specifications (id) ON DELETE CASCADE;

ALTER TABLE velez.shared_volumes
    DROP CONSTRAINT shared_volumes_service_id_fkey,
    DROP CONSTRAINT shared_volumes_specification_id_fkey,
    DROP CONSTRAINT shared_volumes_deployment_id_fkey,
    ADD CONSTRAINT shared_volumes_service_id_fkey
        FOREIGN KEY (service_id) REFERENCES velez.services (id) ON DELETE CASCADE,
    ADD CONSTRAINT shared_volumes_specification_id_fkey
        FOREIGN KEY (specification_id) REFERENCES velez.deployment_specifications (id) ON DELETE CASCADE,
    ADD CONSTRAINT shared_volumes_deployment_id_fkey
        FOREIGN KEY (deployment_id) REFERENCES velez.deployments (id) ON DELETE CASCADE;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE velez.shared_volumes
    DROP CONSTRAINT shared_volumes_service_id_fkey,
    DROP CONSTRAINT shared_volumes_specification_id_fkey,
    DROP CONSTRAINT shared_volumes_deployment_id_fkey,
    ADD CONSTRAINT shared_volumes_service_id_fkey
        FOREIGN KEY (service_id) REFERENCES velez.services (id),
    ADD CONSTRAINT shared_volumes_specification_id_fkey
        FOREIGN KEY (specification_id) REFERENCES velez.deployment_specifications (id),
    ADD CONSTRAINT shared_volumes_deployment_id_fkey
        FOREIGN KEY (deployment_id) REFERENCES velez.deployments (id);

ALTER TABLE velez.deployments
    DROP CONSTRAINT deployments_spec_id_fkey,
    ADD CONSTRAINT deployments_spec_id_fkey
        FOREIGN KEY (spec_id) REFERENCES velez.deployment_specifications (id);

ALTER TABLE velez.deployment_specifications
    DROP CONSTRAINT deployment_specifications_service_id_fkey,
    ADD CONSTRAINT deployment_specifications_service_id_fkey
        FOREIGN KEY (service_id) REFERENCES velez.services (id);

-- +goose StatementEnd
