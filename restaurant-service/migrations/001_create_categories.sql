-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS categories (
    id         BIGSERIAL PRIMARY KEY,
    name       TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Seed common food categories
INSERT INTO categories (name) VALUES
    ('Pizza'),
    ('Burgers'),
    ('Sushi'),
    ('Chinese'),
    ('Mexican'),
    ('Italian'),
    ('Indian'),
    ('Thai'),
    ('Salads'),
    ('Desserts')
ON CONFLICT (name) DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS categories;
-- +goose StatementEnd
