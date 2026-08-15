package route

import (
	"emc_lb/src/internal/handler"
	"emc_lb/src/internal/middleware"

	"github.com/gin-gonic/gin"
)

type ProductRoute struct {
	productHandler *handler.ProductHandler
}

func NewProductRoute(productHandler *handler.ProductHandler) *ProductRoute {
	return &ProductRoute{productHandler: productHandler}
}

func (r *ProductRoute) RegisterPublic(router gin.IRouter) {
	productRoute := router.Group("/products")
	{
		productRoute.GET("", r.productHandler.HandleList)
		productRoute.GET("/:id", r.productHandler.HandleGetByID)
	}
}

func (r *ProductRoute) RegisterProtected(router gin.IRouter) {
	productRoute := router.Group("/products")
	{
		productRoute.POST("", middleware.RequirePermission("product:create"), r.productHandler.HandleCreate)
		productRoute.PATCH("/:id", middleware.RequirePermission("product:update"), r.productHandler.HandleUpdate)
		productRoute.DELETE("/:id", middleware.RequirePermission("product:delete"), r.productHandler.HandleDelete)
	}
}
