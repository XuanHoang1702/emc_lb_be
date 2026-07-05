package route

import (
	"emc_lb/src/internal/handler"
	"emc_lb/src/internal/middleware"

	"github.com/gin-gonic/gin"
)

type CategoryRoute struct {
	categoryHandler *handler.CategoryHandler
}

func NewCategoryRoute(categoryHandler *handler.CategoryHandler) *CategoryRoute {
	return &CategoryRoute{categoryHandler: categoryHandler}
}

func (r *CategoryRoute) RegisterPublic(router gin.IRouter) {
	categoryRoute := router.Group("/categories")
	{
		categoryRoute.GET("", r.categoryHandler.HandleList)
		categoryRoute.GET("/:id", r.categoryHandler.HandleGetByID)
	}
}

func (r *CategoryRoute) RegisterProtected(router gin.IRouter) {
	categoryRoute := router.Group("/categories")
	{
		categoryRoute.POST("", middleware.RequirePermission("category:create"), r.categoryHandler.HandleCreate)
		categoryRoute.PATCH("/:id", middleware.RequirePermission("category:update"), r.categoryHandler.HandleUpdate)
		categoryRoute.DELETE("/:id", middleware.RequirePermission("category:delete"), r.categoryHandler.HandleDelete)
	}
}
