package route

import (
	"emc_lb/src/internal/handler"

	"github.com/gin-gonic/gin"
)

type CartRoute struct {
	cartHandler *handler.CartHandler
}

func NewCartRoute(cartHandler *handler.CartHandler) *CartRoute {
	return &CartRoute{cartHandler: cartHandler}
}

func (r *CartRoute) RegisterPublic(router gin.IRouter) {
	// Cart requires user context, no public routes
}

func (r *CartRoute) RegisterProtected(router gin.IRouter) {
	cartGroup := router.Group("/cart")
	{
		cartGroup.GET("", r.cartHandler.HandleGetCart)
		cartGroup.POST("/items", r.cartHandler.HandleAddItem)
		cartGroup.PUT("/items/:productId", r.cartHandler.HandleUpdateItem)
		cartGroup.DELETE("/items/:productId", r.cartHandler.HandleRemoveItem)
		cartGroup.DELETE("", r.cartHandler.HandleClearCart)
		cartGroup.POST("/coupon", r.cartHandler.HandleApplyCoupon)
	}
}
