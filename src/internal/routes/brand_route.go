package route

import (
	"emc_lb/src/internal/handler"
	"emc_lb/src/internal/middleware"

	"github.com/gin-gonic/gin"
)

type BrandRoute struct {
	brandHandler *handler.BrandHandler
}

func NewBrandRoute(brandHandler *handler.BrandHandler) *BrandRoute {
	return &BrandRoute{brandHandler: brandHandler}
}

func (r *BrandRoute) Register(router gin.IRouter) {
	brandRoute := router.Group("/brands")
	{
		brandRoute.POST("", middleware.AccessTokenMiddleware(), middleware.RequirePermission("brand:create"), r.brandHandler.HandleCreate)
		brandRoute.GET("", r.brandHandler.HandleList)
		brandRoute.GET("/:id", r.brandHandler.HandleGetByID)
		brandRoute.PATCH("/:id", middleware.AccessTokenMiddleware(), middleware.RequirePermission("brand:update"), r.brandHandler.HandleUpdate)
		brandRoute.DELETE("/:id", middleware.AccessTokenMiddleware(), middleware.RequirePermission("brand:delete"), r.brandHandler.HandleDelete)
	}
}
