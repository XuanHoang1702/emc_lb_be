package route

import (
	"emc_lb/src/internal/handler"
	"emc_lb/src/internal/middleware"

	"github.com/gin-gonic/gin"
)

type OrderRoute struct {
	orderHandler *handler.OrderHandler
}

func NewOrderRoute(orderHandler *handler.OrderHandler) *OrderRoute {
	return &OrderRoute{orderHandler: orderHandler}
}

func (r *OrderRoute) RegisterPublic(router gin.IRouter) {
	// Orders don't have public routes (must be logged in to order)
}

func (r *OrderRoute) RegisterProtected(router gin.IRouter) {
	orderGroup := router.Group("/orders")
	{
		orderGroup.POST("", r.orderHandler.HandleCreate)
		orderGroup.GET("/my", r.orderHandler.HandleMyOrders)
		orderGroup.GET("/:id", r.orderHandler.HandleGetOrder)
		
		// Admin only
		orderGroup.GET("", middleware.RequirePermission("manage_orders"), r.orderHandler.HandleGetAllOrders)
		orderGroup.PATCH("/:id/status", middleware.RequirePermission("manage_orders"), r.orderHandler.HandleUpdateStatus)
	}
}
