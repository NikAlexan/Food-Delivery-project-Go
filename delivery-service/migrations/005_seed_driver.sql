-- +goose Up
-- +goose StatementBegin
INSERT INTO drivers (id, user_id, name, email, phone) OVERRIDING SYSTEM VALUE VALUES
  (1, 3, 'Test Driver', 'delivery@test.kz', '+77003334455')
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('drivers', 'id'), GREATEST((SELECT MAX(id) FROM drivers), 1));
-- +goose StatementEnd

-- +goose Down
DELETE FROM drivers WHERE id = 1;
