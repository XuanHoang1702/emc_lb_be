CREATE TABLE orders (
    id BIGSERIAL PRIMARY KEY,
    uuid UUID NOT NULL DEFAULT gen_random_uuid() UNIQUE,
    payment_group_id VARCHAR(50),
    shop_id VARCHAR(50), -- Currently shop_id is string/ObjectID in MongoDB. Keep as VARCHAR for now if shops are in Mongo.
    user_id BIGINT NOT NULL REFERENCES users(id),
    invoice_number VARCHAR(100) NOT NULL UNIQUE,
    sub_total DECIMAL(15,2) NOT NULL DEFAULT 0,
    coupon_code VARCHAR(50),
    discount_amount DECIMAL(15,2) NOT NULL DEFAULT 0,
    tax_amount DECIMAL(15,2) NOT NULL DEFAULT 0,
    total_amount DECIMAL(15,2) NOT NULL DEFAULT 0,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    payment_status VARCHAR(50) NOT NULL DEFAULT 'unpaid',
    payment_method VARCHAR(50),
    payment_transaction_id VARCHAR(100),
    shipping_address TEXT,
    contact_phone VARCHAR(20),
    is_deleted BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX orders_payment_transaction_id_idx ON orders(payment_transaction_id) WHERE payment_transaction_id IS NOT NULL;
CREATE INDEX orders_user_id_idx ON orders(user_id);
CREATE INDEX orders_status_idx ON orders(status);
CREATE INDEX orders_created_at_idx ON orders(created_at DESC);

CREATE TRIGGER set_orders_updated_at
BEFORE UPDATE ON orders
FOR EACH ROW EXECUTE FUNCTION trigger_set_updated_at();

CREATE TABLE order_items (
    id BIGSERIAL PRIMARY KEY,
    uuid UUID NOT NULL DEFAULT gen_random_uuid() UNIQUE,
    order_id BIGINT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id VARCHAR(50) NOT NULL, -- Keeping VARCHAR since products are still in MongoDB
    product_name TEXT NOT NULL,
    sku VARCHAR(100),
    thumbnail TEXT,
    quantity INT NOT NULL,
    price DECIMAL(15,2) NOT NULL,
    sub_total DECIMAL(15,2) NOT NULL
);

CREATE INDEX order_items_order_id_idx ON order_items(order_id);
