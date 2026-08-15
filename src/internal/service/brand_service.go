package service

import (
	"context"
	"net/http"
	"strings"
	"time"

	"emc_lb/src/internal/repository"
	"emc_lb/src/pkg/entities"
	erres "emc_lb/src/pkg/errors"
	"emc_lb/src/pkg/mapping"
	"emc_lb/src/pkg/res"
	"emc_lb/src/pkg/utils"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

type BrandService interface {
	Create(context.Context, entities.CreateBrandRequest) (entities.BrandResponse, error)
	List(context.Context) ([]entities.BrandResponse, error)
	GetByID(context.Context, string) (entities.BrandResponse, error)
	Update(context.Context, string, entities.UpdateBrandRequest) (entities.BrandResponse, error)
	Delete(context.Context, string) error
}

type brandService struct {
	brandRepository repository.BrandRepository
}

func NewBrandService(brandRepository repository.BrandRepository) BrandService {
	return &brandService{brandRepository: brandRepository}
}

func (s *brandService) Create(ctx context.Context, req entities.CreateBrandRequest) (entities.BrandResponse, error) {
	normalizedRequest := req
	s.normalizeBrandCreateRequest(&normalizedRequest)

	if normalizedRequest.Name == "" {
		return entities.BrandResponse{}, newBrandError("Brand name is required", erres.BrandValidationFailed, http.StatusBadRequest)
	}
	if normalizedRequest.Slug == "" {
		normalizedRequest.Slug = buildSlug(normalizedRequest.Name)
	}
	if normalizedRequest.Status == "" {
		normalizedRequest.Status = "active"
	}
	if err := validateBrandKeywords(normalizedRequest.MetaKeywords); err != nil {
		return entities.BrandResponse{}, err
	}
	if err := s.validateBrandUniqueness(ctx, normalizedRequest.Name, normalizedRequest.Slug, &normalizedRequest.Position, normalizedRequest.Email, nil); err != nil {
		return entities.BrandResponse{}, err
	}

	now := time.Now().UTC()
	brand, err := s.brandRepository.Create(ctx, entities.Brand{
		Name:            normalizedRequest.Name,
		Slug:            normalizedRequest.Slug,
		Description:     normalizedRequest.Description,
		Logo:            normalizedRequest.Logo,
		Banner:          normalizedRequest.Banner,
		Website:         normalizedRequest.Website,
		Email:           normalizedRequest.Email,
		Phone:           normalizedRequest.Phone,
		Country:         normalizedRequest.Country,
		CompanyName:     normalizedRequest.CompanyName,
		MetaTitle:       normalizedRequest.MetaTitle,
		MetaDescription: normalizedRequest.MetaDescription,
		MetaKeywords:    normalizedRequest.MetaKeywords,
		Position:        normalizedRequest.Position,
		IsFeatured:      normalizedRequest.IsFeatured,
		Status:          normalizedRequest.Status,
		IsDeleted:       false,
		CreatedAt:       now,
		UpdatedAt:       now,
	})
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return entities.BrandResponse{}, newBrandError("Brand name, slug, email or position already exists", erres.BrandAlreadyExists, http.StatusConflict)
		}
		return entities.BrandResponse{}, res.WrapError(err, "Can not create brand now", erres.BrandCreateFailed)
	}

	return mapping.ToBrandResponse(brand), nil
}

func (s *brandService) List(ctx context.Context) ([]entities.BrandResponse, error) {
	brands, err := s.brandRepository.List(ctx)
	if err != nil {
		return nil, res.WrapError(err, "Can not get brands now", erres.BrandGetFailed)
	}

	responses := make([]entities.BrandResponse, 0, len(brands))
	for _, brand := range brands {
		responses = append(responses, mapping.ToBrandResponse(brand))
	}

	return responses, nil
}

func (s *brandService) GetByID(ctx context.Context, id string) (entities.BrandResponse, error) {
	if len(id) != 24 {
		return entities.BrandResponse{}, newBrandError("Brand id is invalid", erres.BrandValidationFailed, http.StatusBadRequest)
	}

	brand, err := s.brandRepository.GetByID(ctx, id)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return entities.BrandResponse{}, newBrandError("Brand not found", erres.BrandNotFound, http.StatusNotFound)
		}
		return entities.BrandResponse{}, res.WrapError(err, "Can not get brand now", erres.BrandGetFailed)
	}

	return mapping.ToBrandResponse(brand), nil
}
//nolint:gocyclo
func (s *brandService) Update(ctx context.Context, id string, req entities.UpdateBrandRequest) (entities.BrandResponse, error) {
	if len(id) != 24 {
		return entities.BrandResponse{}, newBrandError("Brand id is invalid", erres.BrandValidationFailed, http.StatusBadRequest)
	}

	update := map[string]any{}
	if req.Name != nil {
		value := strings.TrimSpace(*req.Name)
		if value == "" {
			return entities.BrandResponse{}, newBrandError("Brand name is required", erres.BrandValidationFailed, http.StatusBadRequest)
		}
		update["name"] = value
		if req.Slug == nil {
			update["slug"] = buildSlug(value)
		}
	}
	if req.Slug != nil {
		value := strings.TrimSpace(*req.Slug)
		if value == "" {
			if req.Name != nil {
				update["slug"] = buildSlug(update["name"].(string))
			} else {
				return entities.BrandResponse{}, newBrandError("Brand slug is invalid", erres.BrandValidationFailed, http.StatusBadRequest)
			}
		} else {
			update["slug"] = value
		}
	}
	if req.Description != nil {
		update["description"] = strings.TrimSpace(*req.Description)
	}
	if req.Logo != nil {
		update["logo"] = strings.TrimSpace(*req.Logo)
	}
	if req.Banner != nil {
		update["banner"] = strings.TrimSpace(*req.Banner)
	}
	if req.Website != nil {
		update["website"] = strings.TrimSpace(*req.Website)
	}
	if req.Email != nil {
		email := strings.ToLower(strings.TrimSpace(*req.Email))
		update["email"] = email
	}
	if req.Phone != nil {
		update["phone"] = strings.TrimSpace(*req.Phone)
	}
	if req.Country != nil {
		update["country"] = strings.TrimSpace(*req.Country)
	}
	if req.CompanyName != nil {
		update["company_name"] = strings.TrimSpace(*req.CompanyName)
	}
	if req.MetaTitle != nil {
		update["meta_title"] = strings.TrimSpace(*req.MetaTitle)
	}
	if req.MetaDescription != nil {
		update["meta_description"] = strings.TrimSpace(*req.MetaDescription)
	}
	if req.MetaKeywords != nil {
		metaKeywords := normalizeStringSliceClone(*req.MetaKeywords)
		if err := validateBrandKeywords(metaKeywords); err != nil {
			return entities.BrandResponse{}, err
		}
		update["meta_keywords"] = metaKeywords
	}
	if req.Position != nil {
		update["position"] = *req.Position
	}
	if req.IsFeatured != nil {
		update["is_featured"] = *req.IsFeatured
	}
	if req.Status != nil {
		update["status"] = strings.TrimSpace(*req.Status)
	}
	if len(update) == 0 {
		return entities.BrandResponse{}, newBrandError("No brand field to update", erres.BrandValidationFailed, http.StatusBadRequest)
	}

	derivedSlug := stringPointerValue(req.Slug)
	if derivedSlug == "" && req.Name != nil {
		derivedSlug = buildSlug(strings.TrimSpace(*req.Name))
	}

	if err := s.validateBrandUniqueness(
		ctx,
		stringPointerValue(req.Name),
		derivedSlug,
		req.Position,
		emailPointerValue(req.Email),
		&id,
	); err != nil {
		return entities.BrandResponse{}, err
	}

	update["updated_at"] = time.Now().UTC()

	brand, err := s.brandRepository.Update(ctx, id, update)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return entities.BrandResponse{}, newBrandError("Brand not found", erres.BrandNotFound, http.StatusNotFound)
		}
		if mongo.IsDuplicateKeyError(err) {
			return entities.BrandResponse{}, newBrandError("Brand name, slug, email or position already exists", erres.BrandAlreadyExists, http.StatusConflict)
		}
		return entities.BrandResponse{}, res.WrapError(err, "Can not update brand now", erres.BrandUpdateFailed)
	}

	return mapping.ToBrandResponse(brand), nil
}

func (s *brandService) Delete(ctx context.Context, id string) error {
	if len(id) != 24 {
		return newBrandError("Brand id is invalid", erres.BrandValidationFailed, http.StatusBadRequest)
	}

	now := time.Now().UTC()
	err := s.brandRepository.Delete(ctx, id, map[string]any{
		"is_deleted": true,
		"deleted_at": now,
		"updated_at": now,
	})
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return newBrandError("Brand not found", erres.BrandNotFound, http.StatusNotFound)
		}
		return res.WrapError(err, "Can not delete brand now", erres.BrandDeleteFailed)
	}

	return nil
}

func (s *brandService) validateBrandUniqueness(ctx context.Context, name, slug string, position *int64, email string, excludeID *string) error {
	if name != "" {
		exists, err := s.brandRepository.ExistsByName(ctx, name, excludeID)
		if err != nil {
			return res.WrapError(err, "Can not validate brand name now", erres.BrandGetFailed)
		}
		if exists {
			return newBrandError("Brand name already exists", erres.BrandAlreadyExists, http.StatusConflict)
		}
	}

	if slug != "" {
		exists, err := s.brandRepository.ExistsBySlug(ctx, slug, excludeID)
		if err != nil {
			return res.WrapError(err, "Can not validate brand slug now", erres.BrandGetFailed)
		}
		if exists {
			return newBrandError("Brand slug already exists", erres.BrandAlreadyExists, http.StatusConflict)
		}
	}

	if position != nil {
		exists, err := s.brandRepository.ExistsByPosition(ctx, *position, excludeID)
		if err != nil {
			return res.WrapError(err, "Can not validate brand position now", erres.BrandGetFailed)
		}
		if exists {
			return newBrandError("Brand position already exists", erres.BrandAlreadyExists, http.StatusConflict)
		}
	}

	if email != "" {
		exists, err := s.brandRepository.ExistsByEmail(ctx, email, excludeID)
		if err != nil {
			return res.WrapError(err, "Can not validate brand email now", erres.BrandGetFailed)
		}
		if exists {
			return newBrandError("Brand email already exists", erres.BrandAlreadyExists, http.StatusConflict)
		}
	}

	return nil
}

func (s *brandService) normalizeBrandCreateRequest(req *entities.CreateBrandRequest) {
	utils.NormalizeStrings(
		&req.Name,
		&req.Slug,
		&req.Description,
		&req.Logo,
		&req.Banner,
		&req.Website,
		&req.Phone,
		&req.Country,
		&req.CompanyName,
		&req.MetaTitle,
		&req.MetaDescription,
		&req.Status,
	)
	utils.NormalizeEmail(&req.Email)
	req.MetaKeywords = normalizeStringSliceClone(req.MetaKeywords)
}

func newBrandError(message string, code erres.Code, statusCode int) error {
	return &res.AppError{
		Message:    message,
		Code:       code,
		StatusCode: statusCode,
	}
}

func validateBrandKeywords(values []string) error {
	for _, value := range values {
		if len(value) > 100 {
			return newBrandError("Each meta keyword must not exceed 100 characters", erres.BrandValidationFailed, http.StatusBadRequest)
		}
	}
	if len(values) > 20 {
		return newBrandError("Meta keywords must not exceed 20 items", erres.BrandValidationFailed, http.StatusBadRequest)
	}
	return nil
}

func normalizeStringSliceClone(values []string) []string {
	if values == nil {
		return nil
	}

	normalized := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		normalized = append(normalized, trimmed)
	}

	return normalized
}

func stringPointerValue(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}

func emailPointerValue(value *string) string {
	if value == nil {
		return ""
	}
	return strings.ToLower(strings.TrimSpace(*value))
}
