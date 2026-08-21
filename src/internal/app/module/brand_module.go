package module

import (
	"context"

	"emc_lb/src/internal/handler"
	"emc_lb/src/internal/repository"
	route "emc_lb/src/internal/routes"
	"emc_lb/src/internal/service"
	"emc_lb/src/pkg/cache"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type BrandModule struct {
	routes route.Route
}

func NewBrandModule(database *mongo.Database, redisClient *redis.Client) (*BrandModule, error) {
	brandRepository := repository.NewBrandRepository(database.Collection("brands"))
	if err := brandRepository.EnsureIndexes(context.Background()); err != nil {
		return nil, err
	}
	brandCacheStore := cache.NewRedisBrandCacheStore(redisClient)
	brandService := service.NewBrandService(brandRepository, brandCacheStore)
	brandHandler := handler.NewBrandHandler(brandService)
	brandRoute := route.NewBrandRoute(brandHandler)

	return &BrandModule{
		routes: brandRoute,
	}, nil
}

func (m *BrandModule) Routes() route.Route {
	return m.routes
}
