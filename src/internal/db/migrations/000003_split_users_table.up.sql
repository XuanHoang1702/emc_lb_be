-- 1. Create extension for UUIDs (if < PG13) and trigger function
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE OR REPLACE FUNCTION trigger_set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 2. Create the new tables
CREATE TABLE users_new (
    id BIGSERIAL PRIMARY KEY,
    uuid UUID NOT NULL DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    password_changed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    email_verified BOOLEAN NOT NULL DEFAULT false,
    email_verified_at TIMESTAMPTZ,
    
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    is_banned BOOLEAN NOT NULL DEFAULT false,
    banned_reason TEXT,
    role VARCHAR(50) NOT NULL DEFAULT 'customer',
    
    last_login_at TIMESTAMPTZ,
    last_login_ip VARCHAR(50),
    failed_login_attempts INT NOT NULL DEFAULT 0,
    locked_until TIMESTAMPTZ,
    
    is_deleted BOOLEAN NOT NULL DEFAULT false,
    deleted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX users_uuid_idx ON users_new(uuid);
CREATE UNIQUE INDEX users_email_partial_idx ON users_new(LOWER(email)) WHERE is_deleted = false;

CREATE TRIGGER set_users_new_updated_at
BEFORE UPDATE ON users_new
FOR EACH ROW EXECUTE FUNCTION trigger_set_updated_at();

CREATE TABLE user_profiles (
    user_id BIGINT PRIMARY KEY REFERENCES users_new(id) ON DELETE CASCADE,
    full_name VARCHAR(255),
    user_name VARCHAR(100) UNIQUE,
    phone VARCHAR(20) UNIQUE,
    phone_verified BOOLEAN NOT NULL DEFAULT false,
    phone_verified_at TIMESTAMPTZ,
    avatar_url TEXT,
    gender VARCHAR(20),
    birth_date DATE,
    language_code VARCHAR(10),
    timezone VARCHAR(50),
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TRIGGER set_user_profiles_updated_at
BEFORE UPDATE ON user_profiles
FOR EACH ROW EXECUTE FUNCTION trigger_set_updated_at();

CREATE TABLE customer_stats (
    user_id BIGINT PRIMARY KEY REFERENCES users_new(id) ON DELETE CASCADE,
    total_orders INT NOT NULL DEFAULT 0,
    total_spent DECIMAL(15,2) NOT NULL DEFAULT 0.00,
    reward_points INT NOT NULL DEFAULT 0,
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TRIGGER set_customer_stats_updated_at
BEFORE UPDATE ON customer_stats
FOR EACH ROW EXECUTE FUNCTION trigger_set_updated_at();

-- 3. Data Migration (using JOIN)
INSERT INTO users_new (
    uuid, email, password_hash, password_changed_at, email_verified, email_verified_at,
    status, is_banned, banned_reason, role, last_login_at, last_login_ip,
    failed_login_attempts, locked_until, is_deleted, deleted_at, created_at, updated_at
)
SELECT 
    id, email, password_hash, password_changed_at, email_verified, email_verified_at,
    status, is_banned, banned_reason, role, last_login_at, last_login_ip,
    failed_login_attempts, locked_until, is_deleted, deleted_at, created_at, updated_at
FROM users;

INSERT INTO user_profiles (
    user_id, full_name, user_name, phone, phone_verified, phone_verified_at,
    avatar_url, gender, birth_date, language_code, timezone, created_at, updated_at
)
SELECT 
    un.id, u.full_name, u.user_name, u.phone, u.phone_verified, u.phone_verified_at,
    u.avatar_url, u.gender, u.birth_date, u.language_code, u.timezone, u.created_at, u.updated_at
FROM users u
JOIN users_new un ON un.uuid = u.id;

INSERT INTO customer_stats (
    user_id, total_orders, total_spent, reward_points, created_at, updated_at
)
SELECT 
    un.id, u.total_orders, u.total_spent, u.reward_points, u.created_at, u.updated_at
FROM users u
JOIN users_new un ON un.uuid = u.id;

-- 4. Swap Tables
ALTER TABLE users RENAME TO users_old;
ALTER TABLE users_new RENAME TO users;
