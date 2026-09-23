package module

import (
	"context"
	"log"

	"emc_lb/src/internal/handler"
	"emc_lb/src/internal/repository"
	route "emc_lb/src/internal/routes"
	"emc_lb/src/internal/service"
	"emc_lb/src/pkg/cache"
	"emc_lb/src/pkg/worker"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type OrderModule struct {
	routes       route.Route
	orderService service.OrderService
}

func NewOrderModule(database *mongo.Database, mongoClient *mongo.Client, couponSvc service.CouponService, redisClient *redis.Client, productCache cache.ProductCacheStore, taskDistributor worker.TaskDistributor) *OrderModule {
	orderRepository := repository.NewOrderRepository(database.Collection("orders"))
	productRepository := repository.NewProductRepository(database.Collection("products"))

	// Ensure indexes for optimized lookups
	if err := orderRepository.EnsureIndexes(context.Background()); err != nil {
		log.Printf("warning: failed to ensure order indexes: %v", err)
	}

	inventoryService := service.NewInventoryService(redisClient)
	orderService := service.NewOrderService(orderRepository, productRepository, couponSvc, inventoryService, productCache, mongoClient, taskDistributor)
	orderHandler := handler.NewOrderHandler(orderService)
	orderRoute := route.NewOrderRoute(orderHandler)

	return &OrderModule{
		routes:       orderRoute,
		orderService: orderService,
	}
}

func (m *OrderModule) Routes() route.Route {
	return m.routes
}

func (m *OrderModule) Service() service.OrderService {
	return m.orderService
}
