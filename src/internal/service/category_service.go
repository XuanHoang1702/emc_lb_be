package service

import (
	"context"
	"net/http"
	"strings"
	"time"

	"emc_lb/src/internal/repository"
	"emc_lb/src/pkg/entities"
	erres "emc_lb/src/pkg/errors"
	"emc_lb/src/pkg/res"
	"emc_lb/src/pkg/utils"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

type CategoryService interface {
	Create(context.Context, entities.CreateCategoryRequest) (entities.CategoryResponse, error)
	List(context.Context) ([]entities.CategoryResponse, error)
	GetByID(context.Context, string) (entities.CategoryResponse, error)
	Update(context.Context, string, entities.UpdateCategoryRequest) (entities.CategoryResponse, error)
	Delete(context.Context, string) error
}

type categoryService struct {
	categoryRepository repository.CategoryRepository
}

func NewCategoryService(categoryRepository repository.CategoryRepository) CategoryService {
	return &categoryService{categoryRepository: categoryRepository}
}

func (s *categoryService) Create(ctx context.Context, req entities.CreateCategoryRequest) (entities.CategoryResponse, error) {
	normalizedRequest := req
	utils.NormalizeStrings(
		&normalizedRequest.Name,
		&normalizedRequest.Slug,
		&normalizedRequest.Description,
		&normalizedRequest.ParentID,
		&normalizedRequest.Thumbnail,
		&normalizedRequest.Banner,
		&normalizedRequest.MetaTitle,
		&normalizedRequest.MetaDescription,
		&normalizedRequest.Status,
	)

	if normalizedRequest.Name == "" {
		return entities.CategoryResponse{}, newBadRequestError("Category name is required")
	}
	if normalizedRequest.Slug == "" {
		normalizedRequest.Slug = buildSlug(normalizedRequest.Name)
	}
	if normalizedRequest.Status == "" {
		normalizedRequest.Status = "active"
	}
	if err := s.validateCategoryUniqueness(ctx, normalizedRequest.Name, normalizedRequest.Slug, &normalizedRequest.Position, nil); err != nil {
		return entities.CategoryResponse{}, err
	}

	parentID, err := s.parseParentID(ctx, normalizedRequest.ParentID, "")
	if err != nil {
		return entities.CategoryResponse{}, err
	}

	now := time.Now().UTC()
	category, err := s.categoryRepository.Create(ctx, entities.Category{
		Name:            normalizedRequest.Name,
		Slug:            normalizedRequest.Slug,
		Description:     normalizedRequest.Description,
		ParentID:        parentID,
		Thumbnail:       normalizedRequest.Thumbnail,
		Banner:          normalizedRequest.Banner,
		MetaTitle:       normalizedRequest.MetaTitle,
		MetaDescription: normalizedRequest.MetaDescription,
		Position:        normalizedRequest.Position,
		IsFeatured:      normalizedRequest.IsFeatured,
		Status:          normalizedRequest.Status,
		IsDeleted:       false,
		CreatedAt:       now,
		UpdatedAt:       now,
	})
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return entities.CategoryResponse{}, newConflictError("Category name, slug or position already exists")
		}
		return entities.CategoryResponse{}, res.WrapError(err, "Can not create category now", erres.CommonInternal)
	}

	return repository.ToCategoryResponse(category), nil
}

func (s *categoryService) List(ctx context.Context) ([]entities.CategoryResponse, error) {
	categories, err := s.categoryRepository.List(ctx)
	if err != nil {
		return nil, res.WrapError(err, "Can not get categories now", erres.CommonInternal)
	}

	responses := make([]entities.CategoryResponse, 0, len(categories))
	for _, category := range categories {
		responses = append(responses, repository.ToCategoryResponse(category))
	}

	return responses, nil
}

func (s *categoryService) GetByID(ctx context.Context, id string) (entities.CategoryResponse, error) {
	if len(id) != 24 {
		return entities.CategoryResponse{}, newBadRequestError("Category id is invalid")
	}

	category, err := s.categoryRepository.GetByID(ctx, id)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return entities.CategoryResponse{}, newNotFoundError("Category not found")
		}
		return entities.CategoryResponse{}, res.WrapError(err, "Can not get category now", erres.CommonInternal)
	}

	return repository.ToCategoryResponse(category), nil
}

func (s *categoryService) Update(ctx context.Context, id string, req entities.UpdateCategoryRequest) (entities.CategoryResponse, error) {
	if len(id) != 24 {
		return entities.CategoryResponse{}, newBadRequestError("Category id is invalid")
	}

	update := map[string]any{}
	if req.Name != nil {
		value := strings.TrimSpace(*req.Name)
		if value == "" {
			return entities.CategoryResponse{}, newBadRequestError("Category name is required")
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
				return entities.CategoryResponse{}, newBadRequestError("Category slug is invalid")
			}
		} else {
			update["slug"] = value
		}
	}
	if req.Description != nil {
		update["description"] = strings.TrimSpace(*req.Description)
	}
	if req.Thumbnail != nil {
		update["thumbnail"] = strings.TrimSpace(*req.Thumbnail)
	}
	if req.Banner != nil {
		update["banner"] = strings.TrimSpace(*req.Banner)
	}
	if req.MetaTitle != nil {
		update["meta_title"] = strings.TrimSpace(*req.MetaTitle)
	}
	if req.MetaDescription != nil {
		update["meta_description"] = strings.TrimSpace(*req.MetaDescription)
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
	if req.ParentID != nil {
		parentID, err := s.parseParentID(ctx, strings.TrimSpace(*req.ParentID), id)
		if err != nil {
			return entities.CategoryResponse{}, err
		}
		update["parent_id"] = parentID
	}
	if len(update) == 0 {
		return entities.CategoryResponse{}, newBadRequestError("No category field to update")
	}
	if err := s.validateCategoryUniqueness(
		ctx,
		stringValueOrFallback(req.Name),
		stringValueOrFallback(req.Slug),
		req.Position,
		&id,
	); err != nil {
		return entities.CategoryResponse{}, err
	}

	update["updated_at"] = time.Now().UTC()

	category, err := s.categoryRepository.Update(ctx, id, update)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return entities.CategoryResponse{}, newNotFoundError("Category not found")
		}
		if mongo.IsDuplicateKeyError(err) {
			return entities.CategoryResponse{}, newConflictError("Category name, slug or position already exists")
		}
		return entities.CategoryResponse{}, res.WrapError(err, "Can not update category now", erres.CommonInternal)
	}

	return repository.ToCategoryResponse(category), nil
}

func (s *categoryService) Delete(ctx context.Context, id string) error {
	if len(id) != 24 {
		return newBadRequestError("Category id is invalid")
	}

	now := time.Now().UTC()
	err := s.categoryRepository.Delete(ctx, id, map[string]any{
		"is_deleted": true,
		"deleted_at": now,
		"updated_at": now,
	})
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return newNotFoundError("Category not found")
		}
		return res.WrapError(err, "Can not delete category now", erres.CommonInternal)
	}

	return nil
}

func (s *categoryService) parseParentID(ctx context.Context, parentIDValue string, currentID string) (*string, error) {
	if parentIDValue == "" {
		return nil, nil
	}
	if currentID != "" && parentIDValue == currentID {
		return nil, newBadRequestError("Category parent can not be itself")
	}

	if len(parentIDValue) != 24 {
		return nil, newBadRequestError("Parent category id is invalid")
	}

	exists, err := s.categoryRepository.ExistsByID(ctx, parentIDValue)
	if err != nil {
		return nil, res.WrapError(err, "Can not validate parent category now", erres.CommonInternal)
	}
	if !exists {
		return nil, newNotFoundError("Parent category not found")
	}

	return &parentIDValue, nil
}

func newBadRequestError(message string) error {
	return &res.AppError{
		Message:    message,
		Code:       erres.CommonBadRequest,
		StatusCode: http.StatusBadRequest,
	}
}

func newNotFoundError(message string) error {
	return &res.AppError{
		Message:    message,
		Code:       erres.CommonNotFound,
		StatusCode: http.StatusNotFound,
	}
}

func newConflictError(message string) error {
	return &res.AppError{
		Message:    message,
		Code:       erres.CommonConflict,
		StatusCode: http.StatusConflict,
	}
}

func (s *categoryService) validateCategoryUniqueness(ctx context.Context, name, slug string, position *int64, excludeID *string) error {
	if name != "" {
		exists, err := s.categoryRepository.ExistsByName(ctx, name, excludeID)
		if err != nil {
			return res.WrapError(err, "Can not validate category name now", erres.CommonInternal)
		}
		if exists {
			return newConflictError("Category name already exists")
		}
	}

	if slug != "" {
		exists, err := s.categoryRepository.ExistsBySlug(ctx, slug, excludeID)
		if err != nil {
			return res.WrapError(err, "Can not validate category slug now", erres.CommonInternal)
		}
		if exists {
			return newConflictError("Category slug already exists")
		}
	}

	if position != nil {
		exists, err := s.categoryRepository.ExistsByPosition(ctx, *position, excludeID)
		if err != nil {
			return res.WrapError(err, "Can not validate category position now", erres.CommonInternal)
		}
		if exists {
			return newConflictError("Category position already exists")
		}
	}

	return nil
}

func stringValueOrFallback(value *string) string {
	if value != nil {
		return strings.TrimSpace(*value)
	}
	return ""
}
