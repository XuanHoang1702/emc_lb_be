package module

import (
	"emc_lb/src/internal/handler"
	"emc_lb/src/internal/repository"
	route "emc_lb/src/internal/routes"
	"emc_lb/src/internal/service"
	"emc_lb/src/pkg/cache"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"github.com/meilisearch/meilisearch-go"
)

type ProductModule struct {
	routes     route.Route
	repository repository.ProductRepository
	cacheStore cache.ProductCacheStore
}

func NewProductModule(database *mongo.Database, redisClient *redis.Client, searchClient meilisearch.ServiceManager) *ProductModule {
	productRepository := repository.NewProductRepository(database.Collection("products"))
	categoryRepository := repository.NewCategoryRepository(database.Collection("categories"))
	brandRepository := repository.NewBrandRepository(database.Collection("brands"))
	productCacheStore := cache.NewRedisProductCacheStore(redisClient)
	productSearchRepository := repository.NewProductSearchRepository(searchClient)
	productService := service.NewProductService(productRepository, categoryRepository, brandRepository, productCacheStore, productSearchRepository)
	productHandler := handler.NewProductHandler(productService)
	productRoute := route.NewProductRoute(productHandler)

	return &ProductModule{
		routes:     productRoute,
		repository: productRepository,
		cacheStore: productCacheStore,
	}
}

func (m *ProductModule) Repository() repository.ProductRepository {
	return m.repository
}

func (m *ProductModule) CacheStore() cache.ProductCacheStore {
	return m.cacheStore
}

func (m *ProductModule) Routes() route.Route {
	return m.routes
}
