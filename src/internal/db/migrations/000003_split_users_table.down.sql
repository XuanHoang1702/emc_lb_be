ALTER TABLE users RENAME TO users_new;
ALTER TABLE users_old RENAME TO users;

DROP TABLE IF EXISTS customer_stats;
DROP TABLE IF EXISTS user_profiles;
DROP TABLE IF EXISTS users_new;

DROP FUNCTION IF EXISTS trigger_set_updated_at CASCADE;
