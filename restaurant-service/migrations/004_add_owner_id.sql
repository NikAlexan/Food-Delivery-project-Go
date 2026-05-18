-- +goose Up
ALTER TABLE restaurants ADD COLUMN IF NOT EXISTS owner_id BIGINT NOT NULL DEFAULT 0;

-- +goose Down
ALTER TABLE restaurants DROP COLUMN IF EXISTS owner_id;
