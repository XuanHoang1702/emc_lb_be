package middleware

import (
	"crypto/subtle"
	"net/http"

	"github.com/gin-gonic/gin"
)

func APIKeyMiddleware(expectedKey string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		apiKey := ctx.GetHeader("X-API-Key")
		if apiKey == "" {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Missing X-API-Key"})
			return
		}

		if subtle.ConstantTimeCompare([]byte(apiKey), []byte(expectedKey)) != 1 {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid API Key"})
			return
		}

		ctx.Next()
	}
}

func APIKeyTransacionMiddleware(expectedKey string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		apiKey := ctx.GetHeader("X-API-KEY-Transaction")
		if apiKey == "" {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Missing X-API-KEY-Transaction"})
			return
		}
		if subtle.ConstantTimeCompare([]byte(apiKey), []byte(expectedKey)) != 1 {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid API KEY Transaction"})
			return
		}
		ctx.Next()
	}
}
