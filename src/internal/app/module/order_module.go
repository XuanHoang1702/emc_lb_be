package module

import (
	"emc_lb/src/internal/handler"
	"emc_lb/src/internal/repository"
	route "emc_lb/src/internal/routes"
	"emc_lb/src/internal/service"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type OrderModule struct {
	routes       route.Route
	orderService service.OrderService
}

func NewOrderModule(database *mongo.Database, couponSvc service.CouponService, redisClient *redis.Client) *OrderModule {
	orderRepository := repository.NewOrderRepository(database.Collection("orders"))
	productRepository := repository.NewProductRepository(database.Collection("products"))
	
	orderService := service.NewOrderService(orderRepository, productRepository, couponSvc, redisClient)
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
