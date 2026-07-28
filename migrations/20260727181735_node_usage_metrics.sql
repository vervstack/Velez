-- +goose Up
ALTER TABLE velez.nodes ADD COLUMN cpu_percent DOUBLE PRECISION;
ALTER TABLE velez.nodes ADD COLUMN mem_percent DOUBLE PRECISION;

-- +goose Down
ALTER TABLE velez.nodes DROP COLUMN cpu_percent;
ALTER TABLE velez.nodes DROP COLUMN mem_percent;
