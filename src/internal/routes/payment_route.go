package route

import (
	"time"

	"emc_lb/src/internal/handler"
	"emc_lb/src/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type PaymentRoute struct {
	paymentHandler *handler.PaymentHandler
	redisClient    *redis.Client
}

func NewPaymentRoute(paymentHandler *handler.PaymentHandler) *PaymentRoute {
	return &PaymentRoute{paymentHandler: paymentHandler}
}

func (r *PaymentRoute) SetRedisClient(client *redis.Client) {
	r.redisClient = client
}

func (r *PaymentRoute) RegisterPublic(router gin.IRouter) {
	paymentGroup := router.Group("/payment")

	if r.redisClient != nil {
		// IPN endpoint: strict rate limit (30 requests per minute per IP)
		// SePay should not send more than this under normal conditions.
		ipnGroup := paymentGroup.Group("")
		ipnGroup.Use(middleware.AuthRateLimitMiddleware(r.redisClient, 30, time.Minute))
		{
			ipnGroup.POST("/ipn", r.paymentHandler.HandleSepayIPN)
		}
	} else {
		paymentGroup.POST("/ipn", r.paymentHandler.HandleSepayIPN)
	}

	// Test endpoint: Generates a clickable link/redirect for testing the payment gateway in browser
	paymentGroup.GET("/test-checkout", r.paymentHandler.HandleTestCheckoutLink)

	// Callback endpoints for SePay redirection after payment
	paymentGroup.GET("/success", r.paymentHandler.HandlePaymentSuccess)
	paymentGroup.GET("/error", r.paymentHandler.HandlePaymentError)
	paymentGroup.GET("/cancel", r.paymentHandler.HandlePaymentCancel)
}

func (r *PaymentRoute) RegisterProtected(router gin.IRouter) {
	paymentGroup := router.Group("/payment")
	{
		// Endpoint for initializing Checkout (used by Frontend)
		paymentGroup.POST("/checkout", r.paymentHandler.HandleInitCheckout)
	}
}
