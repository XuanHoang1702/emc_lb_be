package mapping

import "emc_lb/src/pkg/entities"

func ToCategoryResponse(category entities.Category) entities.CategoryResponse {
	return entities.CategoryResponse{
		ID:              category.ID,
		Name:            category.Name,
		Slug:            category.Slug,
		Description:     category.Description,
		ParentID:        category.ParentID,
		Thumbnail:       category.Thumbnail,
		Banner:          category.Banner,
		MetaTitle:       category.MetaTitle,
		MetaDescription: category.MetaDescription,
		Position:        category.Position,
		IsFeatured:      category.IsFeatured,
		Status:          category.Status,
		CreatedAt:       category.CreatedAt,
		UpdatedAt:       category.UpdatedAt,
	}
}
