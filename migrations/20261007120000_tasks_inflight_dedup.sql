-- +goose Up
-- +goose StatementBegin

ALTER TABLE velez.tasks
    DROP CONSTRAINT tasks_entity_id_action_key;

CREATE UNIQUE INDEX tasks_entity_id_action_inflight_key
    ON velez.tasks (entity_id, action)
    WHERE status IN ('PENDING', 'RUNNING');

CREATE INDEX tasks_entity_id_action_id_idx
    ON velez.tasks (entity_id, action, id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS velez.tasks_entity_id_action_id_idx;
DROP INDEX IF EXISTS velez.tasks_entity_id_action_inflight_key;

-- The unique constraint allows one row per (entity_id, action): keep only the newest task of each pair.
DELETE
FROM velez.jobs
WHERE task_id IN (SELECT t.id
                  FROM velez.tasks t
                  WHERE EXISTS (SELECT 1
                                FROM velez.tasks newer
                                WHERE newer.entity_id = t.entity_id
                                  AND newer.action = t.action
                                  AND newer.id > t.id));

DELETE
FROM velez.tasks t
WHERE EXISTS (SELECT 1
              FROM velez.tasks newer
              WHERE newer.entity_id = t.entity_id
                AND newer.action = t.action
                AND newer.id > t.id);

ALTER TABLE velez.tasks
    ADD CONSTRAINT tasks_entity_id_action_key UNIQUE (entity_id, action);

-- +goose StatementEnd
