package entities

import (
	"time"
)

type Brand struct {
	ID string `json:"id"`

	// Basic Information
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description,omitempty"`

	// Media
	Logo   string `json:"logo,omitempty"`
	Banner string `json:"banner,omitempty"`

	// Contact / Company
	Website     string `json:"website,omitempty"`
	Email       string `json:"email,omitempty"`
	Phone       string `json:"phone,omitempty"`
	Country     string `json:"country,omitempty"`
	CompanyName string `json:"company_name,omitempty"`

	// SEO
	MetaTitle       string   `json:"meta_title,omitempty"`
	MetaDescription string   `json:"meta_description,omitempty"`
	MetaKeywords    []string `json:"meta_keywords,omitempty"`

	// Display
	Position   int64 `json:"position"`
	IsFeatured bool  `json:"is_featured"`

	// Status
	Status    string `json:"status"` // active, inactive
	IsDeleted bool   `json:"is_deleted"`

	// Audit
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt time.Time `json:"deleted_at"`
}

type CreateBrandRequest struct {
	Name            string   `json:"name" binding:"required,min=1,max=255"`
	Slug            string   `json:"slug" binding:"omitempty,min=1,max=255,slug"`
	Description     string   `json:"description" binding:"max=2000"`
	Logo            string   `json:"logo" binding:"omitempty,max=2048"`
	Banner          string   `json:"banner" binding:"omitempty,max=2048"`
	Website         string   `json:"website" binding:"omitempty,url,max=2048"`
	Email           string   `json:"email" binding:"omitempty,email,max=255"`
	Phone           string   `json:"phone" binding:"omitempty,phone"`
	Country         string   `json:"country" binding:"omitempty,max=120"`
	CompanyName     string   `json:"company_name" binding:"omitempty,max=255"`
	MetaTitle       string   `json:"meta_title" binding:"omitempty,max=255"`
	MetaDescription string   `json:"meta_description" binding:"omitempty,max=500"`
	MetaKeywords    []string `json:"meta_keywords"`
	Position        int64    `json:"position"`
	IsFeatured      bool     `json:"is_featured"`
	Status          string   `json:"status" binding:"omitempty,oneof=active inactive draft"`
}

type UpdateBrandRequest struct {
	Name            *string   `json:"name" binding:"omitempty,min=1,max=255"`
	Slug            *string   `json:"slug" binding:"omitempty,min=1,max=255,slug"`
	Description     *string   `json:"description" binding:"omitempty,max=2000"`
	Logo            *string   `json:"logo" binding:"omitempty,max=2048"`
	Banner          *string   `json:"banner" binding:"omitempty,max=2048"`
	Website         *string   `json:"website" binding:"omitempty,url,max=2048"`
	Email           *string   `json:"email" binding:"omitempty,email,max=255"`
	Phone           *string   `json:"phone" binding:"omitempty,phone"`
	Country         *string   `json:"country" binding:"omitempty,max=120"`
	CompanyName     *string   `json:"company_name" binding:"omitempty,max=255"`
	MetaTitle       *string   `json:"meta_title" binding:"omitempty,max=255"`
	MetaDescription *string   `json:"meta_description" binding:"omitempty,max=500"`
	MetaKeywords    *[]string `json:"meta_keywords"`
	Position        *int64    `json:"position"`
	IsFeatured      *bool     `json:"is_featured"`
	Status          *string   `json:"status" binding:"omitempty,oneof=active inactive draft"`
}

type BrandURIRequest struct {
	ID string `uri:"id" binding:"required,len=24"`
}

type BrandResponse struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	Slug            string    `json:"slug"`
	Description     string    `json:"description,omitempty"`
	Logo            string    `json:"logo,omitempty"`
	Banner          string    `json:"banner,omitempty"`
	Website         string    `json:"website,omitempty"`
	Email           string    `json:"email,omitempty"`
	Phone           string    `json:"phone,omitempty"`
	Country         string    `json:"country,omitempty"`
	CompanyName     string    `json:"company_name,omitempty"`
	MetaTitle       string    `json:"meta_title,omitempty"`
	MetaDescription string    `json:"meta_description,omitempty"`
	MetaKeywords    []string  `json:"meta_keywords,omitempty"`
	Position        int64     `json:"position"`
	IsFeatured      bool      `json:"is_featured"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}
