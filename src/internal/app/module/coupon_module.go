package module

import (
	"emc_lb/src/internal/handler"
	"emc_lb/src/internal/repository"
	"emc_lb/src/internal/routes"
	"emc_lb/src/internal/service"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

type CouponModule struct {
	Repository repository.CouponRepository
	Service    service.CouponService
	Handler    *handler.CouponHandler
	Route      *route.CouponRoute
}

func NewCouponModule(mongoDB *mongo.Database) *CouponModule {
	couponCollection := mongoDB.Collection("coupons")
	repo := repository.NewCouponRepository(couponCollection)
	svc := service.NewCouponService(repo)
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
