package entities

import "time"

type Coupon struct {
	ID                string    `json:"id" bson:"_id,omitempty"`
	Code              string    `json:"code" bson:"code"`
	Type              string    `json:"type" bson:"type"` // "percentage" or "fixed_amount"
	Value             float64   `json:"value" bson:"value"`
	MinOrderAmount    float64   `json:"min_order_amount" bson:"min_order_amount"`
	MaxDiscountAmount float64   `json:"max_discount_amount" bson:"max_discount_amount"` // useful for percentage (e.g. 10% up to 50k)
	UsageLimit        int64     `json:"usage_limit" bson:"usage_limit"`
	UsageCount        int64     `json:"usage_count" bson:"usage_count"`
	StartDate         time.Time `json:"start_date" bson:"start_date"`
	EndDate           time.Time `json:"end_date" bson:"end_date"`
	IsActive          bool      `json:"is_active" bson:"is_active"`
	IsDeleted         bool      `json:"is_deleted" bson:"is_deleted"`
	CreatedAt         time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" bson:"updated_at"`
}

type CreateCouponRequest struct {
	Code              string    `json:"code" binding:"required"`
	Type              string    `json:"type" binding:"required,oneof=percentage fixed_amount"`
	Value             float64   `json:"value" binding:"required,gt=0"`
	MinOrderAmount    float64   `json:"min_order_amount" binding:"min=0"`
	MaxDiscountAmount float64   `json:"max_discount_amount" binding:"min=0"`
	UsageLimit        int64     `json:"usage_limit" binding:"min=0"`
	StartDate         time.Time `json:"start_date" binding:"required"`
	EndDate           time.Time `json:"end_date" binding:"required"`
	IsActive          bool      `json:"is_active"`
}

type UpdateCouponRequest struct {
	Type              string    `json:"type" binding:"omitempty,oneof=percentage fixed_amount"`
	Value             float64   `json:"value" binding:"omitempty,gt=0"`
	MinOrderAmount    float64   `json:"min_order_amount" binding:"omitempty,min=0"`
	MaxDiscountAmount float64   `json:"max_discount_amount" binding:"omitempty,min=0"`
	UsageLimit        int64     `json:"usage_limit" binding:"omitempty,min=0"`
	StartDate         time.Time `json:"start_date"`
	EndDate           time.Time `json:"end_date"`
	IsActive          *bool     `json:"is_active"`
}

type CouponResponse struct {
	ID                string    `json:"id"`
	Code              string    `json:"code"`
	Type              string    `json:"type"`
	Value             float64   `json:"value"`
	MinOrderAmount    float64   `json:"min_order_amount"`
	MaxDiscountAmount float64   `json:"max_discount_amount"`
	UsageLimit        int64     `json:"usage_limit"`
	UsageCount        int64     `json:"usage_count"`
	StartDate         time.Time `json:"start_date"`
	EndDate           time.Time `json:"end_date"`
	IsActive          bool      `json:"is_active"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}
