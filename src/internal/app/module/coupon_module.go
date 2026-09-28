package module

import (
	"emc_lb/src/internal/handler"
	"emc_lb/src/internal/repository"
	route "emc_lb/src/internal/routes"
	"emc_lb/src/internal/service"
	"emc_lb/src/pkg/cache"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type CouponModule struct {
	Repository repository.CouponRepository
	Service    service.CouponService
	Handler    *handler.CouponHandler
	Route      *route.CouponRoute
}

func NewCouponModule(mongoDB *mongo.Database, redisClient *redis.Client) *CouponModule {
	couponCollection := mongoDB.Collection("coupons")
	repo := repository.NewCouponRepository(couponCollection)
	couponCacheStore := cache.NewRedisCouponCacheStore(redisClient)
	svc := service.NewCouponService(repo, couponCacheStore)
	h := handler.NewCouponHandler(svc)
	r := route.NewCouponRoute(h)

	return &CouponModule{
		Repository: repo,
		Service:    svc,
		Handler:    h,
		Route:      r,
	}
}

func (m *CouponModule) Routes() route.Route {
	return m.Route
}

func (m *CouponModule) ServiceInstance() service.CouponService {
	return m.Service
}
