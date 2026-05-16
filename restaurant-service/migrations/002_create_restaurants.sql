-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS restaurants (
    id          BIGSERIAL PRIMARY KEY,
    name        TEXT        NOT NULL,
    description TEXT        NOT NULL DEFAULT '',
    category_id BIGINT      REFERENCES categories(id) ON DELETE SET NULL,
    address     TEXT        NOT NULL,
    phone       TEXT        NOT NULL DEFAULT '',
    image_url   TEXT        NOT NULL DEFAULT '',
    is_active   BOOLEAN     NOT NULL DEFAULT true,
    rating      NUMERIC(3,2) NOT NULL DEFAULT 0.00,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_restaurants_category  ON restaurants(category_id);
CREATE INDEX IF NOT EXISTS idx_restaurants_is_active ON restaurants(is_active);
CREATE INDEX IF NOT EXISTS idx_restaurants_rating    ON restaurants(rating DESC);

-- Full-text search index on name + description + address
CREATE INDEX IF NOT EXISTS idx_restaurants_search
    ON restaurants USING gin(to_tsvector('english', name || ' ' || description || ' ' || address));
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS restaurants;
-- +goose StatementEnd
