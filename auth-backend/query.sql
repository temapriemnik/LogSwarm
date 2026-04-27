-- name: CreateUser :one
INSERT INTO users (name, email, password_hash)
VALUES ($1, $2, $3)
RETURNING id, name, email, password_hash, created_at, is_active;

-- name: GetUserByID :one
SELECT id, name, email, password_hash, created_at, is_active
FROM users
WHERE id = $1;

-- name: GetUserByEmail :one
SELECT id, name, email, password_hash, created_at, is_active
FROM users
WHERE lower(email) = lower($1);

-- name: UpdateUserPassword :exec
UPDATE users
SET password_hash = $2
WHERE id = $1;

-- name: ActivateUser :exec
UPDATE users
SET is_active = true
WHERE id = $1;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1;