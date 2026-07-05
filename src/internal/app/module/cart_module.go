package module

import (
	"emc_lb/src/internal/handler"
	"emc_lb/src/internal/repository"
	"emc_lb/src/internal/routes"
	"emc_lb/src/internal/service"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

type CartModule struct {
	Repository repository.CartRepository
	Service    service.CartService
	Handler    *handler.CartHandler
	Route      *route.CartRoute
}

func NewCartModule(mongoDB *mongo.Database, productRepo repository.ProductRepository, couponSvc service.CouponService) *CartModule {
	cartCollection := mongoDB.Collection("carts")
	repo := repository.NewCartRepository(cartCollection)
	svc := service.NewCartService(repo, productRepo, couponSvc)
	h := handler.NewCartHandler(svc)
	r := route.NewCartRoute(h)

	return &CartModule{
		Repository: repo,
		Service:    svc,
		Handler:    h,
		Route:      r,
	}
}
