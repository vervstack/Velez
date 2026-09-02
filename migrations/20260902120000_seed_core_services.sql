-- +goose Up
-- +goose StatementBegin
INSERT INTO velez.services (name)
VALUES ('velez'), ('postgres')
ON CONFLICT (name) DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM velez.services WHERE name IN ('velez', 'postgres');
-- +goose StatementEnd
