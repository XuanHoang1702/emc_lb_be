package entities

import "time"

type OrderItem struct {
	ProductID   string  `json:"product_id" binding:"required"`
	ProductName string  `json:"product_name"` // Snapshot
	SKU         string  `json:"sku"`          // Snapshot
	Thumbnail   string  `json:"thumbnail"`    // Snapshot
	Quantity    int64   `json:"quantity" binding:"required,gt=0"`
	Price       float64 `json:"price"` // Captured at the time of order
	SubTotal    float64 `json:"sub_total"`
}

type Order struct {
	ID                   string      `json:"id"`
	PaymentGroupID       string      `json:"payment_group_id"` // Used for group payment
	ShopID               string      `json:"shop_id"`
	UserID               string      `json:"user_id"` // User who placed the order
	InvoiceNumber        string      `json:"invoice_number"`
	Items                []OrderItem `json:"items"`
	SubTotal             float64     `json:"sub_total"`
	CouponCode           string      `json:"coupon_code"`
	DiscountAmount       float64     `json:"discount_amount"`
	TaxAmount            float64     `json:"tax_amount"`
	TotalAmount          float64     `json:"total_amount"`
	Status               string      `json:"status"`         // pending, processing, shipped, delivered, cancelled
	PaymentStatus        string      `json:"payment_status"` // unpaid, paid, failed, refunded
	PaymentTransactionID *string     `json:"payment_transaction_id"`
	PaymentMethod        string      `json:"payment_method"`
	ShippingAddress      string      `json:"shipping_address"`
	ContactPhone         string      `json:"contact_phone"`
	IsDeleted            bool        `json:"is_deleted"`
	CreatedAt            time.Time   `json:"created_at"`
	UpdatedAt            time.Time   `json:"updated_at"`
}

type CreateOrderRequest struct {
	Items           []OrderItem `json:"items" binding:"required,min=1,dive"`
	CouponCode      string      `json:"coupon_code"`
	PaymentMethod   string      `json:"payment_method" binding:"required"`
	ShippingAddress string      `json:"shipping_address" binding:"required"`
	ContactPhone    string      `json:"contact_phone" binding:"required"`
}

type OrderResponse struct {
	ID                   string      `json:"id"`
	PaymentGroupID       string      `json:"payment_group_id,omitempty"`
	ShopID               string      `json:"shop_id,omitempty"`
	UserID               string      `json:"user_id"`
	InvoiceNumber        string      `json:"invoice_number"`
	Items                []OrderItem `json:"items"`
	SubTotal             float64     `json:"sub_total"`
	CouponCode           string      `json:"coupon_code"`
	DiscountAmount       float64     `json:"discount_amount"`
	TaxAmount            float64     `json:"tax_amount"`
	TotalAmount          float64     `json:"total_amount"`
	Status               string      `json:"status"`
	PaymentStatus        string      `json:"payment_status"`
	PaymentTransactionID *string     `json:"payment_transaction_id,omitempty"`
	PaymentMethod        string      `json:"payment_method"`
	ShippingAddress      string      `json:"shipping_address"`
	ContactPhone         string      `json:"contact_phone"`
	CreatedAt            time.Time   `json:"created_at"`
	UpdatedAt            time.Time   `json:"updated_at"`
}

type CheckoutItemRequest struct {
	ProductID string `json:"product_id" binding:"required"`
	Quantity  int64  `json:"quantity" binding:"required,gt=0"`
}

type CheckoutRequest struct {
	Items           []CheckoutItemRequest `json:"items" binding:"required,min=1,dive"`
	CouponCode      string                `json:"coupon_code"`
	PaymentMethod   string                `json:"payment_method" binding:"required"`
	ShippingAddress string                `json:"shipping_address" binding:"required"`
	ContactPhone    string                `json:"contact_phone" binding:"required"`
}

type CancelOrderRequest struct {
	Reason string `json:"reason" binding:"required"`
}
