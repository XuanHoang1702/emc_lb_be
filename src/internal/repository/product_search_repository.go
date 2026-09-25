package repository

import (
	"context"
	"fmt"
	
	"emc_lb/src/pkg/entities"
	"emc_lb/src/pkg/logs"
	"github.com/meilisearch/meilisearch-go"
)

type ProductSearchRepository interface {
	IndexProduct(ctx context.Context, product entities.Product) error
	RemoveProduct(ctx context.Context, id string) error
	Search(ctx context.Context, query string, limit int64, offset int64) (*meilisearch.SearchResponse, error)
	InitIndex(ctx context.Context) error
}

type productSearchRepository struct {
	client meilisearch.ServiceManager
	index  string
}

func NewProductSearchRepository(client meilisearch.ServiceManager) ProductSearchRepository {
	return &productSearchRepository{
		client: client,
		index:  "products", // The index name in Meilisearch
	}
}

func (r *productSearchRepository) InitIndex(ctx context.Context) error {
	// Settings for Meilisearch index: Filterable attributes, searchable attributes, typo tolerance
	_, err := r.client.Index(r.index).UpdateSettings(&meilisearch.Settings{
		SearchableAttributes: []string{
			"name",
			"description",
			"short_desc",
			"sku",
			"tags",
		},
		FilterableAttributes: []string{
			"status",
			"is_deleted",
			"category_id",
			"brand_id",
			"price",
			"shop_id",
		},
		SortableAttributes: []string{
			"price",
			"created_at",
			"sold_count",
		},
	})
	if err != nil {
		logs.L().Error("Failed to init Meilisearch settings: " + err.Error())
		return err
	}
	return nil
}

func (r *productSearchRepository) IndexProduct(ctx context.Context, product entities.Product) error {
	// Meilisearch accepts a slice of maps or structs
	// Convert entities.Product to a format suitable for Meilisearch
	doc := map[string]interface{}{
		"id":             product.ID,
		"shop_id":        product.ShopID,
		"name":           product.Name,
		"slug":           product.Slug,
		"description":    product.Description,
		"short_desc":     product.ShortDesc,
		"price":          product.Price,
		"sku":            product.SKU,
		"thumbnail":      product.Thumbnail,
		"category_id":    product.CategoryID,
		"brand_id":       product.BrandID,
		"tags":           product.Tags,
		"status":         product.Status,
		"is_featured":    product.IsFeatured,
		"is_deleted":     product.IsDeleted,
		"average_rating": product.AverageRating,
		"sold_count":     product.SoldCount,
		"created_at":     product.CreatedAt.Unix(),
	}

	task, err := r.client.Index(r.index).AddDocuments([]map[string]interface{}{doc}, nil)
	if err != nil {
		return fmt.Errorf("meilisearch index product failed: %w", err)
	}

	// We don't wait for the task to finish synchronously to keep API fast
	_ = task
	return nil
}

func (r *productSearchRepository) RemoveProduct(ctx context.Context, id string) error {
	_, err := r.client.Index(r.index).DeleteDocument(id, nil)
	if err != nil {
		return fmt.Errorf("meilisearch delete product failed: %w", err)
	}
	return nil
}

func (r *productSearchRepository) Search(ctx context.Context, query string, limit int64, offset int64) (*meilisearch.SearchResponse, error) {
	// Default to active, not deleted products
	searchReq := &meilisearch.SearchRequest{
		Filter: []string{"status = active", "is_deleted = false"},
		Limit:  limit,
		Offset: offset,
	}

	resp, err := r.client.Index(r.index).Search(query, searchReq)
	if err != nil {
		return nil, fmt.Errorf("meilisearch search failed: %w", err)
	}

	return resp, nil
}
