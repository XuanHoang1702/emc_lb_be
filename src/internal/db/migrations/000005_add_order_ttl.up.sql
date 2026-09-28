ALTER TABLE orders ADD COLUMN expires_at TIMESTAMPTZ;
ALTER TABLE orders ADD COLUMN inventory_returned BOOLEAN NOT NULL DEFAULT false;

CREATE INDEX orders_expires_at_idx ON orders(expires_at) WHERE status = 'pending' AND payment_status = 'unpaid';
