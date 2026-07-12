package service

import (
	"context"
	"errors"

	"emc_lb/src/internal/repository"
	"emc_lb/src/pkg/cache"
	"emc_lb/src/pkg/entities"

	"github.com/google/uuid"
)

type ShopService interface {
	CreateShop(ctx context.Context, ownerID string, req entities.CreateShopRequest) (entities.ShopResponse, error)
	GetShopByOwnerID(ctx context.Context, ownerID string) (entities.ShopResponse, error)
}

type shopService struct {
	shopRepo   repository.ShopRepository
	cacheStore cache.ShopCacheStore
}

func NewShopService(shopRepo repository.ShopRepository, cacheStore cache.ShopCacheStore) ShopService {
	return &shopService{shopRepo: shopRepo, cacheStore: cacheStore}
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

	result := toShopResponse(createdShop)
	if s.cacheStore != nil {
		_ = s.cacheStore.InvalidateByOwnerID(ctx, ownerID)
	}

	return result, nil
}

func (s *shopService) GetShopByOwnerID(ctx context.Context, ownerID string) (entities.ShopResponse, error) {
	if s.cacheStore != nil {
		if cached, err := s.cacheStore.GetByOwnerID(ctx, ownerID); err == nil {
			return cached, nil
		}
	}

	shop, err := s.shopRepo.GetByOwnerID(ctx, ownerID)
	if err != nil {
		return entities.ShopResponse{}, err
	}

	result := toShopResponse(shop)
	if s.cacheStore != nil {
		_ = s.cacheStore.SetByOwnerID(ctx, ownerID, result)
	}

	return result, nil
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


