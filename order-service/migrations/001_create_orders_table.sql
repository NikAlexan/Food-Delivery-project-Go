-- +goose Up
CREATE TABLE IF NOT EXISTS orders (
    id            BIGSERIAL PRIMARY KEY,
    user_id       BIGINT        NOT NULL,
    restaurant_id BIGINT        NOT NULL,
    status        VARCHAR(20)   NOT NULL DEFAULT 'pending',
    total         NUMERIC(10,2) NOT NULL DEFAULT 0,
    created_at    TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS order_items (
    id           BIGSERIAL PRIMARY KEY,
    order_id     BIGINT        NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    menu_item_id BIGINT        NOT NULL,
    name         VARCHAR(255)  NOT NULL,
    quantity     INT           NOT NULL CHECK (quantity > 0),
    price        NUMERIC(10,2) NOT NULL
);

CREATE INDEX idx_orders_user_id ON orders(user_id);
CREATE INDEX idx_orders_status  ON orders(status);

-- +goose Down
DROP TABLE IF EXISTS order_items;
DROP TABLE IF EXISTS orders;
