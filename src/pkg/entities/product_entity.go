package entities

import (
	"time"
)

type Product struct {
	ID     string `json:"id"`
	ShopID string `json:"shop_id"`

	// Basic Information
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	ShortDesc   string `json:"short_desc"`

	// Pricing
	Price         float64 `json:"price"`
	OriginalPrice float64 `json:"original_price"`
	CostPrice     float64 `json:"cost_price"`

	// Inventory
	SKU            string `json:"sku"`
	Barcode        string `json:"barcode"`
	Stock          int64  `json:"stock"`
	SoldCount      int64  `json:"sold_count"`
	AllowBackorder bool   `json:"allow_backorder"`

	// Media
	Thumbnail string   `json:"thumbnail"`
	Images    []string `json:"images"`

	// Category & Brand
	CategoryID string `json:"category_id"`
	BrandID    string `json:"brand_id"`

	// Attributes / Variants
	Attributes map[string]string `json:"attributes"`
	Tags       []string          `json:"tags"`

	// Shipping
	Weight float64 `json:"weight"`
	Length float64 `json:"length"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`

	// SEO
	MetaTitle       string `json:"meta_title"`
	MetaDescription string `json:"meta_description"`

	// Status
	Status     string `json:"status"` // active, draft, inactive
	IsFeatured bool   `json:"is_featured"`
	IsDeleted  bool   `json:"is_deleted"`

	// Rating
	AverageRating float64 `json:"average_rating"`
	ReviewCount   int64   `json:"review_count"`

	// Audit
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateProductRequest struct {
	Name           string   `json:"name" binding:"required,min=1,max=255"`
	Slug           string   `json:"slug" binding:"omitempty,min=1,max=255,slug"`
	Description    string   `json:"description" binding:"max=2000"`
	ShortDesc      string   `json:"short_desc" binding:"max=500"`
	Price          float64  `json:"price" binding:"required,gte=0"`
	OriginalPrice  float64  `json:"original_price" binding:"gte=0"`
	SKU            string   `json:"sku" binding:"max=100"`
	Stock          int64    `json:"stock" binding:"gte=0"`
	AllowBackorder bool     `json:"allow_backorder"`
	Thumbnail      string   `json:"thumbnail" binding:"max=2048"`
	Images         []string `json:"images"`
	Tags           []string `json:"tags"`
	Status         string   `json:"status" binding:"omitempty,oneof=active draft inactive"`
	IsFeatured     bool     `json:"is_featured"`
	CategoryID     string   `json:"category_id" binding:"omitempty"`
	BrandID        string   `json:"brand_id" binding:"omitempty"`
	ShopID         string   `json:"shop_id" binding:"omitempty"`
}

type UpdateProductRequest struct {
	Name           *string   `json:"name" binding:"omitempty,min=1,max=255"`
	Slug           *string   `json:"slug" binding:"omitempty,min=1,max=255,slug"`
	Description    *string   `json:"description" binding:"omitempty,max=2000"`
	ShortDesc      *string   `json:"short_desc" binding:"omitempty,max=500"`
	Price          *float64  `json:"price" binding:"omitempty,gte=0"`
	OriginalPrice  *float64  `json:"original_price" binding:"omitempty,gte=0"`
	SKU            *string   `json:"sku" binding:"omitempty,max=100"`
	Stock          *int64    `json:"stock" binding:"omitempty,gte=0"`
	AllowBackorder *bool     `json:"allow_backorder"`
	Thumbnail      *string   `json:"thumbnail" binding:"omitempty,max=2048"`
	Images         []string  `json:"images"`
	Tags           []string  `json:"tags"`
	Status         *string   `json:"status" binding:"omitempty,oneof=active draft inactive"`
	IsFeatured     *bool     `json:"is_featured"`
	CategoryID     *string   `json:"category_id" binding:"omitempty"`
	BrandID        *string   `json:"brand_id" binding:"omitempty"`
	ShopID         *string   `json:"shop_id" binding:"omitempty"`
}

type ProductResponse struct {
	ID             string    `json:"id"`
	ShopID         string    `json:"shop_id,omitempty"`
	Name           string    `json:"name"`
	Slug           string    `json:"slug"`
	Description    string    `json:"description,omitempty"`
	ShortDesc      string    `json:"short_desc,omitempty"`
	Price          float64   `json:"price"`
	OriginalPrice  float64   `json:"original_price,omitempty"`
	SKU            string    `json:"sku,omitempty"`
	Stock          int64     `json:"stock"`
	SoldCount      int64     `json:"sold_count"`
	AllowBackorder bool      `json:"allow_backorder"`
	Thumbnail      string    `json:"thumbnail,omitempty"`
	Images         []string  `json:"images,omitempty"`
	Tags           []string  `json:"tags,omitempty"`
	Status         string    `json:"status"`
	IsFeatured     bool      `json:"is_featured"`
	CategoryID     string    `json:"category_id,omitempty"`
	BrandID        string    `json:"brand_id,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
