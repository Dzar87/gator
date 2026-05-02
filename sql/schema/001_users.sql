-- +goose Up
CREATE TABLE users(
    id UUID PRIMARY KEY DEFAULT (uuidv7()),
    created_at TIMESTAMPTZ NOT NULL
    GENERATED ALWAYS AS (uuid_extract_timestamp("id")) STORED,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    name TEXT UNIQUE NOT NULL
);
-- +goose Down
DROP TABLE users;
