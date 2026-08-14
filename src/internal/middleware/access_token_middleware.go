package middleware

import (
	"net/http"
	"strings"

	"emc_lb/src/pkg/errors"
	"emc_lb/src/pkg/res"
	"emc_lb/src/pkg/utils"

	"github.com/gin-gonic/gin"
)

const ContextUserIDKey = "user_id"
const ContextRoleKey = "user_role"

func AccessTokenMiddleware(accessSecret string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authorizationHeader := strings.TrimSpace(ctx.GetHeader("Authorization"))
		if authorizationHeader == "" {
			res.Error(ctx, &res.AppError{
				Message:    "Missing authorization header",
				Code:       errors.UserUnauthorized,
				StatusCode: http.StatusUnauthorized,
			})
			ctx.Abort()
			return
		}

		tokenParts := strings.SplitN(authorizationHeader, " ", 2)
		if len(tokenParts) != 2 || !strings.EqualFold(tokenParts[0], "Bearer") || strings.TrimSpace(tokenParts[1]) == "" {
			res.Error(ctx, &res.AppError{
				Message:    "Invalid authorization header",
				Code:       errors.UserUnauthorized,
				StatusCode: http.StatusUnauthorized,
			})
			ctx.Abort()
			return
		}

		userID, role, err := utils.ParseAccessToken(strings.TrimSpace(tokenParts[1]), accessSecret)
		if err != nil {
			res.Error(ctx, &res.AppError{
				Message:    "Invalid access token",
				Code:       errors.UserUnauthorized,
				StatusCode: http.StatusUnauthorized,
			})
			ctx.Abort()
			return
		}

		ctx.Set(ContextUserIDKey, userID)
		ctx.Set(ContextRoleKey, role)
		ctx.Next()
	}
}
