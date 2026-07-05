package route

import (
	"emc_lb/src/internal/handler"
	"emc_lb/src/internal/middleware"

	"github.com/gin-gonic/gin"
)

type CouponRoute struct {
	couponHandler *handler.CouponHandler
}

func NewCouponRoute(couponHandler *handler.CouponHandler) *CouponRoute {
	return &CouponRoute{couponHandler: couponHandler}
}

func (r *CouponRoute) RegisterPublic(router gin.IRouter) {
	// Public routes
}

func (r *CouponRoute) RegisterProtected(router gin.IRouter) {
	couponGroup := router.Group("/coupons")
	{
		couponGroup.POST("", middleware.RequirePermission("coupon:create"), r.couponHandler.HandleCreate)
		couponGroup.GET("", middleware.RequirePermission("coupon:read"), r.couponHandler.HandleList)
		couponGroup.GET("/:code", middleware.RequirePermission("coupon:read"), r.couponHandler.HandleGet)
		couponGroup.PATCH("/:id", middleware.RequirePermission("coupon:update"), r.couponHandler.HandleUpdate)
	}
}
