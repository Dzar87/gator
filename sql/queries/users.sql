-- name: CreateUser :one
INSERT INTO users (id, name)
VALUES ($1, $2)
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
