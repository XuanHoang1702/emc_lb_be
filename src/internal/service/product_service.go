package service

import (
	"context"
	"net/http"
	"regexp"
	"strings"
	"time"

	"emc_lb/src/internal/repository"
	"emc_lb/src/pkg/cache"
	"emc_lb/src/pkg/entities"
	erres "emc_lb/src/pkg/errors"
	"emc_lb/src/pkg/mapping"
	"emc_lb/src/pkg/res"
	"emc_lb/src/pkg/utils"

	"golang.org/x/sync/singleflight"
)

type ProductService interface {
	Create(context.Context, entities.CreateProductRequest) (entities.ProductResponse, error)
	List(context.Context) ([]entities.ProductResponse, error)
	GetByID(context.Context, string) (entities.ProductResponse, error)
	Update(context.Context, string, entities.UpdateProductRequest) (entities.ProductResponse, error)
	Delete(context.Context, string) error
}

type productService struct {
	productRepository  repository.ProductRepository
	categoryRepository repository.CategoryRepository
	brandRepository    repository.BrandRepository
	cacheStore         cache.ProductCacheStore
	sg                 singleflight.Group
}

func NewProductService(productRepository repository.ProductRepository, categoryRepository repository.CategoryRepository, brandRepository repository.BrandRepository, cacheStore cache.ProductCacheStore) ProductService {
	return &productService{
		productRepository:  productRepository,
		categoryRepository: categoryRepository,
		brandRepository:    brandRepository,
		cacheStore:         cacheStore,
	}
}

func (s *productService) Create(ctx context.Context, req entities.CreateProductRequest) (entities.ProductResponse, error) {
	normalizedRequest := req
	utils.NormalizeStrings(
		&normalizedRequest.Name,
		&normalizedRequest.Slug,
		&normalizedRequest.Description,
		&normalizedRequest.ShortDesc,
		&normalizedRequest.SKU,
		&normalizedRequest.Thumbnail,
	)
	normalizeStringSlice(normalizedRequest.Images)
	normalizeStringSlice(normalizedRequest.Tags)

	if normalizedRequest.Name == "" {
		return entities.ProductResponse{}, &res.AppError{
			Message:    "Product name is required",
			Code:       erres.CommonBadRequest,
			StatusCode: http.StatusBadRequest,
		}
	}
	if normalizedRequest.Slug == "" {
		normalizedRequest.Slug = buildSlug(normalizedRequest.Name)
	}
	if normalizedRequest.Status == "" {
		normalizedRequest.Status = "active"
	}

	if normalizedRequest.CategoryID != "" {
		exists, err := s.categoryRepository.ExistsByID(ctx, normalizedRequest.CategoryID)
		if err != nil || !exists {
			return entities.ProductResponse{}, &res.AppError{
				Message:    "Category does not exist",
				Code:       erres.CommonBadRequest,
				StatusCode: http.StatusBadRequest,
			}
		}
	}

	if normalizedRequest.BrandID != "" {
		exists, err := s.brandRepository.ExistsByID(ctx, normalizedRequest.BrandID)
		if err != nil || !exists {
			return entities.ProductResponse{}, &res.AppError{
				Message:    "Brand does not exist",
				Code:       erres.CommonBadRequest,
				StatusCode: http.StatusBadRequest,
			}
		}
	}

	now := time.Now().UTC()
	product, err := s.productRepository.Create(ctx, entities.Product{
		Name:           normalizedRequest.Name,
		Slug:           normalizedRequest.Slug,
		Description:    normalizedRequest.Description,
		ShortDesc:      normalizedRequest.ShortDesc,
		Price:          normalizedRequest.Price,
		OriginalPrice:  normalizedRequest.OriginalPrice,
		SKU:            normalizedRequest.SKU,
		Stock:          normalizedRequest.Stock,
		AllowBackorder: normalizedRequest.AllowBackorder,
		Thumbnail:      normalizedRequest.Thumbnail,
		Images:         normalizedRequest.Images,
		Tags:           normalizedRequest.Tags,
		Status:         normalizedRequest.Status,
		IsFeatured:     normalizedRequest.IsFeatured,
		CategoryID:     normalizedRequest.CategoryID,
		BrandID:        normalizedRequest.BrandID,
		ShopID:         normalizedRequest.ShopID,
		CreatedAt:      now,
		UpdatedAt:      now,
	})
	if err != nil {
		return entities.ProductResponse{}, res.WrapError(err, "Can not create product now", erres.CommonInternal)
	}

	if s.cacheStore != nil {
		_ = s.cacheStore.InvalidateList(ctx)
	}

	return mapping.ToProductResponse(product), nil
}

func (s *productService) List(ctx context.Context) ([]entities.ProductResponse, error) {
	queryHash := "default" // In the future, this would be a hash of pagination/filter params

	if s.cacheStore != nil {
		if cached, err := s.cacheStore.GetList(ctx, queryHash); err == nil {
			return cached, nil
		}
	}

	// Singleflight for list
	sgKey := "list:" + queryHash
	result, err, _ := s.sg.Do(sgKey, func() (interface{}, error) {
		products, err := s.productRepository.List(ctx)
		if err != nil {
			return nil, res.WrapError(err, "Can not get products now", erres.CommonInternal)
		}

		responses := make([]entities.ProductResponse, 0, len(products))
		for _, product := range products {
			responses = append(responses, mapping.ToProductResponse(product))
		}

		if s.cacheStore != nil {
			_ = s.cacheStore.SetList(ctx, queryHash, responses)
		}

		return responses, nil
	})

	if err != nil {
		return nil, err
	}

	return result.([]entities.ProductResponse), nil
}

func (s *productService) GetByID(ctx context.Context, id string) (entities.ProductResponse, error) {
	if s.cacheStore != nil {
		if cached, err := s.cacheStore.GetByID(ctx, id); err == nil {
			return cached, nil
		}
	}

	// Singleflight for detail
	sgKey := "detail:" + id
	result, err, _ := s.sg.Do(sgKey, func() (interface{}, error) {
		product, err := s.productRepository.GetByID(ctx, id)
		if err != nil {
			return entities.ProductResponse{}, res.WrapError(err, "Product not found", erres.CommonNotFound)
		}

		response := mapping.ToProductResponse(product)
		if s.cacheStore != nil {
			_ = s.cacheStore.SetByID(ctx, id, response)
		}

		return response, nil
	})

	if err != nil {
		return entities.ProductResponse{}, err
	}

	return result.(entities.ProductResponse), nil
}

//nolint:gocyclo
func (s *productService) Update(ctx context.Context, id string, req entities.UpdateProductRequest) (entities.ProductResponse, error) {
	updateData := make(map[string]any)

	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			return entities.ProductResponse{}, &res.AppError{
				Message:    "Product name cannot be empty",
				Code:       erres.CommonBadRequest,
				StatusCode: http.StatusBadRequest,
			}
		}
		updateData["name"] = name
		if req.Slug == nil {
			updateData["slug"] = buildSlug(name)
		}
	}

	if req.Slug != nil {
		updateData["slug"] = buildSlug(*req.Slug)
	}
	if req.Description != nil {
		updateData["description"] = strings.TrimSpace(*req.Description)
	}
	if req.ShortDesc != nil {
		updateData["short_desc"] = strings.TrimSpace(*req.ShortDesc)
	}
	if req.Price != nil {
		updateData["price"] = *req.Price
	}
	if req.OriginalPrice != nil {
		updateData["original_price"] = *req.OriginalPrice
	}
	if req.SKU != nil {
		updateData["sku"] = strings.TrimSpace(*req.SKU)
	}
	if req.Stock != nil {
		updateData["stock"] = *req.Stock
	}
	if req.AllowBackorder != nil {
		updateData["allow_backorder"] = *req.AllowBackorder
	}
	if req.Thumbnail != nil {
		updateData["thumbnail"] = strings.TrimSpace(*req.Thumbnail)
	}
	if req.Images != nil {
		normalizeStringSlice(req.Images)
		updateData["images"] = req.Images
	}
	if req.Tags != nil {
		normalizeStringSlice(req.Tags)
		updateData["tags"] = req.Tags
	}
	if req.Status != nil {
		updateData["status"] = *req.Status
	}
	if req.IsFeatured != nil {
		updateData["is_featured"] = *req.IsFeatured
	}
	if req.CategoryID != nil {
		if *req.CategoryID != "" {
			exists, err := s.categoryRepository.ExistsByID(ctx, *req.CategoryID)
			if err != nil || !exists {
				return entities.ProductResponse{}, &res.AppError{
					Message:    "Category does not exist",
					Code:       erres.CommonBadRequest,
					StatusCode: http.StatusBadRequest,
				}
			}
		}
		updateData["category_id"] = *req.CategoryID
	}
	if req.BrandID != nil {
		if *req.BrandID != "" {
			exists, err := s.brandRepository.ExistsByID(ctx, *req.BrandID)
			if err != nil || !exists {
				return entities.ProductResponse{}, &res.AppError{
					Message:    "Brand does not exist",
					Code:       erres.CommonBadRequest,
					StatusCode: http.StatusBadRequest,
				}
			}
		}
		updateData["brand_id"] = *req.BrandID
	}
	if req.ShopID != nil {
		updateData["shop_id"] = *req.ShopID
	}

	updateData["updated_at"] = time.Now().UTC()

	product, err := s.productRepository.Update(ctx, id, updateData)
	if err != nil {
		return entities.ProductResponse{}, res.WrapError(err, "Failed to update product", erres.CommonInternal)
	}

	if s.cacheStore != nil {
		_ = s.cacheStore.Invalidate(ctx, id)
	}

	return mapping.ToProductResponse(product), nil
}

func (s *productService) Delete(ctx context.Context, id string) error {
	updateData := map[string]any{
		"is_deleted": true,
		"updated_at": time.Now().UTC(),
	}
	err := s.productRepository.Delete(ctx, id, updateData)
	if err != nil {
		return res.WrapError(err, "Failed to delete product", erres.CommonInternal)
	}

	if s.cacheStore != nil {
		_ = s.cacheStore.Invalidate(ctx, id)
	}

	return nil
}

func normalizeStringSlice(values []string) {
	for i := range values {
		values[i] = strings.TrimSpace(values[i])
	}
}

var nonSlugChars = regexp.MustCompile(`[^a-z0-9-]+`)
var multiHyphens = regexp.MustCompile(`-+`)

func buildSlug(name string) string {
	normalized := strings.ToLower(strings.TrimSpace(name))
	normalized = strings.ReplaceAll(normalized, " ", "-")
	normalized = nonSlugChars.ReplaceAllString(normalized, "-")
	normalized = multiHyphens.ReplaceAllString(normalized, "-")
	return strings.Trim(normalized, "-")
}
