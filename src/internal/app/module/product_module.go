package module

import (
	"emc_lb/src/internal/handler"
	"emc_lb/src/internal/repository"
	route "emc_lb/src/internal/routes"
	"emc_lb/src/internal/service"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

type ProductModule struct {
	routes route.Route
}

func NewProductModule(database *mongo.Database) *ProductModule {
	productRepository := repository.NewProductRepository(database.Collection("products"))
	categoryRepository := repository.NewCategoryRepository(database.Collection("categories"))
	brandRepository := repository.NewBrandRepository(database.Collection("brands"))
	productService := service.NewProductService(productRepository, categoryRepository, brandRepository)
	productHandler := handler.NewProductHandler(productService)
	productRoute := route.NewProductRoute(productHandler)

	return &ProductModule{
		routes: productRoute,
	}
}

func (m *ProductModule) Routes() route.Route {
	return m.routes
}
