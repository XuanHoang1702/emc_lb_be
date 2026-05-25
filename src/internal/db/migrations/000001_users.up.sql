CREATE TABLE users (
    id UUID PRIMARY KEY,

    -- ===== Basic Information =====
    email VARCHAR(255) NOT NULL,
    phone VARCHAR(20),
    full_name VARCHAR(255) NULL,
    user_name VARCHAR(100) NOT NULL,

    -- ===== Authentication =====
    password_hash TEXT NOT NULL,
    password_changed_at TIMESTAMPTZ,

    -- ===== Verification =====
    email_verified BOOLEAN NOT NULL DEFAULT FALSE,
    email_verified_at TIMESTAMPTZ,

    phone_verified BOOLEAN NOT NULL DEFAULT FALSE,
    phone_verified_at TIMESTAMPTZ,

    -- ===== Avatar / Profile =====
    avatar_url TEXT,
    gender VARCHAR(20),
    birth_date DATE,

    -- ===== Account Status =====
    status VARCHAR(30) NOT NULL DEFAULT 'active',
    is_banned BOOLEAN NOT NULL DEFAULT FALSE,
    banned_reason TEXT,

    -- ===== Roles =====
    role VARCHAR(30) NOT NULL DEFAULT 'customer',

    -- ===== Loyalty / Ecommerce =====
    total_orders INT NOT NULL DEFAULT 0,
    total_spent NUMERIC(18,2) NOT NULL DEFAULT 0,
    reward_points BIGINT NOT NULL DEFAULT 0,

    -- ===== Security =====
    last_login_at TIMESTAMPTZ,
    last_login_ip INET,

    failed_login_attempts INT NOT NULL DEFAULT 0,
    locked_until TIMESTAMPTZ,

    -- ===== Preferences =====
    language_code VARCHAR(10) DEFAULT 'vi',
    timezone VARCHAR(100) DEFAULT 'Asia/Ho_Chi_Minh',

    -- ===== Soft Delete =====
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    deleted_at TIMESTAMPTZ,

    -- ===== Audit =====
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ===== Unique Indexes =====
CREATE UNIQUE INDEX idx_users_email_unique
ON users(email);

CREATE UNIQUE INDEX idx_users_phone_unique
ON users(phone)
WHERE phone IS NOT NULL;

CREATE UNIQUE INDEX idx_users_username_unique
ON users(user_name)
WHERE user_name IS NOT NULL;

-- ===== Search / Query Indexes =====
CREATE INDEX idx_users_status
ON users(status);

CREATE INDEX idx_users_role
ON users(role);

CREATE INDEX idx_users_created_at
ON users(created_at DESC);

CREATE INDEX idx_users_last_login_at
ON users(last_login_at DESC);

CREATE INDEX idx_users_deleted
ON users(is_deleted);

-- ===== Composite Indexes =====
CREATE INDEX idx_users_status_role
ON users(status, role);

CREATE INDEX idx_users_email_status
ON users(email, status);