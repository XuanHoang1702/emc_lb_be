DROP INDEX IF EXISTS idx_users_email_unique;
DROP INDEX IF EXISTS idx_users_phone_unique;
DROP INDEX IF EXISTS idx_users_username_unique;

DROP INDEX IF EXISTS idx_users_status;
DROP INDEX IF EXISTS idx_users_role;
DROP INDEX IF EXISTS idx_users_created_at;
DROP INDEX IF EXISTS idx_users_last_login_at;
DROP INDEX IF EXISTS idx_users_deleted;

DROP INDEX IF EXISTS idx_users_status_role;
DROP INDEX IF EXISTS idx_users_email_status;

DROP TABLE IF EXISTS users;