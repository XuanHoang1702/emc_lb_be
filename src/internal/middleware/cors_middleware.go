package middleware

import (
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func CORSMiddleware() gin.HandlerFunc {
	config := cors.DefaultConfig()

	// Read allowed origins from env, default to localhost for dev safety.
	// In production, set CORS_ALLOWED_ORIGINS=https://your-frontend.com,https://admin.your-frontend.com
	allowedOrigins := os.Getenv("CORS_ALLOWED_ORIGINS")
	if allowedOrigins != "" {
		origins := strings.Split(allowedOrigins, ",")
		for i := range origins {
			origins[i] = strings.TrimSpace(origins[i])
		}
		config.AllowOrigins = origins
	} else {
		config.AllowOrigins = []string{
			"http://localhost:3000",
			"http://localhost:5173",
		}
	}

	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{
		"Origin",
		"Content-Length",
		"Content-Type",
		"Authorization",
		"Accept-Language", // used for i18n
	}
	config.ExposeHeaders = []string{"Content-Length"}
	config.AllowCredentials = true
	config.MaxAge = 12 * time.Hour

	return cors.New(config)
}
