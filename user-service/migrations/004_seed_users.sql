-- +goose Up
-- +goose StatementBegin
INSERT INTO users (id, email, password_hash, name, phone) OVERRIDING SYSTEM VALUE VALUES
  (1, 'restaurant@test.kz', '$2a$10$gJ9N0oJqp3XYt3tBYOmtXuAy0z06pXSn2LaeJc.D34sNUPsfoiizm', 'Restaurant Owner', '+77001112233'),
  (2, 'user@test.kz',       '$2a$10$gJ9N0oJqp3XYt3tBYOmtXuAy0z06pXSn2LaeJc.D34sNUPsfoiizm', 'Test Customer',    '+77002223344'),
  (3, 'delivery@test.kz',   '$2a$10$gJ9N0oJqp3XYt3tBYOmtXuAy0z06pXSn2LaeJc.D34sNUPsfoiizm', 'Test Driver',      '+77003334455')
ON CONFLICT (email) DO NOTHING;

SELECT setval(pg_get_serial_sequence('users', 'id'), GREATEST((SELECT MAX(id) FROM users), 1));
-- +goose StatementEnd

-- +goose Down
DELETE FROM users WHERE email IN ('restaurant@test.kz', 'user@test.kz', 'delivery@test.kz');
