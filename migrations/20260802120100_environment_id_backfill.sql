-- +goose Up
-- +goose StatementBegin

ALTER TABLE velez.services ADD COLUMN environment_id INT8 REFERENCES velez.environments (id);
ALTER TABLE velez.deployments ADD COLUMN environment_id INT8 REFERENCES velez.environments (id);
ALTER TABLE velez.tasks ADD COLUMN environment_id INT8 REFERENCES velez.environments (id);
ALTER TABLE velez.jobs ADD COLUMN environment_id INT8 REFERENCES velez.environments (id);
ALTER TABLE velez.nodes ADD COLUMN environment_id INT8 REFERENCES velez.environments (id);
ALTER TABLE velez.service_dependencies ADD COLUMN environment_id INT8 REFERENCES velez.environments (id);
ALTER TABLE velez.service_resources ADD COLUMN environment_id INT8 REFERENCES velez.environments (id);

UPDATE velez.services SET environment_id = (SELECT id FROM velez.environments WHERE name = 'PROD') WHERE environment_id IS NULL;
UPDATE velez.deployments SET environment_id = (SELECT id FROM velez.environments WHERE name = 'PROD') WHERE environment_id IS NULL;
UPDATE velez.tasks SET environment_id = (SELECT id FROM velez.environments WHERE name = 'PROD') WHERE environment_id IS NULL;
UPDATE velez.jobs SET environment_id = (SELECT id FROM velez.environments WHERE name = 'PROD') WHERE environment_id IS NULL;
UPDATE velez.nodes SET environment_id = (SELECT id FROM velez.environments WHERE name = 'PROD') WHERE environment_id IS NULL;
UPDATE velez.service_dependencies SET environment_id = (SELECT id FROM velez.environments WHERE name = 'PROD') WHERE environment_id IS NULL;
UPDATE velez.service_resources SET environment_id = (SELECT id FROM velez.environments WHERE name = 'PROD') WHERE environment_id IS NULL;

ALTER TABLE velez.services ALTER COLUMN environment_id SET NOT NULL;
ALTER TABLE velez.deployments ALTER COLUMN environment_id SET NOT NULL;
ALTER TABLE velez.tasks ALTER COLUMN environment_id SET NOT NULL;
ALTER TABLE velez.jobs ALTER COLUMN environment_id SET NOT NULL;
ALTER TABLE velez.nodes ALTER COLUMN environment_id SET NOT NULL;
ALTER TABLE velez.service_dependencies ALTER COLUMN environment_id SET NOT NULL;
ALTER TABLE velez.service_resources ALTER COLUMN environment_id SET NOT NULL;

-- Every INSERT that predates environments (all of them today - none of the
-- sqlc-generated inserts name environment_id) must keep working, so the column
-- defaults to the default environment: "no environment given" means PROD, the
-- same rule ResolveEnvironmentSuffix applies on the wire. Without this, NOT
-- NULL turns every pre-existing insert into a constraint violation.
DO
$$
    DECLARE
        default_environment_id INT8;
        target_table           TEXT;
    BEGIN
        SELECT id INTO STRICT default_environment_id FROM velez.environments WHERE name = 'PROD';

        FOREACH target_table IN ARRAY ARRAY ['services', 'deployments', 'tasks', 'jobs',
            'nodes', 'service_dependencies', 'service_resources']
            LOOP
                EXECUTE format('ALTER TABLE velez.%I ALTER COLUMN environment_id SET DEFAULT %L',
                               target_table, default_environment_id);
            END LOOP;
    END
$$;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE velez.service_resources DROP COLUMN environment_id;
ALTER TABLE velez.service_dependencies DROP COLUMN environment_id;
ALTER TABLE velez.nodes DROP COLUMN environment_id;
ALTER TABLE velez.jobs DROP COLUMN environment_id;
ALTER TABLE velez.tasks DROP COLUMN environment_id;
ALTER TABLE velez.deployments DROP COLUMN environment_id;
ALTER TABLE velez.services DROP COLUMN environment_id;

-- +goose StatementEnd
