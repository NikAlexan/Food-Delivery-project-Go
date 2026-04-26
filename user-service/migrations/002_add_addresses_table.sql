-- +goose Up
CREATE TABLE IF NOT EXISTS addresses (
    id         BIGSERIAL PRIMARY KEY,
    user_id    BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    street     TEXT NOT NULL DEFAULT '',
    city       TEXT NOT NULL DEFAULT '',
    zip        TEXT NOT NULL DEFAULT '',
    is_default BOOLEAN NOT NULL DEFAULT FALSE
);

-- +goose Down
DROP TABLE IF EXISTS addresses;