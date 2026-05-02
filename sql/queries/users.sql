-- name: CreateUser :one
INSERT INTO users (name)
VALUES ($1)
RETURNING *;
--
-- name: GetUserByName :one
SELECT * FROM users WHERE name = $1;

-- name: DeleteUsers :exec
DELETE FROM users;
--
-- name: GetUsers :many
SELECT * FROM users;
--
