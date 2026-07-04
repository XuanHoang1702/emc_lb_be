-- name: CreateUser :one
INSERT INTO users (
    id,
    email,
    password_hash,
    user_name,
    phone,
    role,
    status,
    email_verified,
    phone_verified
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    'customer',
    'active',
    false,
    false
)
RETURNING id, email, user_name, phone, created_at;

-- name: GetUserByEmail :one
SELECT
    id,
    email,
    password_hash,
    email_verified,
    role
FROM users
WHERE email = $1 AND is_deleted = false
LIMIT 1;

-- name: GetUserIDByID :one
SELECT id
FROM users
WHERE id = $1 AND is_deleted = false
LIMIT 1;

-- name: GetUserByID :one
SELECT
    id,
    email,
    password_hash,
    email_verified,
    role
FROM users
WHERE id = $1 AND is_deleted = false
LIMIT 1;

-- name: VerifyUserEmail :exec
UPDATE users
SET
    email_verified = true,
    email_verified_at = NOW(),
    updated_at = NOW()
WHERE email = $1;

-- name: SoftDeleteUserByEmail :exec
UPDATE users
SET
    is_deleted = true,
    deleted_at = NOW(),
    updated_at = NOW()
WHERE email = $1 AND is_deleted = false;

-- name: UpdateUserAvatarByEmail :exec
UPDATE users
SET
    avatar_url = $2,
    updated_at = NOW()
WHERE email = $1 AND is_deleted = false;

-- name: UpdateUserAvatarByID :exec
UPDATE users
SET
    avatar_url = $2,
    updated_at = NOW()
WHERE id = $1 AND is_deleted = false;
