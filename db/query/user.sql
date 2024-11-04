-- name: CreateUser :one
INSERT INTO users (
telegram_user_id, full_name, row_number
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
WHERE telegram_user_id = $1;

-- name: GetUserByInternalID :one
SELECT * fROM users
WHERE internal_user_id = $1;

-- name: ListUsers :many
SELECT *
FROM users
WHERE telegram_user_id <> -1;

-- name: DeleteUser :exec
DELETE FROM users WHERE internal_user_id = $1;

-- name: ListGuests :many
SELECT 
  * 
FROM users
WHERE telegram_user_id = -1;

