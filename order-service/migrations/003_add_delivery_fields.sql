-- +goose Up
ALTER TABLE orders ADD COLUMN delivery_address TEXT NOT NULL DEFAULT '';
ALTER TABLE orders ADD COLUMN user_email       TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE orders DROP COLUMN delivery_address;
ALTER TABLE orders DROP COLUMN user_email;
