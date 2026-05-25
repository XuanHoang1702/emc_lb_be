package route

import (
	"net/http"

	"emc_lb/src/internal/middleware"

	"github.com/gin-gonic/gin"
)

type Route interface {
	Register(gin.IRouter)
}

type RouteGroups struct {
	Public      []Route
	Protected   []Route
	Transaction []Route
}

func RegisterRoutes(router *gin.Engine, groups RouteGroups) {
	router.GET("/health", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "ok")
	})

	apiV1 := router.Group("/api/v1")

	publicGroup := apiV1.Group("")
	registerGroup(publicGroup, groups.Public)

	protectedGroup := apiV1.Group("")
	protectedGroup.Use(middleware.ApiKeyMiddleware())
	registerGroup(protectedGroup, groups.Protected)

	transactionGroup := apiV1.Group("/transaction")
	transactionGroup.Use(middleware.ApiKeyTransacionMiddleware())
	registerGroup(transactionGroup, groups.Transaction)
}

func registerGroup(router gin.IRouter, routes []Route) {
	for _, currentRoute := range routes {
		if currentRoute == nil {
			continue
		}

		currentRoute.Register(router)
	}
}
