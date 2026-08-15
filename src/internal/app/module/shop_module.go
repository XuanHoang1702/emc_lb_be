package module

import (
	"emc_lb/src/internal/handler"
	"emc_lb/src/internal/repository"
	route "emc_lb/src/internal/routes"
	"emc_lb/src/internal/service"
	"emc_lb/src/pkg/cache"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type ShopModule struct {
	repo    repository.ShopRepository
	service service.ShopService
	handler handler.ShopHandler
}

func NewShopModule(db *mongo.Database, redisClient *redis.Client) *ShopModule {
	repo := repository.NewShopRepository(db.Collection("shops"))
	shopCacheStore := cache.NewRedisShopCacheStore(redisClient)
	svc := service.NewShopService(repo, shopCacheStore)
	hdl := handler.NewShopHandler(svc)

	return &ShopModule{
		repo:    repo,
		service: svc,
		handler: hdl,
	}
}

func (m *ShopModule) Routes() route.Route {
	return m
}

func (m *ShopModule) RegisterPublic(router gin.IRouter) {
	// Public routes for shop if any (e.g. get shop details by slug)
}

func (m *ShopModule) RegisterProtected(router gin.IRouter) {
	shopGroup := router.Group("/shops")
	{
		shopGroup.POST("", m.handler.CreateShop)
		shopGroup.GET("/my-shop", m.handler.GetMyShop)
	}
}
