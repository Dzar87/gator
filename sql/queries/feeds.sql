-- name: CreateFeed :one
INSERT INTO feeds (id, created_at, updated_at, name, url, user_id)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetFeedByURL :one
SELECT * FROM feeds WHERE url = $1;

-- name: GetFeedsWithUser :many
SELECT
    sqlc.embed(f),
    sqlc.embed(u)
FROM feeds f
JOIN users u ON f.user_id = u.id;
