-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS users (
    id BIGINT NOT NULL PRIMARY KEY,
    email TEXT UNIQUE,
    pwd_hash TEXT,
    status TEXT,
    confirmed BOOLEAN DEFAULT FALSE NOT NULL,
    reg_date TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
