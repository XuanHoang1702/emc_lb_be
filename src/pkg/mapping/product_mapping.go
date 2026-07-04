package mapping

import "emc_lb/src/pkg/entities"

func ToProductResponse(product entities.Product) entities.ProductResponse {
	return entities.ProductResponse{
		ID:             product.ID,
		Name:           product.Name,
		Slug:           product.Slug,
		Description:    product.Description,
		ShortDesc:      product.ShortDesc,
		Price:          product.Price,
		OriginalPrice:  product.OriginalPrice,
		SKU:            product.SKU,
		Stock:          product.Stock,
		SoldCount:      product.SoldCount,
		AllowBackorder: product.AllowBackorder,
		Thumbnail:      product.Thumbnail,
		Images:         product.Images,
		Tags:           product.Tags,
		Status:         product.Status,
		IsFeatured:     product.IsFeatured,
		CreatedAt:      product.CreatedAt,
		UpdatedAt:      product.UpdatedAt,
	}
}
