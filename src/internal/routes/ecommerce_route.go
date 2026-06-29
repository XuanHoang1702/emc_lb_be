package route

import (
	"emc_lb/src/internal/handler"

	"github.com/gin-gonic/gin"
)

type EcommerceRoute struct {
	handler *handler.EcommerceHandler
}

func NewEcommerceRoute(handler *handler.EcommerceHandler) *EcommerceRoute {
	return &EcommerceRoute{handler: handler}
}

func (r *EcommerceRoute) Register(router gin.IRouter) {
	ecommerceGroup := router.Group("/ecommerce")
	{
		ecommerceGroup.GET("/home", r.handler.GetHome)
		ecommerceGroup.GET("/products", r.handler.GetProducts)
		ecommerceGroup.GET("/products/:id", r.handler.GetProductDetail)
		ecommerceGroup.POST("/cart", r.handler.AddToCart)
		ecommerceGroup.POST("/checkout", r.handler.Checkout)
	}
}
