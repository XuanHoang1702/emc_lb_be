package route

import (
	"emc_lb/src/internal/handler"

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
		brandRoute.POST("", r.brandHandler.HandleCreate)
		brandRoute.GET("", r.brandHandler.HandleList)
		brandRoute.GET("/:id", r.brandHandler.HandleGetByID)
		brandRoute.PATCH("/:id", r.brandHandler.HandleUpdate)
		brandRoute.DELETE("/:id", r.brandHandler.HandleDelete)
	}
}
