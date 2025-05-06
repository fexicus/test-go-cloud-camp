-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS rate_limits (
    client_id   TEXT    PRIMARY KEY,
    capacity    INTEGER NOT NULL,
    refill_rate INTEGER NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS rate_limits;
-- +goose StatementEnd
