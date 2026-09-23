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
    role,
    status,
    is_banned,
    locked_until,
    failed_login_attempts
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

-- name: SoftDeleteUserByID :exec
UPDATE users
SET
    is_deleted = true,
    deleted_at = NOW(),
    updated_at = NOW()
WHERE id = $1 AND is_deleted = false;

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

-- name: UpdateFailedLoginAttempts :exec
UPDATE users
SET
    failed_login_attempts = failed_login_attempts + 1,
    updated_at = NOW()
WHERE id = $1 AND is_deleted = false;

-- name: LockUserAccount :exec
UPDATE users
SET
    locked_until = $2,
    updated_at = NOW()
WHERE id = $1 AND is_deleted = false;

-- name: UpdateUserLoginStats :exec
UPDATE users
SET
    failed_login_attempts = 0,
    last_login_at = NOW(),
    last_login_ip = $2,
    updated_at = NOW()
WHERE id = $1 AND is_deleted = false;

-- name: GetUserProfileByID :one
SELECT
    id,
    email,
    user_name,
    phone,
    avatar_url,
    role,
    status,
    email_verified,
    created_at
FROM users
WHERE id = $1 AND is_deleted = false
LIMIT 1;
