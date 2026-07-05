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

func (r *BrandRoute) RegisterPublic(router gin.IRouter) {
	brandRoute := router.Group("/brands")
	{
		brandRoute.GET("", r.brandHandler.HandleList)
		brandRoute.GET("/:id", r.brandHandler.HandleGetByID)
	}
}

func (r *BrandRoute) RegisterProtected(router gin.IRouter) {
	brandRoute := router.Group("/brands")
	{
		brandRoute.POST("", middleware.RequirePermission("brand:create"), r.brandHandler.HandleCreate)
		brandRoute.PATCH("/:id", middleware.RequirePermission("brand:update"), r.brandHandler.HandleUpdate)
		brandRoute.DELETE("/:id", middleware.RequirePermission("brand:delete"), r.brandHandler.HandleDelete)
	}
}
