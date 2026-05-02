-- +goose up
CREATE TABLE feed_follows (
    id UUID PRIMARY KEY DEFAULT (uuidv7()),
    created_at TIMESTAMPTZ NOT NULL
    GENERATED ALWAYS AS (uuid_extract_timestamp("id")) STORED,
    updated_at TIMESTAMP NOT NULL,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    feed_id UUID NOT NULL REFERENCES feeds(id) ON DELETE CASCADE,
    CONSTRAINT uniq_user_id_feed_id UNIQUE (user_id, feed_id)
);

-- +goose Down
DROP TABLE feed_follows;
