-- +goose Up
-- +goose StatementBegin

UPDATE velez.services SET display_name = substr(name, 7) WHERE display_name IN ('', name) AND name LIKE 'pgaas\_%';
UPDATE velez.services SET display_name = substr(name, 4) WHERE display_name IN ('', name) AND name LIKE 'cr\_%';
UPDATE velez.services SET display_name = substr(name, 4) WHERE display_name IN ('', name) AND name LIKE 's3\_%';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

UPDATE velez.services SET display_name = name WHERE name LIKE 'pgaas\_%' AND display_name = substr(name, 7);
UPDATE velez.services SET display_name = name WHERE name LIKE 'cr\_%' AND display_name = substr(name, 4);
UPDATE velez.services SET display_name = name WHERE name LIKE 's3\_%' AND display_name = substr(name, 4);

-- +goose StatementEnd
