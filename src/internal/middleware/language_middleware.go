package middleware

import (
	"github.com/gin-gonic/gin"
)

func AcceptLanguageMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if ctx.GetHeader("Accept-Language") == "" {
			ctx.Request.Header.Set("Accept-Language", "en")
		}
		ctx.Next()
	}
}
