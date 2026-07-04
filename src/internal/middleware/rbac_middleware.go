package middleware

import (
	"net/http"

	"emc_lb/src/pkg/auth"
	"emc_lb/src/pkg/errors"
	"emc_lb/src/pkg/res"

	"github.com/gin-gonic/gin"
)

func RequirePermission(permissionCode string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		role, exists := ctx.Get(ContextRoleKey)
		if !exists {
			res.Error(ctx, &res.AppError{
				Message:    "Unauthorized: Role not found",
				Code:       errors.UserUnauthorized,
				StatusCode: http.StatusUnauthorized,
			})
			ctx.Abort()
			return
		}

		roleStr, ok := role.(string)
		if !ok || roleStr == "" {
			res.Error(ctx, &res.AppError{
				Message:    "Unauthorized: Invalid role",
				Code:       errors.UserUnauthorized,
				StatusCode: http.StatusUnauthorized,
			})
			ctx.Abort()
			return
		}

		rbacManager := auth.GetRBACManager()
		if !rbacManager.HasPermission(roleStr, permissionCode) {
			res.Error(ctx, &res.AppError{
				Message:    "Forbidden: You do not have permission to access this resource",
				Code:       "FORBIDDEN",
				StatusCode: http.StatusForbidden,
			})
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}
