package entities

import "time"

type Shop struct {
	ID          string    `json:"id" bson:"_id,omitempty"`
	OwnerID     string    `json:"owner_id" bson:"owner_id"`
	Name        string    `json:"name" bson:"name"`
	Slug        string    `json:"slug" bson:"slug"`
	Description string    `json:"description" bson:"description"`
	Logo        string    `json:"logo" bson:"logo"`
	Status      string    `json:"status" bson:"status"` // pending, active, banned
	IsDeleted   bool      `json:"is_deleted" bson:"is_deleted"`
	CreatedAt   time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" bson:"updated_at"`
}

type CreateShopRequest struct {
	Name        string `json:"name" binding:"required,min=3,max=255"`
	Description string `json:"description" binding:"max=2000"`
	Logo        string `json:"logo" binding:"omitempty,max=2048"`
}

type UpdateShopRequest struct {
	Name        *string `json:"name" binding:"omitempty,min=3,max=255"`
	Description *string `json:"description" binding:"omitempty,max=2000"`
	Logo        *string `json:"logo" binding:"omitempty,max=2048"`
}

type ShopResponse struct {
	ID          string    `json:"id"`
	OwnerID     string    `json:"owner_id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description string    `json:"description,omitempty"`
	Logo        string    `json:"logo,omitempty"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
