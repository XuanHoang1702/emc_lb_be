package entities

import "time"

type CartItem struct {
	ProductID string `json:"product_id" bson:"product_id"`
	Quantity  int64  `json:"quantity" bson:"quantity"`
}

type Cart struct {
	ID         string     `json:"id" bson:"_id,omitempty"`
	UserID     string     `json:"user_id" bson:"user_id"`
	Items      []CartItem `json:"items" bson:"items"`
	CouponCode string     `json:"coupon_code" bson:"coupon_code"`
	CreatedAt  time.Time  `json:"created_at" bson:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at" bson:"updated_at"`
}

// Request payloads
type AddCartItemRequest struct {
	ProductID string `json:"product_id" binding:"required"`
	Quantity  int64  `json:"quantity" binding:"required,min=1"`
}

type UpdateCartItemRequest struct {
	Quantity int64 `json:"quantity" binding:"required,min=1"`
}

type ApplyCouponRequest struct {
	CouponCode string `json:"coupon_code" binding:"required"`
}

// Response payload with populated product details and calculated totals
type CartItemResponse struct {
	ProductID    string  `json:"product_id"`
	ProductName  string  `json:"product_name"`
	ProductImage string  `json:"product_image"`
	Price        float64 `json:"price"`
	Quantity     int64   `json:"quantity"`
	SubTotal     float64 `json:"sub_total"`
}

type CartResponse struct {
	ID             string             `json:"id"`
	UserID         string             `json:"user_id"`
	Items          []CartItemResponse `json:"items"`
	SubTotal       float64            `json:"sub_total"`
	CouponCode     string             `json:"coupon_code"`
	DiscountAmount float64            `json:"discount_amount"`
	TaxAmount      float64            `json:"tax_amount"`
	TotalAmount    float64            `json:"total_amount"`
	CreatedAt      time.Time          `json:"created_at"`
	UpdatedAt      time.Time          `json:"updated_at"`
}
