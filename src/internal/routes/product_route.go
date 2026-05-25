package route

import (
	"emc_lb/src/internal/handler"

	"github.com/gin-gonic/gin"
)

type ProductRoute struct {
	productHandler *handler.ProductHandler
}

func NewProductRoute(productHandler *handler.ProductHandler) *ProductRoute {
	return &ProductRoute{productHandler: productHandler}
}

func (r *ProductRoute) Register(router gin.IRouter) {
	productRoute := router.Group("/products")
	{
		productRoute.POST("", r.productHandler.HandleCreate)
		productRoute.GET("", r.productHandler.HandleList)
	}
}
