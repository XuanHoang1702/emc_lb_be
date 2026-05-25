package route

import (
	"emc_lb/src/internal/handler"
	"emc_lb/src/internal/middleware"

	"github.com/gin-gonic/gin"
)

type UserRoute struct {
	userHandler *handler.UserHandler
}

func NewUserRoute(userHandler *handler.UserHandler) *UserRoute {
	return &UserRoute{userHandler: userHandler}
}

func (r *UserRoute) Register(router gin.IRouter) {
	userRoute := router.Group("/user")
	{
		userRoute.POST("/register", r.userHandler.HandleRegister)
		userRoute.POST("/verify-email-otp", r.userHandler.HandleVerifyEmailOTP)
		userRoute.POST("/login", r.userHandler.HandleLogin)
		userRoute.POST("/refresh-token", r.userHandler.HandleRefreshToken)
		userRoute.POST("/logout", r.userHandler.HandleLogout)
		userRoute.POST("/delete", r.userHandler.HandleDelete)
		userRoute.PUT("/avatar", middleware.AccessTokenMiddleware(), r.userHandler.HandleUpsertAvatar)
	}
}
