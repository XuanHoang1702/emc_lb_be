package route

import (
	"emc_lb/src/internal/handler"
	"emc_lb/src/internal/middleware"

	"github.com/gin-gonic/gin"
)

type PaymentRoute struct {
	paymentHandler *handler.PaymentHandler
}

func NewPaymentRoute(paymentHandler *handler.PaymentHandler) *PaymentRoute {
	return &PaymentRoute{paymentHandler: paymentHandler}
}

func (r *PaymentRoute) RegisterPublic(router gin.IRouter) {
	paymentGroup := router.Group("/payment")
	{
		// IPN endpoint for SePay Payment Gateway
		paymentGroup.POST("/ipn", middleware.ApiKeyTransacionMiddleware("a"), r.paymentHandler.HandleSepayIPN)

		// Test endpoint: Generates a clickable link/redirect for testing the payment gateway in browser
		paymentGroup.GET("/test-checkout", r.paymentHandler.HandleTestCheckoutLink)
	}
}

func (r *PaymentRoute) RegisterProtected(router gin.IRouter) {
	paymentGroup := router.Group("/payment")
	{
		// Endpoint for initializing Checkout (used by Frontend)
		paymentGroup.POST("/checkout", r.paymentHandler.HandleInitCheckout)
	}
}
