package route

import (
	"time"

	"emc_lb/src/internal/handler"
	"emc_lb/src/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type UserRoute struct {
	userHandler *handler.UserHandler
	redisClient *redis.Client
}

func NewUserRoute(userHandler *handler.UserHandler) *UserRoute {
	return &UserRoute{userHandler: userHandler}
}

func (r *UserRoute) SetRedisClient(client *redis.Client) {
	r.redisClient = client
}

func (r *UserRoute) RegisterPublic(router gin.IRouter) {
	userRoute := router.Group("/user")

	// Apply strict auth rate limiting for sensitive endpoints:
	// - Login/Register: 10 requests per minute per IP (prevents brute-force)
	// - OTP verification: 5 requests per minute per IP (prevents OTP guessing)
	// - Refresh token: 20 requests per minute per IP
	if r.redisClient != nil {
		loginGroup := userRoute.Group("")
		loginGroup.Use(middleware.AuthRateLimitMiddleware(r.redisClient, 10, time.Minute, middleware.PolicyFailClosed))
		{
			loginGroup.POST("/register", r.userHandler.HandleRegister)
			loginGroup.POST("/login", r.userHandler.HandleLogin)
		}

		otpGroup := userRoute.Group("")
		otpGroup.Use(middleware.AuthRateLimitMiddleware(r.redisClient, 5, time.Minute, middleware.PolicyFailClosed))
		{
			otpGroup.POST("/verify-email-otp", r.userHandler.HandleVerifyEmailOTP)
			otpGroup.POST("/forgot-password", r.userHandler.HandleForgotPassword)
			otpGroup.POST("/reset-password", r.userHandler.HandleResetPassword)
		}

		refreshGroup := userRoute.Group("")
		refreshGroup.Use(middleware.AuthRateLimitMiddleware(r.redisClient, 20, time.Minute, middleware.PolicyFailClosed))
		{
			refreshGroup.POST("/refresh-token", r.userHandler.HandleRefreshToken)
		}
	} else {
		// Fallback: no auth rate limiting if Redis is not available
		userRoute.POST("/register", r.userHandler.HandleRegister)
		userRoute.POST("/verify-email-otp", r.userHandler.HandleVerifyEmailOTP)
		userRoute.POST("/login", r.userHandler.HandleLogin)
		userRoute.POST("/refresh-token", r.userHandler.HandleRefreshToken)
		userRoute.POST("/forgot-password", r.userHandler.HandleForgotPassword)
		userRoute.POST("/reset-password", r.userHandler.HandleResetPassword)
	}
}

func (r *UserRoute) RegisterProtected(router gin.IRouter) {
	userRoute := router.Group("/user")
	{
		userRoute.GET("/me", r.userHandler.HandleGetProfile)
		userRoute.POST("/logout", r.userHandler.HandleLogout)
		userRoute.POST("/delete", r.userHandler.HandleDelete)
		userRoute.PUT("/avatar", r.userHandler.HandleUpsertAvatar)
		userRoute.POST("/change-password", r.userHandler.HandleChangePassword)

		adminRoute := userRoute.Group("/admin")
		adminRoute.Use(middleware.RequirePermission("user:update"))
		{
			adminRoute.POST("/change-password/:uuid", r.userHandler.HandleAdminChangePassword)
		}
	}
}
