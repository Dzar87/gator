-- +goose Up
CREATE TABLE posts (
    id UUID PRIMARY KEY DEFAULT (uuidv7()),
    created_at TIMESTAMPTZ NOT NULL
    GENERATED ALWAYS AS (uuid_extract_timestamp("id")) STORED,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    title TEXT,
    url TEXT NOT NULL UNIQUE,
    description TEXT,
    published_at TIMESTAMP,
    feed_id UUID NOT NULL REFERENCES feeds(id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE posts;
