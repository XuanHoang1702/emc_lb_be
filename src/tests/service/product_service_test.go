package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"emc_lb/src/internal/service"
	"emc_lb/src/pkg/entities"
)

// stubProductCacheStore for testing cache behavior
type stubProductCacheStore struct {
	itemData       map[string]entities.ProductResponse
	invalidateID   string
}

func newStubProductCache() *stubProductCacheStore {
	return &stubProductCacheStore{
		itemData: make(map[string]entities.ProductResponse),
	}
}


func (c *stubProductCacheStore) GetByID(_ context.Context, id string) (entities.ProductResponse, error) {
	if p, ok := c.itemData[id]; ok {
		return p, nil
	}
	return entities.ProductResponse{}, errors.New("cache miss")
}

func (c *stubProductCacheStore) SetByID(_ context.Context, id string, product entities.ProductResponse) error {
	c.itemData[id] = product
	return nil
}

func (c *stubProductCacheStore) Invalidate(_ context.Context, id string) error {
	c.invalidateID = id
	delete(c.itemData, id)
	return nil
}



// stubCategoryRepo for product service tests (just needs ExistsByID)
type stubCategoryRepoForProduct struct{}

func (r *stubCategoryRepoForProduct) Create(_ context.Context, _ entities.Category) (entities.Category, error) {
	return entities.Category{}, nil
}
func (r *stubCategoryRepoForProduct) List(_ context.Context) ([]entities.Category, error) {
	return nil, nil
}
func (r *stubCategoryRepoForProduct) GetByID(_ context.Context, _ string) (entities.Category, error) {
	return entities.Category{}, nil
}
func (r *stubCategoryRepoForProduct) Update(_ context.Context, _ string, _ map[string]any) (entities.Category, error) {
	return entities.Category{}, nil
}
func (r *stubCategoryRepoForProduct) Delete(_ context.Context, _ string, _ map[string]any) error {
	return nil
}
func (r *stubCategoryRepoForProduct) ExistsByID(_ context.Context, _ string) (bool, error) {
	return true, nil
}
func (r *stubCategoryRepoForProduct) ExistsByName(_ context.Context, _ string, _ *string) (bool, error) {
	return false, nil
}
func (r *stubCategoryRepoForProduct) ExistsBySlug(_ context.Context, _ string, _ *string) (bool, error) {
	return false, nil
}
func (r *stubCategoryRepoForProduct) ExistsByPosition(_ context.Context, _ int64, _ *string) (bool, error) {
	return false, nil
}
func (r *stubCategoryRepoForProduct) EnsureIndexes(_ context.Context) error {
	return nil
}

// stubBrandRepo for product service tests
type stubBrandRepo struct{}

func (r *stubBrandRepo) Create(_ context.Context, _ entities.Brand) (entities.Brand, error) {
	return entities.Brand{}, nil
}
func (r *stubBrandRepo) List(_ context.Context) ([]entities.Brand, error) { return nil, nil }
func (r *stubBrandRepo) GetByID(_ context.Context, _ string) (entities.Brand, error) {
	return entities.Brand{}, nil
}
func (r *stubBrandRepo) Update(_ context.Context, _ string, _ map[string]any) (entities.Brand, error) {
	return entities.Brand{}, nil
}
func (r *stubBrandRepo) Delete(_ context.Context, _ string, _ map[string]any) error { return nil }
func (r *stubBrandRepo) ExistsByID(_ context.Context, _ string) (bool, error) {
	return true, nil
}
func (r *stubBrandRepo) ExistsByName(_ context.Context, _ string, _ *string) (bool, error) {
	return false, nil
}
func (r *stubBrandRepo) ExistsBySlug(_ context.Context, _ string, _ *string) (bool, error) {
	return false, nil
}
func (r *stubBrandRepo) ExistsByPosition(_ context.Context, _ int64, _ *string) (bool, error) {
	return false, nil
}
func (r *stubBrandRepo) ExistsByEmail(_ context.Context, _ string, _ *string) (bool, error) {
	return false, nil
}
func (r *stubBrandRepo) EnsureIndexes(_ context.Context) error { return nil }

// listableProductRepo extends stubProductRepository to support List
type listableProductRepo struct {
	products map[string]entities.Product
}

func newListableProductRepo() *listableProductRepo {
	return &listableProductRepo{products: make(map[string]entities.Product)}
}

func (r *listableProductRepo) Create(_ context.Context, p entities.Product) (entities.Product, error) {
	p.ID = "new-product-id"
	r.products[p.ID] = p
	return p, nil
}

func (r *listableProductRepo) List(_ context.Context) ([]entities.Product, error) {
	var result []entities.Product
	for _, p := range r.products {
		result = append(result, p)
	}
	return result, nil
}

func (r *listableProductRepo) GetByID(_ context.Context, id string) (entities.Product, error) {
	p, ok := r.products[id]
	if !ok {
		return entities.Product{}, errors.New("not found")
	}
	return p, nil
}

func (r *listableProductRepo) Update(_ context.Context, id string, data map[string]any) (entities.Product, error) {
	p, ok := r.products[id]
	if !ok {
		return entities.Product{}, errors.New("not found")
	}
	if name, exists := data["name"]; exists {
		p.Name = name.(string)
	}
	p.UpdatedAt = time.Now()
	r.products[id] = p
	return p, nil
}

func (r *listableProductRepo) Delete(_ context.Context, id string, _ map[string]any) error {
	if _, ok := r.products[id]; !ok {
		return errors.New("not found")
	}
	return nil
}

func (r *listableProductRepo) UpdateStock(_ context.Context, id string, delta int64, _ int64) error {
	if p, ok := r.products[id]; ok {
		p.Stock += delta
		r.products[id] = p
	}
	return nil
}

func (r *listableProductRepo) DeductStock(_ context.Context, id string, quantity int64) error {
	p, ok := r.products[id]
	if !ok {
		return errors.New("not found")
	}
	if p.Stock < quantity {
		return errors.New("insufficient stock")
	}
	p.Stock -= quantity
	p.SoldCount += quantity
	r.products[id] = p
	return nil
}

func (r *listableProductRepo) RestoreStockIdempotent(ctx context.Context, id string, quantity int64, orderID string) error {
	if p, ok := r.products[id]; ok {
		p.Stock += quantity
		r.products[id] = p
	}
	return nil
}



func TestUpdate_InvalidatesCache(t *testing.T) {
	productRepo := newListableProductRepo()
	productRepo.products["p1"] = entities.Product{ID: "p1", Name: "Old Name", Price: 100}

	cacheStore := newStubProductCache()
	cacheStore.itemData["p1"] = entities.ProductResponse{ID: "p1", Name: "Old Name"}

	svc := service.NewProductService(productRepo, &stubCategoryRepoForProduct{}, &stubBrandRepo{}, cacheStore, nil)

	newName := "New Name"
	_, err := svc.Update(context.Background(), "p1", entities.UpdateProductRequest{Name: &newName})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cacheStore.invalidateID != "p1" {
		t.Fatalf("expected cache invalidation for p1, got %s", cacheStore.invalidateID)
	}
}

func TestDelete_InvalidatesCache(t *testing.T) {
	productRepo := newListableProductRepo()
	productRepo.products["p1"] = entities.Product{ID: "p1", Name: "To Delete"}

	cacheStore := newStubProductCache()

	svc := service.NewProductService(productRepo, &stubCategoryRepoForProduct{}, &stubBrandRepo{}, cacheStore, nil)

	err := svc.Delete(context.Background(), "p1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cacheStore.invalidateID != "p1" {
		t.Fatalf("expected cache invalidation for p1, got %s", cacheStore.invalidateID)
	}
}
