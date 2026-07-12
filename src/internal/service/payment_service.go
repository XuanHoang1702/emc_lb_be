package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"strings"

	"emc_lb/src/pkg/entities"
	erres "emc_lb/src/pkg/errors"
	"emc_lb/src/pkg/res"
	"emc_lb/src/pkg/utils"
)

type PaymentService interface {
	InitCheckout(ctx context.Context, req entities.CheckoutInitRequest) (*entities.CheckoutInitResponse, error)
	ProcessIPN(ctx context.Context, req entities.SePayIPNRequest) error
}

type paymentService struct {
	merchantID  string
	secretKey   string
	env         string // "sandbox" or "production"
	successURL  string
	errorURL    string
	cancelURL   string
	orderService OrderService
}

func NewPaymentService(orderService OrderService) PaymentService {
	env := utils.GetEnv("SEPAY_ENV", "sandbox")
	return &paymentService{
		merchantID:   utils.GetEnv("SEPAY_MERCHANT_ID", ""),
		secretKey:    utils.GetEnv("SEPAY_SECRET_KEY", ""),
		env:          env,
		successURL:   utils.GetEnv("SEPAY_SUCCESS_URL", ""),
		errorURL:     utils.GetEnv("SEPAY_ERROR_URL", ""),
		cancelURL:    utils.GetEnv("SEPAY_CANCEL_URL", ""),
		orderService: orderService,
	}
}

func (s *paymentService) checkoutBaseURL() string {
	if s.env == "production" {
		return "https://pay.sepay.vn/v1/checkout/init"
	}
	return "https://pay-sandbox.sepay.vn/v1/checkout/init"
}

func (s *paymentService) InitCheckout(ctx context.Context, req entities.CheckoutInitRequest) (*entities.CheckoutInitResponse, error) {
	if s.merchantID == "" || s.secretKey == "" {
		return nil, &res.AppError{
			Message:    "Payment gateway not configured",
			Code:       erres.CommonInternal,
			StatusCode: http.StatusInternalServerError,
		}
	}

	// Resolve callback URLs: request overrides > env config
	successURL := s.successURL
	if req.SuccessURL != "" {
		successURL = req.SuccessURL
	}
	errorURL := s.errorURL
	if req.ErrorURL != "" {
		errorURL = req.ErrorURL
	}
	cancelURL := s.cancelURL
	if req.CancelURL != "" {
		cancelURL = req.CancelURL
	}

	// Build form fields per official SePay documentation
	fields := map[string]string{
		"merchant":             s.merchantID,
		"operation":            "PURCHASE",
		"order_amount":         fmt.Sprintf("%d", int64(req.OrderAmount)),
		"currency":             "VND",
		"order_invoice_number": req.OrderInvoiceNumber,
		"order_description":    req.OrderDescription,
	}

	if req.PaymentMethod != "" {
		fields["payment_method"] = req.PaymentMethod
	}
	if req.CustomerID != "" {
		fields["customer_id"] = req.CustomerID
	}
	if successURL != "" {
		fields["success_url"] = successURL
	}
	if errorURL != "" {
		fields["error_url"] = errorURL
	}
	if cancelURL != "" {
		fields["cancel_url"] = cancelURL
	}

	// Canonical field order per official SePay PHP signFields()
	// Reference: https://developer.sepay.vn/vi/cong-thanh-toan/gioi-thieu
	signFieldOrder := []string{
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
	}

	var parts []string
	for _, key := range signFieldOrder {
		val, ok := fields[key]
		if !ok {
			continue
		}
		parts = append(parts, key+"="+val)
	}
	signString := strings.Join(parts, ",")
	log.Printf("[SePay DEBUG] Sign string: %s", signString)

	mac := hmac.New(sha256.New, []byte(s.secretKey))
	mac.Write([]byte(signString))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	fields["signature"] = signature

	log.Printf("[SePay DEBUG] Checkout URL: %s", s.checkoutBaseURL())
	for k, v := range fields {
		log.Printf("[SePay DEBUG] Field: %s = %s", k, v)
	}

	return &entities.CheckoutInitResponse{
		CheckoutURL: s.checkoutBaseURL(),
		FormValues:  fields,
	}, nil
}

func (s *paymentService) ProcessIPN(ctx context.Context, req entities.SePayIPNRequest) error {
	if req.NotificationType != "ORDER_PAID" {
		log.Printf("Ignoring SePay IPN with type: %s", req.NotificationType)
		return nil
	}

	if req.Order.OrderStatus != "CAPTURED" {
		log.Printf("Ignoring SePay IPN with order status: %s", req.Order.OrderStatus)
		return nil
	}

	invoiceNumber := strings.TrimSpace(req.Order.OrderInvoiceNumber)
	if invoiceNumber == "" {
		return &res.AppError{
			Message:    "Missing order_invoice_number in IPN",
			Code:       erres.CommonBadRequest,
			StatusCode: http.StatusBadRequest,
		}
	}

	log.Printf("SePay IPN: ORDER_PAID invoice=%s amount=%s method=%s",
		invoiceNumber, req.Order.OrderAmount, req.Transaction.PaymentMethod)

	if err := s.orderService.MarkAsPaidByInvoice(ctx, invoiceNumber); err != nil {
		log.Printf("Failed to mark order as paid: %v", err)
		return err
	}

	return nil
}
