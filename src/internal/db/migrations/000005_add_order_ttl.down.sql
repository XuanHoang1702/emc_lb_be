DROP INDEX IF EXISTS orders_expires_at_idx;
ALTER TABLE orders DROP COLUMN inventory_returned;
ALTER TABLE orders DROP COLUMN expires_at;
