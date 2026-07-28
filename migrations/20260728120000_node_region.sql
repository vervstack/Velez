-- +goose Up
ALTER TABLE velez.nodes ADD COLUMN region TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE velez.nodes DROP COLUMN region;
