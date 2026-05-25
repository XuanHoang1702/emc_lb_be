package entities

import (
	"time"
)

type Category struct {
	ID string `json:"id"`

	// Basic Information
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description,omitempty"`

	// Hierarchy
	ParentID *string `json:"parent_id,omitempty"`

	// Media
	Thumbnail string `json:"thumbnail,omitempty"`
	Banner    string `json:"banner,omitempty"`

	// SEO
	MetaTitle       string `json:"meta_title,omitempty"`
	MetaDescription string `json:"meta_description,omitempty"`

	// Display
	Position   int64 `json:"position"`
	IsFeatured bool  `json:"is_featured"`

	// Status
	Status    string `json:"status"`
	IsDeleted bool   `json:"is_deleted"`

	// Audit
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt time.Time `json:"deleted_at"`
}

type CreateCategoryRequest struct {
	Name            string `json:"name" binding:"required,min=1,max=255"`
	Slug            string `json:"slug" binding:"omitempty,min=1,max=255,slug"`
	Description     string `json:"description" binding:"max=2000"`
	ParentID        string `json:"parent_id" binding:"omitempty,len=24"`
	Thumbnail       string `json:"thumbnail" binding:"max=2048"`
	Banner          string `json:"banner" binding:"max=2048"`
	MetaTitle       string `json:"meta_title" binding:"max=255"`
	MetaDescription string `json:"meta_description" binding:"max=500"`
	Position        int64  `json:"position"`
	IsFeatured      bool   `json:"is_featured"`
	Status          string `json:"status" binding:"omitempty,oneof=active inactive draft"`
}

type UpdateCategoryRequest struct {
	Name            *string `json:"name" binding:"omitempty,min=1,max=255"`
	Slug            *string `json:"slug" binding:"omitempty,min=1,max=255,slug"`
	Description     *string `json:"description" binding:"omitempty,max=2000"`
	ParentID        *string `json:"parent_id" binding:"omitempty,len=24"`
	Thumbnail       *string `json:"thumbnail" binding:"omitempty,max=2048"`
	Banner          *string `json:"banner" binding:"omitempty,max=2048"`
	MetaTitle       *string `json:"meta_title" binding:"omitempty,max=255"`
	MetaDescription *string `json:"meta_description" binding:"omitempty,max=500"`
	Position        *int64  `json:"position"`
	IsFeatured      *bool   `json:"is_featured"`
	Status          *string `json:"status" binding:"omitempty,oneof=active inactive draft"`
}

type CategoryURIRequest struct {
	ID string `uri:"id" binding:"required,len=24"`
}

type CategoryResponse struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	Slug            string    `json:"slug"`
	Description     string    `json:"description,omitempty"`
	ParentID        *string   `json:"parent_id,omitempty"`
	Thumbnail       string    `json:"thumbnail,omitempty"`
	Banner          string    `json:"banner,omitempty"`
	MetaTitle       string    `json:"meta_title,omitempty"`
	MetaDescription string    `json:"meta_description,omitempty"`
	Position        int64     `json:"position"`
	IsFeatured      bool      `json:"is_featured"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}
