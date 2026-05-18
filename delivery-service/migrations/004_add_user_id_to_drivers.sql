-- +goose Up
ALTER TABLE drivers ADD COLUMN IF NOT EXISTS user_id BIGINT UNIQUE;

-- +goose Down
ALTER TABLE drivers DROP COLUMN IF EXISTS user_id;
