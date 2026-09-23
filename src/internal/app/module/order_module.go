package module

import (

	"emc_lb/src/internal/handler"
	"emc_lb/src/internal/repository"
	route "emc_lb/src/internal/routes"
	"emc_lb/src/internal/service"
	"emc_lb/src/pkg/cache"
	"emc_lb/src/pkg/worker"

	"github.com/jackc/pgx/v5/pgxpool"
	"emc_lb/src/internal/db/sqlc"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type OrderModule struct {
	routes       route.Route
	orderService service.OrderService
}

func NewOrderModule(database *mongo.Database, pgxpool *pgxpool.Pool, sqlcQuerier sqlc.Querier, couponSvc service.CouponService, redisClient *redis.Client, productCache cache.ProductCacheStore, taskDistributor worker.TaskDistributor) *OrderModule {
	orderRepository := repository.NewOrderRepository(pgxpool, sqlcQuerier)
	productRepository := repository.NewProductRepository(database.Collection("products"))


	inventoryService := service.NewInventoryService(redisClient)
	cartRepository := repository.NewCartRepository(database.Collection("carts"))
	orderService := service.NewOrderService(orderRepository, productRepository, cartRepository, couponSvc, inventoryService, productCache, pgxpool, taskDistributor)
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
