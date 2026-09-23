package handler

import (
	"html"
	"net/http"

	"emc_lb/src/internal/service"
	"emc_lb/src/pkg/entities"
	"emc_lb/src/pkg/errors"
	"emc_lb/src/pkg/logs"
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

// HandleInitCheckout godoc
// @Summary      Initialize SePay checkout
// @Description  Creates a checkout session with SePay Payment Gateway and returns the checkout URL
// @Tags         Payments
// @Accept       json
// @Produce      json
// @Param        body body entities.CheckoutInitRequest true "Checkout details"
// @Success      200  {object}  res.APIResponse
// @Failure      400  {object}  res.APIResponse
// @Failure      401  {object}  res.APIResponse
// @Failure      500  {object}  res.APIResponse
// @Security     BearerAuth
// @Router       /api/v1/payment/checkout [post]
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

// HandleSepayIPN godoc
// @Summary      SePay IPN Webhook
// @Description  Receives Instant Payment Notification from SePay when payment status changes.
// @Description  Authenticated via the X-Secret-Key header (merchant SECRET_KEY).
// @Tags         Payments
// @Accept       json
// @Produce      json
// @Param        X-Secret-Key header string true "SePay merchant secret key"
// @Param        body body entities.SePayIPNRequest true "IPN payload from SePay"
// @Success      200  {object}  res.APIResponse
// @Failure      400  {object}  res.APIResponse
// @Failure      401  {object}  res.APIResponse
// @Router       /api/v1/payment/ipn [post]
func (h *PaymentHandler) HandleSepayIPN(ctx *gin.Context) {
	var req entities.SePayIPNRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		res.Error(ctx, &res.AppError{
			Code:       errors.UserInvalidFormat,
			Message:    "Invalid IPN payload format",
			StatusCode: http.StatusBadRequest,
		})
		return
	}

	logs.WithContext(ctx.Request.Context()).Info("Received SePay IPN webhook", 
		"invoice", req.Order.OrderInvoiceNumber,
		"transaction_id", req.Transaction.TransactionID,
	)

	secretHeader := ctx.GetHeader("X-Secret-Key")
	if err := h.paymentService.ProcessIPN(ctx.Request.Context(), req, secretHeader); err != nil {
		res.Error(ctx, err)
		return
	}

	// SePay expects 200 OK to acknowledge receipt
	res.Success(ctx, http.StatusOK, "IPN received successfully")
}

// HandleTestCheckoutLink godoc
// @Summary      Test checkout link
// @Description  Renders an auto-submitting HTML form that redirects to SePay for testing
// @Tags         Payments
// @Produce      html
// @Success      200  {string}  string "HTML redirect form"
// @Failure      500  {string}  string "Error message"
// @Router       /api/v1/payment/test-checkout [get]
func (h *PaymentHandler) HandleTestCheckoutLink(ctx *gin.Context) {
	req := entities.CheckoutInitRequest{
		OrderAmount:        150000,
		OrderInvoiceNumber: "TEST_LINK_001",
		OrderDescription:   "Test thanh toan link",
		PaymentMethod:      "BANK_TRANSFER",
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
		<p>Redirecting to SePay Payment Gateway...</p>
		<form method="POST" action="` + data.CheckoutURL + `" id="sepayForm" style="display:none;">`

	keysOrder := []string{
		"merchant",
		"operation",
		"payment_method",
		"order_amount",
		"currency",
		"order_invoice_number",
		"order_description",
		"customer_id",
		"success_url",
		"error_url",
		"cancel_url",
		"signature",
	}

	for _, key := range keysOrder {
		if value, exists := data.FormValues[key]; exists {
			htmlContent += `<input type="hidden" name="` + html.EscapeString(key) + `" value="` + html.EscapeString(value) + `">`
		}
	}

	htmlContent += `
		</form>
	</body>
	</html>`

	ctx.Data(http.StatusOK, "text/html; charset=utf-8", []byte(htmlContent))
}

// HandlePaymentSuccess godoc
// @Summary      Payment success callback
// @Description  Handles the redirect from SePay after a successful payment
// @Tags         Payments
// @Produce      html
// @Success      200  {string}  string "Success page"
// @Router       /api/v1/payment/success [get]
func (h *PaymentHandler) HandlePaymentSuccess(ctx *gin.Context) {
	htmlContent := `
	<!DOCTYPE html>
	<html>
	<head>
		<title>Payment Successful</title>
		<style>
			body { font-family: Arial, sans-serif; text-align: center; padding: 50px; background-color: #f0fdf4; }
			h1 { color: #166534; }
			.container { background: white; padding: 30px; border-radius: 8px; box-shadow: 0 4px 6px rgba(0,0,0,0.1); max-width: 500px; margin: auto; }
			a { display: inline-block; margin-top: 20px; padding: 10px 20px; background: #16a34a; color: white; text-decoration: none; border-radius: 5px; }
		</style>
	</head>
	<body>
		<div class="container">
			<h1>🎉 Thanh toán thành công!</h1>
			<p>Cảm ơn bạn đã mua sắm. Đơn hàng của bạn đang được xử lý.</p>
			<a href="/">Quay lại trang chủ</a>
		</div>
	</body>
	</html>`
	ctx.Data(http.StatusOK, "text/html; charset=utf-8", []byte(htmlContent))
}

// HandlePaymentError godoc
// @Summary      Payment error callback
// @Description  Handles the redirect from SePay after a failed payment
// @Tags         Payments
// @Produce      html
// @Success      200  {string}  string "Error page"
// @Router       /api/v1/payment/error [get]
func (h *PaymentHandler) HandlePaymentError(ctx *gin.Context) {
	htmlContent := `
	<!DOCTYPE html>
	<html>
	<head>
		<title>Payment Failed</title>
		<style>
			body { font-family: Arial, sans-serif; text-align: center; padding: 50px; background-color: #fef2f2; }
			h1 { color: #991b1b; }
			.container { background: white; padding: 30px; border-radius: 8px; box-shadow: 0 4px 6px rgba(0,0,0,0.1); max-width: 500px; margin: auto; }
			a { display: inline-block; margin-top: 20px; padding: 10px 20px; background: #dc2626; color: white; text-decoration: none; border-radius: 5px; }
		</style>
	</head>
	<body>
		<div class="container">
			<h1>❌ Thanh toán thất bại</h1>
			<p>Đã xảy ra lỗi trong quá trình thanh toán. Vui lòng thử lại sau.</p>
			<a href="/">Quay lại trang chủ</a>
		</div>
	</body>
	</html>`
	ctx.Data(http.StatusOK, "text/html; charset=utf-8", []byte(htmlContent))
}

// HandlePaymentCancel godoc
// @Summary      Payment cancel callback
// @Description  Handles the redirect from SePay when user cancels payment
// @Tags         Payments
// @Produce      html
// @Success      200  {string}  string "Cancel page"
// @Router       /api/v1/payment/cancel [get]
func (h *PaymentHandler) HandlePaymentCancel(ctx *gin.Context) {
	htmlContent := `
	<!DOCTYPE html>
	<html>
	<head>
		<title>Payment Cancelled</title>
		<style>
			body { font-family: Arial, sans-serif; text-align: center; padding: 50px; background-color: #f8fafc; }
			h1 { color: #334155; }
			.container { background: white; padding: 30px; border-radius: 8px; box-shadow: 0 4px 6px rgba(0,0,0,0.1); max-width: 500px; margin: auto; }
			a { display: inline-block; margin-top: 20px; padding: 10px 20px; background: #475569; color: white; text-decoration: none; border-radius: 5px; }
		</style>
	</head>
	<body>
		<div class="container">
			<h1>⚠️ Đã hủy thanh toán</h1>
			<p>Bạn đã hủy giao dịch thanh toán. Đơn hàng của bạn sẽ không được xử lý.</p>
			<a href="/">Quay lại trang chủ</a>
		</div>
	</body>
	</html>`
	ctx.Data(http.StatusOK, "text/html; charset=utf-8", []byte(htmlContent))
}

