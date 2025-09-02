-- name: CreateUser :one
INSERT INTO users (name, email, password_hash, created_at, updated_at)
VALUES ($1, $2, $3, NOW(), NOW())
RETURNING id, name, email, created_at, updated_at;

-- name: GetUser :one
SELECT id, name, email, created_at, updated_at
FROM users
WHERE id = $1;

-- name: GetUsers :many
SELECT id, name, email, created_at, updated_at
FROM users;

-- name: GetUserByEmail :one
SELECT id, name, email, password_hash, created_at, updated_at
FROM users
WHERE email = $1;

-- name: ListUsers :many
SELECT id, name, email, created_at, updated_at
FROM users
ORDER BY created_at DESC;

-- name: UpdateUser :one
UPDATE users
SET 
    name = $2, 
    email = $3, 
    password_hash = CASE 
        WHEN sqlc.arg(password_hash) != '' THEN sqlc.arg(password_hash)
        ELSE password_hash 
    END,
    updated_at = NOW()
WHERE id = $1
RETURNING id, name, email, created_at, updated_at;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1;
