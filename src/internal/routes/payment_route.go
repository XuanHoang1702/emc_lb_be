package route

import (
	"emc_lb/src/internal/handler"

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
		// Webhook endpoint for SePay IPN
		paymentGroup.POST("/webhook", r.paymentHandler.HandleSepayWebhook)

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
