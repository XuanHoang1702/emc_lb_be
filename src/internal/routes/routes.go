package route

import (
	"net/http"
	"time"

	"emc_lb/src/internal/middleware"
	"emc_lb/src/pkg/config"

	_ "emc_lb/src/docs"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type Route interface {
	RegisterPublic(gin.IRouter)
	RegisterProtected(gin.IRouter)
}

// AuthRateLimitedRoute is an optional interface for routes that need
// stricter per-endpoint auth rate limiting (login, register, OTP, etc.).
type AuthRateLimitedRoute interface {
	Route
	SetRedisClient(client *redis.Client)
}

func RegisterRoutes(router *gin.Engine, modules []Route, redisClient *redis.Client, pgPool *pgxpool.Pool, mongoClient *mongo.Client, cfg *config.AppConfig) {
	// Prometheus Metrics Layer
	router.Use(middleware.PrometheusMiddleware())
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Public health check: simple up/down for load balancer probes
	router.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"status": "up",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	// Internal health check: detailed service status, protected by API key
	router.GET("/health/detail", middleware.APIKeyMiddleware(cfg.App.APIKey), func(ctx *gin.Context) {
		reqCtx := ctx.Request.Context()
		status := http.StatusOK
		services := gin.H{}

		// Postgres
		if pgPool != nil {
			start := time.Now()
			if err := pgPool.Ping(reqCtx); err != nil {
				services["postgres"] = gin.H{"status": "down", "error": err.Error()}
				status = http.StatusServiceUnavailable
			} else {
				stat := pgPool.Stat()
				services["postgres"] = gin.H{
					"status":     "up",
					"latency_ms": time.Since(start).Milliseconds(),
					"pool": gin.H{
						"total":  stat.TotalConns(),
						"idle":   stat.IdleConns(),
						"in_use": stat.AcquiredConns(),
					},
				}
			}
		} else {
			services["postgres"] = gin.H{"status": "not_configured"}
		}

		// MongoDB
		if mongoClient != nil {
			start := time.Now()
			if err := mongoClient.Ping(reqCtx, nil); err != nil {
				services["mongodb"] = gin.H{"status": "down", "error": err.Error()}
				status = http.StatusServiceUnavailable
			} else {
				services["mongodb"] = gin.H{
					"status":     "up",
					"latency_ms": time.Since(start).Milliseconds(),
				}
			}
		} else {
			services["mongodb"] = gin.H{"status": "not_configured"}
		}

		// Redis
		if redisClient != nil {
			start := time.Now()
			if err := redisClient.Ping(reqCtx).Err(); err != nil {
				services["redis"] = gin.H{"status": "down", "error": err.Error()}
				status = http.StatusServiceUnavailable
			} else {
				services["redis"] = gin.H{
					"status":     "up",
					"latency_ms": time.Since(start).Milliseconds(),
				}
			}
		} else {
			services["redis"] = gin.H{"status": "not_configured"}
		}

		overallStatus := "up"
		if status != http.StatusOK {
			overallStatus = "degraded"
		}

		ctx.JSON(status, gin.H{
			"status":   overallStatus,
			"time":     time.Now().Format(time.RFC3339),
			"services": services,
		})
	})

	// Swagger docs: protected by API key in production
	docsGroup := router.Group("/docs")
	if gin.Mode() == gin.ReleaseMode {
		docsGroup.Use(middleware.APIKeyMiddleware(cfg.App.APIKey))
	}
	docsGroup.GET("/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	apiV1 := router.Group("/api/v1")

	publicGroup := apiV1.Group("")
	// Apply rate limiting: max 100 requests per 10 seconds per IP
	if redisClient != nil {
		publicGroup.Use(middleware.RateLimitMiddleware(redisClient, 100, 10*time.Second))
	}

	protectedGroup := apiV1.Group("")
	protectedGroup.Use(middleware.AccessTokenMiddleware(cfg.JWT.AccessSecret))
	// Apply rate limiting for authenticated users: max 200 requests per 10 seconds per user
	if redisClient != nil {
		protectedGroup.Use(middleware.RateLimitByUserMiddleware(redisClient, 200, 10*time.Second))
	}

	// Inject redis client into routes that implement AuthRateLimitedRoute
	for _, moduleRoute := range modules {
		if moduleRoute == nil {
			continue
		}
		if authRoute, ok := moduleRoute.(AuthRateLimitedRoute); ok && redisClient != nil {
			authRoute.SetRedisClient(redisClient)
		}
		moduleRoute.RegisterPublic(publicGroup)
		moduleRoute.RegisterProtected(protectedGroup)
	}
}

