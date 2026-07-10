package route

import (
	"net/http"
	"time"

	"emc_lb/src/internal/middleware"

	_ "emc_lb/src/docs"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type Route interface {
	RegisterPublic(gin.IRouter)
	RegisterProtected(gin.IRouter)
}

func RegisterRoutes(router *gin.Engine, modules []Route, redisClient *redis.Client, pgPool *pgxpool.Pool, mongoClient *mongo.Client) {
	router.GET("/health", func(ctx *gin.Context) {
		status := http.StatusOK
		response := gin.H{
			"status": "up",
			"time":   time.Now().Format(time.RFC3339),
			"services": gin.H{
				"postgres": "up",
				"mongodb":  "up",
				"redis":    "up",
			},
		}

		// Ping Postgres
		if pgPool != nil {
			if err := pgPool.Ping(ctx.Request.Context()); err != nil {
				response["services"].(gin.H)["postgres"] = "down"
				status = http.StatusServiceUnavailable
			}
		} else {
			response["services"].(gin.H)["postgres"] = "not_configured"
		}

		// Ping MongoDB
		if mongoClient != nil {
			if err := mongoClient.Ping(ctx.Request.Context(), nil); err != nil {
				response["services"].(gin.H)["mongodb"] = "down"
				status = http.StatusServiceUnavailable
			}
		} else {
			response["services"].(gin.H)["mongodb"] = "not_configured"
		}

		// Ping Redis
		if redisClient != nil {
			if err := redisClient.Ping(ctx.Request.Context()).Err(); err != nil {
				response["services"].(gin.H)["redis"] = "down"
				status = http.StatusServiceUnavailable
			}
		} else {
			response["services"].(gin.H)["redis"] = "not_configured"
		}

		if status != http.StatusOK {
			response["status"] = "down"
		}

		ctx.JSON(status, response)
	})

	router.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	apiV1 := router.Group("/api/v1")

	publicGroup := apiV1.Group("")
	// Apply rate limiting: max 100 requests per 10 seconds per IP
	if redisClient != nil {
		publicGroup.Use(middleware.RateLimitMiddleware(redisClient, 100, 10*time.Second))
	}

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
