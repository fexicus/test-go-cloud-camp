-- +goose Up
-- +goose StatementBegin
INSERT INTO rate_limits (client_id, capacity, refill_rate) VALUES
    ('127.0.0.1', 10, 1),
    ('127.0.0.2', 20, 2),
    ('127.0.0.3', 5, 1)
    ON CONFLICT (client_id) DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS rate_limits;
-- +goose StatementEnd
