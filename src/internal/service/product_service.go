package service

import (
	"context"
	"net/http"
	"regexp"
	"strings"
	"time"

	"emc_lb/src/internal/repository"
	"emc_lb/src/pkg/entities"
	erres "emc_lb/src/pkg/errors"
	"emc_lb/src/pkg/res"
	"emc_lb/src/pkg/utils"
)

type ProductService interface {
	Create(context.Context, entities.CreateProductRequest) (entities.ProductResponse, error)
	List(context.Context) ([]entities.ProductResponse, error)
}

type productService struct {
	productRepository repository.ProductRepository
}

func NewProductService(productRepository repository.ProductRepository) ProductService {
	return &productService{productRepository: productRepository}
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
		CreatedAt:      now,
		UpdatedAt:      now,
	})
	if err != nil {
		return entities.ProductResponse{}, res.WrapError(err, "Can not create product now", erres.CommonInternal)
	}

	return repository.ToProductResponse(product), nil
}

func (s *productService) List(ctx context.Context) ([]entities.ProductResponse, error) {
	products, err := s.productRepository.List(ctx)
	if err != nil {
		return nil, res.WrapError(err, "Can not get products now", erres.CommonInternal)
	}

	responses := make([]entities.ProductResponse, 0, len(products))
	for _, product := range products {
		responses = append(responses, repository.ToProductResponse(product))
	}

	return responses, nil
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
