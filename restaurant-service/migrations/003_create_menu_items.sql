-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS menu_items (
    id            BIGSERIAL PRIMARY KEY,
    restaurant_id BIGINT       NOT NULL REFERENCES restaurants(id) ON DELETE CASCADE,
    name          TEXT         NOT NULL,
    description   TEXT         NOT NULL DEFAULT '',
    price         NUMERIC(10,2) NOT NULL CHECK (price > 0),
    category      TEXT         NOT NULL DEFAULT '',
    image_url     TEXT         NOT NULL DEFAULT '',
    is_available  BOOLEAN      NOT NULL DEFAULT true,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_menu_items_restaurant ON menu_items(restaurant_id);
CREATE INDEX IF NOT EXISTS idx_menu_items_category   ON menu_items(category);
CREATE INDEX IF NOT EXISTS idx_menu_items_available  ON menu_items(is_available);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS menu_items;
-- +goose StatementEnd
