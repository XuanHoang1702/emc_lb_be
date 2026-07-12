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

type CategoryModule struct {
	routes route.Route
}

func NewCategoryModule(database *mongo.Database, redisClient *redis.Client) (*CategoryModule, error) {
	categoryRepository := repository.NewCategoryRepository(database.Collection("categories"))
	if err := categoryRepository.EnsureIndexes(context.Background()); err != nil {
		return nil, err
	}
	categoryCacheStore := cache.NewRedisCategoryCacheStore(redisClient)
	categoryService := service.NewCategoryService(categoryRepository, categoryCacheStore)
	categoryHandler := handler.NewCategoryHandler(categoryService)
	categoryRoute := route.NewCategoryRoute(categoryHandler)

	return &CategoryModule{
		routes: categoryRoute,
	}, nil
}

func (m *CategoryModule) Routes() route.Route {
	return m.routes
}

