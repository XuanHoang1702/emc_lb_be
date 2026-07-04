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

func (r *CategoryRoute) Register(router gin.IRouter) {
	categoryRoute := router.Group("/categories")
	{
		categoryRoute.POST("", middleware.AccessTokenMiddleware(), middleware.RequirePermission("category:create"), r.categoryHandler.HandleCreate)
		categoryRoute.GET("", r.categoryHandler.HandleList)
		categoryRoute.GET("/:id", r.categoryHandler.HandleGetByID)
		categoryRoute.PATCH("/:id", middleware.AccessTokenMiddleware(), middleware.RequirePermission("category:update"), r.categoryHandler.HandleUpdate)
		categoryRoute.DELETE("/:id", middleware.AccessTokenMiddleware(), middleware.RequirePermission("category:delete"), r.categoryHandler.HandleDelete)
	}
}
