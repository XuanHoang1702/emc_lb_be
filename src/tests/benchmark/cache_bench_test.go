package benchmark

import (
	"context"
	"errors"
	"testing"
	"time"

	"emc_lb/src/internal/service"
	"emc_lb/src/pkg/entities"
)

// mockProductRepo for benchmarking
type mockProductRepo struct{}

func (m *mockProductRepo) GetByID(ctx context.Context, id string) (entities.Product, error) {
	// Simulate DB latency
	time.Sleep(10 * time.Millisecond)
	return entities.Product{
		ID:    id,
		Name:  "Test Product",
		Price: 100,
	}, nil
}

func (m *mockProductRepo) Create(ctx context.Context, p entities.Product) (entities.Product, error) { return p, nil }
func (m *mockProductRepo) List(ctx context.Context) ([]entities.Product, error) { return nil, nil }
func (m *mockProductRepo) Update(ctx context.Context, id string, data map[string]any) (entities.Product, error) { return entities.Product{}, nil }
func (m *mockProductRepo) Delete(ctx context.Context, id string, data map[string]any) error { return nil }
func (m *mockProductRepo) UpdateStock(ctx context.Context, id string, stockDelta int64, soldDelta int64) error { return nil }
func (m *mockProductRepo) DeductStock(ctx context.Context, id string, quantity int64) error { return nil }

// mockCacheStore
type mockCacheStore struct {
	hit  bool
	data entities.ProductResponse
}

func (m *mockCacheStore) GetByID(ctx context.Context, id string) (entities.ProductResponse, error) {
	if m.hit {
		// Simulate Redis latency (much faster than DB)
		time.Sleep(1 * time.Millisecond)
		return m.data, nil
	}
	return entities.ProductResponse{}, errors.New("miss")
}

func (m *mockCacheStore) SetByID(ctx context.Context, id string, p entities.ProductResponse) error {
	m.data = p
	m.hit = true // Next time it's a hit
	return nil
}

func (m *mockCacheStore) GetList(ctx context.Context, queryHash string) ([]entities.ProductResponse, error) { return nil, nil }
func (m *mockCacheStore) SetList(ctx context.Context, queryHash string, products []entities.ProductResponse) error { return nil }
func (m *mockCacheStore) Invalidate(ctx context.Context, id string) error { return nil }
func (m *mockCacheStore) InvalidateList(ctx context.Context) error { return nil }

// mock dependencies
type mockCategoryRepo struct{}
func (m *mockCategoryRepo) Create(_ context.Context, _ entities.Category) (entities.Category, error) { return entities.Category{}, nil }
func (m *mockCategoryRepo) List(_ context.Context) ([]entities.Category, error) { return nil, nil }
func (m *mockCategoryRepo) GetByID(_ context.Context, _ string) (entities.Category, error) { return entities.Category{}, nil }
func (m *mockCategoryRepo) Update(_ context.Context, _ string, _ map[string]any) (entities.Category, error) { return entities.Category{}, nil }
func (m *mockCategoryRepo) Delete(_ context.Context, _ string, _ map[string]any) error { return nil }
func (m *mockCategoryRepo) ExistsByID(_ context.Context, _ string) (bool, error) { return true, nil }
func (m *mockCategoryRepo) ExistsByName(_ context.Context, _ string, _ *string) (bool, error) { return false, nil }
func (m *mockCategoryRepo) ExistsBySlug(_ context.Context, _ string, _ *string) (bool, error) { return false, nil }
func (m *mockCategoryRepo) ExistsByPosition(_ context.Context, _ int64, _ *string) (bool, error) { return false, nil }
func (m *mockCategoryRepo) EnsureIndexes(_ context.Context) error { return nil }

type mockBrandRepo struct{}
func (m *mockBrandRepo) Create(_ context.Context, _ entities.Brand) (entities.Brand, error) { return entities.Brand{}, nil }
func (m *mockBrandRepo) List(_ context.Context) ([]entities.Brand, error) { return nil, nil }
func (m *mockBrandRepo) GetByID(_ context.Context, _ string) (entities.Brand, error) { return entities.Brand{}, nil }
func (m *mockBrandRepo) Update(_ context.Context, _ string, _ map[string]any) (entities.Brand, error) { return entities.Brand{}, nil }
func (m *mockBrandRepo) Delete(_ context.Context, _ string, _ map[string]any) error { return nil }
func (m *mockBrandRepo) ExistsByID(_ context.Context, _ string) (bool, error) { return true, nil }
func (m *mockBrandRepo) ExistsByName(_ context.Context, _ string, _ *string) (bool, error) { return false, nil }
func (m *mockBrandRepo) ExistsBySlug(_ context.Context, _ string, _ *string) (bool, error) { return false, nil }
func (m *mockBrandRepo) ExistsByPosition(_ context.Context, _ int64, _ *string) (bool, error) { return false, nil }
func (m *mockBrandRepo) ExistsByEmail(_ context.Context, _ string, _ *string) (bool, error) { return false, nil }
func (m *mockBrandRepo) EnsureIndexes(_ context.Context) error { return nil }

func BenchmarkProductGetByID_ColdCache(b *testing.B) {
	svc := service.NewProductService(&mockProductRepo{}, &mockCategoryRepo{}, &mockBrandRepo{}, nil, nil) // No cache

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = svc.GetByID(context.Background(), "p1")
	}
}

func BenchmarkProductGetByID_WarmCache(b *testing.B) {
	cacheStore := &mockCacheStore{
		hit: true,
		data: entities.ProductResponse{ID: "p1", Name: "Test Product"},
	}
	svc := service.NewProductService(&mockProductRepo{}, &mockCategoryRepo{}, &mockBrandRepo{}, cacheStore, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = svc.GetByID(context.Background(), "p1")
	}
}
