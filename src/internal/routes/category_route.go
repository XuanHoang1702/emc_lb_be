package route

import (
	"emc_lb/src/internal/handler"

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
		categoryRoute.POST("", r.categoryHandler.HandleCreate)
		categoryRoute.GET("", r.categoryHandler.HandleList)
		categoryRoute.GET("/:id", r.categoryHandler.HandleGetByID)
		categoryRoute.PATCH("/:id", r.categoryHandler.HandleUpdate)
		categoryRoute.DELETE("/:id", r.categoryHandler.HandleDelete)
	}
}
