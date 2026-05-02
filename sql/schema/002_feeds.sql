-- +goose Up
CREATE TABLE feeds(
    id UUID PRIMARY KEY DEFAULT (uuidv7()),
    created_at TIMESTAMPTZ NOT NULL
    GENERATED ALWAYS AS (uuid_extract_timestamp("id")) STORED,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    name TEXT NOT NULL,
    url TEXT NOT NULL UNIQUE,
    user_id UUID NOT NULL,
    CONSTRAINT fk_user_id
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);

-- +goose Down
DROP TABLE feeds;
