-- name: CreateUser :one
INSERT INTO users (
user_id, full_name, row_number
) VALUES (
$1, $2, $3
)
RETURNING *;

-- -- name: UpdateUser :one
-- UPDATE users SET full_name = $2
-- WHERE user_id = $1
-- RETURNING *;

-- name: GetUser :one
SELECT * fROM users
WHERE user_id = $1;

-- name: ListUsers :many
SELECT * fROM users;

-- name: DeleteUser :exec
DELETE FROM users WHERE user_id = $1;