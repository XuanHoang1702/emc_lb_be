-- name: CreateUser :one
INSERT INTO users (
    email,
    password_hash,
    role,
    status,
    email_verified
) VALUES (
    $1, $2, 'customer', 'active', false
)
RETURNING id, uuid, email, created_at;

-- name: CreateUserProfile :exec
INSERT INTO user_profiles (
    user_id,
    user_name,
    phone,
    phone_verified
) VALUES (
    $1, $2, $3, false
);

-- name: CreateCustomerStats :exec
INSERT INTO customer_stats (
    user_id
) VALUES (
    $1
);

-- name: GetUserByEmail :one
SELECT
    id,
    uuid,
    email,
    password_hash,
    email_verified,
    role,
    status,
    is_banned,
    COALESCE(locked_until, '0001-01-01 00:00:00Z'::timestamptz) AS locked_until,
    failed_login_attempts
FROM users
WHERE email = $1 AND is_deleted = false
LIMIT 1;

-- name: GetUserIDByUUID :one
SELECT id
FROM users
WHERE uuid = $1 AND is_deleted = false
LIMIT 1;

-- name: GetUserByUUID :one
SELECT
    id,
    uuid,
    email,
    password_hash,
    email_verified,
    role
FROM users
WHERE uuid = $1 AND is_deleted = false
LIMIT 1;

-- name: VerifyUserEmail :exec
UPDATE users
SET
    email_verified = true,
    email_verified_at = NOW(),
    updated_at = NOW()
WHERE email = $1 AND is_deleted = false;

-- name: SoftDeleteUserByUUID :exec
UPDATE users
SET
    is_deleted = true,
    deleted_at = NOW(),
    updated_at = NOW()
WHERE uuid = $1 AND is_deleted = false;

-- name: UpdateUserAvatarByUUID :exec
UPDATE user_profiles
SET
    avatar_url = $2
WHERE user_id = (SELECT id FROM users WHERE uuid = $1 AND is_deleted = false);

-- name: UpdateFailedLoginAttempts :exec
UPDATE users
SET
    failed_login_attempts = failed_login_attempts + 1
WHERE id = $1 AND is_deleted = false;

-- name: LockUserAccount :exec
UPDATE users
SET
    locked_until = $2
WHERE id = $1 AND is_deleted = false;

-- name: UpdateUserLoginStats :exec
UPDATE users
SET
    failed_login_attempts = 0,
    last_login_at = NOW(),
    last_login_ip = $2
WHERE id = $1 AND is_deleted = false;

-- name: GetUserProfileByUUID :one
SELECT
    u.uuid,
    u.email,
    p.user_name,
    p.full_name,
    p.phone,
    p.avatar_url,
    u.role,
    u.status,
    u.email_verified,
    u.created_at,
    cs.total_orders,
    cs.total_spent,
    cs.reward_points
FROM users u
JOIN user_profiles p ON u.id = p.user_id
JOIN customer_stats cs ON u.id = cs.user_id
WHERE u.uuid = $1 AND u.is_deleted = false
LIMIT 1;
