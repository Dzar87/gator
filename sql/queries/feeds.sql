-- name: CreateFeed :one
INSERT INTO feeds (name, url, user_id)
VALUES ($1, $2, $3)
RETURNING *;
--
-- name: GetFeedByURL :one
SELECT * FROM feeds WHERE url = $1;
--
-- name: GetFeedsWithUser :many
SELECT
    sqlc.embed(f),
    sqlc.embed(u)
FROM feeds f
JOIN users u ON f.user_id = u.id;
--
-- name: MarkFeedFetched :one
UPDATE feeds
SET last_fetched_at = CURRENT_TIMESTAMP, updated_at = DEFAULT
WHERE feeds.id = $1
RETURNING *;
--
-- name: GetNextFeedToFetch :one
SELECT * FROM feeds
ORDER BY feeds.last_fetched_at ASC NULLS FIRST
LIMIT 1;
--
