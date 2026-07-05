package handler

import (
	"net/http"

	"emc_lb/src/internal/service"
	"emc_lb/src/pkg/entities"
	"emc_lb/src/pkg/errors"
	"emc_lb/src/pkg/res"
	"emc_lb/src/pkg/validation"

	"github.com/gin-gonic/gin"
)

type PaymentHandler struct {
	paymentService service.PaymentService
}

func NewPaymentHandler(paymentService service.PaymentService) *PaymentHandler {
	return &PaymentHandler{paymentService: paymentService}
}

// HandleInitCheckout initializes a SePay checkout session
func (h *PaymentHandler) HandleInitCheckout(ctx *gin.Context) {
	var req entities.CheckoutInitRequest
	if err := validation.BindJSON(ctx, &req, errors.UserInvalidFormat); err != nil {
		res.Error(ctx, err)
		return
	}

	data, err := h.paymentService.InitCheckout(ctx.Request.Context(), req)
	if err != nil {
		res.Error(ctx, err)
		return
	}

	res.Success(ctx, http.StatusOK, data)
}

// HandleSepayWebhook handles webhook events from SePay
func (h *PaymentHandler) HandleSepayWebhook(ctx *gin.Context) {
	var req entities.SePayWebhookRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		res.Error(ctx, &res.AppError{
			Code:       errors.UserInvalidFormat,
			Message:    "Invalid webhook payload format",
			StatusCode: http.StatusBadRequest,
		})
		return
	}

	authHeader := ctx.GetHeader("Authorization")
	if err := h.paymentService.ProcessWebhook(ctx.Request.Context(), req, authHeader); err != nil {
		res.Error(ctx, err)
		return
	}

	res.Success(ctx, http.StatusOK, "Webhook received successfully")
}

// HandleTestCheckoutLink renders an auto-submitting HTML form for testing checkout via a simple GET link
func (h *PaymentHandler) HandleTestCheckoutLink(ctx *gin.Context) {
	req := entities.CheckoutInitRequest{
		OrderAmount:        150000,
		OrderInvoiceNumber: "TEST_LINK_001",
		OrderDescription:   "Test thanh toan link",
	}

	data, err := h.paymentService.InitCheckout(ctx.Request.Context(), req)
	if err != nil {
		ctx.String(http.StatusInternalServerError, "Error generating checkout: %v", err)
		return
	}

	htmlContent := `
	<!DOCTYPE html>
	<html>
	<head><title>Redirecting to SePay...</title></head>
	<body onload="document.getElementById('sepayForm').submit();">
		<p>Đang chuyển hướng sang cổng thanh toán SePay...</p>
		<form method="POST" action="` + data.CheckoutURL + `" id="sepayForm" style="display:none;">`
	
	for key, value := range data.FormValues {
		htmlContent += `<input type="hidden" name="` + key + `" value="` + value + `">`
	}

	htmlContent += `
		</form>
	</body>
	</html>`

	ctx.Data(http.StatusOK, "text/html; charset=utf-8", []byte(htmlContent))
}

