-- +goose Up
-- +goose StatementBegin
INSERT INTO restaurants (id, owner_id, name, description, category_id, address, phone, image_url) OVERRIDING SYSTEM VALUE VALUES
  (1, 1, 'Test Pizza', 'Вкусная пицца для тестирования', 1, 'ул. Тестовая, 1, Алматы', '+77001112233', '')
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('restaurants', 'id'), GREATEST((SELECT MAX(id) FROM restaurants), 1));
-- +goose StatementEnd

-- +goose Down
DELETE FROM restaurants WHERE id = 1;
