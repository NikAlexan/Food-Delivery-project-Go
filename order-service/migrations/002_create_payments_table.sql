-- +goose Up
CREATE TABLE IF NOT EXISTS payments (
    id         BIGSERIAL PRIMARY KEY,
    order_id   BIGINT        NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    status     VARCHAR(20)   NOT NULL DEFAULT 'pending',
    amount     NUMERIC(10,2) NOT NULL,
    method     VARCHAR(50)   NOT NULL,
    created_at TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_payments_order_id ON payments(order_id);

-- +goose Down
DROP TABLE IF EXISTS payments;
