package route

import (
	"net/http"

	"emc_lb/src/internal/middleware"

	"github.com/gin-gonic/gin"
)

type Route interface {
	RegisterPublic(gin.IRouter)
	RegisterProtected(gin.IRouter)
}

func RegisterRoutes(router *gin.Engine, modules []Route) {
	router.GET("/health", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "ok")
	})

	apiV1 := router.Group("/api/v1")

	publicGroup := apiV1.Group("")
	protectedGroup := apiV1.Group("")
	protectedGroup.Use(middleware.AccessTokenMiddleware())

	for _, moduleRoute := range modules {
		if moduleRoute == nil {
			continue
		}
		moduleRoute.RegisterPublic(publicGroup)
		moduleRoute.RegisterProtected(protectedGroup)
	}
}
