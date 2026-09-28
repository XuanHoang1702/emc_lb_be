-- name: CreateOrder :one
INSERT INTO orders (
    payment_group_id, shop_id, user_id, invoice_number, sub_total, 
    coupon_code, discount_amount, tax_amount, total_amount, 
    status, payment_status, payment_method, shipping_address, contact_phone, expires_at
) VALUES (
    $1, $2, (SELECT id FROM users WHERE users.uuid = $3), $4, $5, 
    $6, $7, $8, $9, 
    $10, $11, $12, $13, $14, $15
)
RETURNING *;

-- name: CreateOrderItem :one
INSERT INTO order_items (
    order_id, product_id, product_name, sku, thumbnail, quantity, price, sub_total
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
)
RETURNING *;

-- name: ListOrdersByUserID :many
SELECT * FROM orders 
WHERE user_id = (SELECT id FROM users WHERE users.uuid = $1) AND is_deleted = false
ORDER BY created_at DESC;

-- name: ListAllOrders :many
SELECT * FROM orders 
WHERE is_deleted = false
ORDER BY created_at DESC;

-- name: GetOrderItemsByOrderID :many
SELECT * FROM order_items 
WHERE order_id = $1;

-- name: GetOrderByUUID :one
SELECT o.*, u.uuid as user_uuid 
FROM orders o
JOIN users u ON o.user_id = u.id
WHERE o.uuid = $1 AND o.is_deleted = false;

-- name: GetOrderByInvoiceNumber :one
SELECT o.*, u.uuid as user_uuid 
FROM orders o
JOIN users u ON o.user_id = u.id
WHERE o.invoice_number = $1 AND o.is_deleted = false;

-- name: UpdateOrderStatus :exec
UPDATE orders 
SET status = $2 
WHERE uuid = $1 AND is_deleted = false;

-- name: ConfirmPaymentAtomic :execrows
UPDATE orders
SET payment_status = $2, 
    status = $3, 
    payment_transaction_id = $4,
    updated_at = NOW()
WHERE invoice_number = $1 
  AND payment_status = 'unpaid' 
  AND is_deleted = false;

-- name: UpdateOrderStatusAtomic :execrows
UPDATE orders
SET status = $3, updated_at = NOW()
WHERE uuid = $1 AND status = $2 AND is_deleted = false;

-- name: MarkInventoryReturned :execrows
UPDATE orders
SET inventory_returned = true, updated_at = NOW()
WHERE uuid = $1 AND inventory_returned = false;

-- name: GetExpiredPendingOrders :many
SELECT * FROM orders
WHERE status = 'pending' 
  AND payment_status = 'unpaid' 
  AND expires_at < NOW() 
  AND is_deleted = false
ORDER BY created_at ASC
LIMIT 100;
