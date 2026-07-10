package service

import (
	"context"
	"errors"

	"emc_lb/src/internal/repository"
	"emc_lb/src/pkg/entities"

	"github.com/google/uuid"
)

type ShopService interface {
	CreateShop(ctx context.Context, ownerID string, req entities.CreateShopRequest) (entities.ShopResponse, error)
	GetShopByOwnerID(ctx context.Context, ownerID string) (entities.ShopResponse, error)
}

type shopService struct {
	shopRepo repository.ShopRepository
}

func NewShopService(shopRepo repository.ShopRepository) ShopService {
	return &shopService{shopRepo: shopRepo}
}

func (s *shopService) CreateShop(ctx context.Context, ownerID string, req entities.CreateShopRequest) (entities.ShopResponse, error) {
	// Check if user already has a shop
	_, err := s.shopRepo.GetByOwnerID(ctx, ownerID)
	if err == nil {
		return entities.ShopResponse{}, errors.New("user already owns a shop")
	}

	shop := entities.Shop{
		ID:          uuid.New().String(),
		OwnerID:     ownerID,
		Name:        req.Name,
		Slug:        buildSlug(req.Name),
		Description: req.Description,
		Logo:        req.Logo,
		Status:      "active", // automatically active for now
	}

	createdShop, err := s.shopRepo.Create(ctx, shop)
	if err != nil {
		return entities.ShopResponse{}, err
	}

	return toShopResponse(createdShop), nil
}

func (s *shopService) GetShopByOwnerID(ctx context.Context, ownerID string) (entities.ShopResponse, error) {
	shop, err := s.shopRepo.GetByOwnerID(ctx, ownerID)
	if err != nil {
		return entities.ShopResponse{}, err
	}
	return toShopResponse(shop), nil
}

func toShopResponse(shop entities.Shop) entities.ShopResponse {
	return entities.ShopResponse{
		ID:          shop.ID,
		OwnerID:     shop.OwnerID,
		Name:        shop.Name,
		Slug:        shop.Slug,
		Description: shop.Description,
		Logo:        shop.Logo,
		Status:      shop.Status,
		CreatedAt:   shop.CreatedAt,
		UpdatedAt:   shop.UpdatedAt,
	}
}


