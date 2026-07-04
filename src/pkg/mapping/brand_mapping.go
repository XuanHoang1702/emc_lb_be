package mapping

import "emc_lb/src/pkg/entities"

func ToBrandResponse(brand entities.Brand) entities.BrandResponse {
	return entities.BrandResponse{
		ID:              brand.ID,
		Name:            brand.Name,
		Slug:            brand.Slug,
		Description:     brand.Description,
		Logo:            brand.Logo,
		Banner:          brand.Banner,
		Website:         brand.Website,
		Email:           brand.Email,
		Phone:           brand.Phone,
		Country:         brand.Country,
		CompanyName:     brand.CompanyName,
		MetaTitle:       brand.MetaTitle,
		MetaDescription: brand.MetaDescription,
		MetaKeywords:    brand.MetaKeywords,
		Position:        brand.Position,
		IsFeatured:      brand.IsFeatured,
		Status:          brand.Status,
		CreatedAt:       brand.CreatedAt,
		UpdatedAt:       brand.UpdatedAt,
	}
}
