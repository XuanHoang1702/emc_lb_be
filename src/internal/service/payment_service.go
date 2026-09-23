package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"time"

	"emc_lb/src/pkg/config"
	"emc_lb/src/pkg/entities"
	erres "emc_lb/src/pkg/errors"
	"emc_lb/src/pkg/logs"
	"emc_lb/src/pkg/res"
	"emc_lb/src/pkg/utils"
)

type PaymentService interface {
	InitCheckout(ctx context.Context, req entities.CheckoutInitRequest) (*entities.CheckoutInitResponse, error)
	ProcessIPN(ctx context.Context, req entities.SePayIPNRequest, secretHeader string) error
}

type paymentService struct {
	merchantID   string
	secretKey    string
	env          string // "sandbox" or "production"
	successURL   string
	errorURL     string
	cancelURL    string
	orderService OrderService
}

func NewPaymentService(cfg *config.AppConfig, orderService OrderService) PaymentService {
	return &paymentService{
		merchantID:   cfg.Payment.SepayMerchantID,
		secretKey:    cfg.Payment.SepaySecretKey,
		env:          cfg.Payment.SepayEnv,
		successURL:   cfg.Payment.SepaySuccessURL,
		errorURL:     cfg.Payment.SepayErrorURL,
		cancelURL:    cfg.Payment.SepayCancelURL,
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

	if req.OrderInvoiceNumber == "TEST_LINK_001" && s.env != "production" {
		// Bypass database check for test endpoint in sandbox/dev
	} else {
		order, err := s.orderService.GetOrderByInvoiceNumber(ctx, req.OrderInvoiceNumber)
		if err != nil {
			return nil, err
		}

		if !utils.MoneyEqual(order.TotalAmount, req.OrderAmount) {
			return nil, &res.AppError{
				Message:    "Order amount mismatch",
				Code:       erres.CommonBadRequest,
				StatusCode: http.StatusBadRequest,
			}
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

	mac := hmac.New(sha256.New, []byte(s.secretKey))
	mac.Write([]byte(signString))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	fields["signature"] = signature

	return &entities.CheckoutInitResponse{
		CheckoutURL: s.checkoutBaseURL(),
		FormValues:  fields,
	}, nil
}

func (s *paymentService) ProcessIPN(ctx context.Context, req entities.SePayIPNRequest, secretHeader string) error {
	// The gateway must be configured before any IPN is accepted.
	if s.secretKey == "" {
		return &res.AppError{
			Message:    "Payment gateway not configured",
			Code:       erres.CommonInternal,
			StatusCode: http.StatusInternalServerError,
		}
	}

	// Authenticate the IPN with the X-Secret-Key header (constant-time compare).
	if !hmac.Equal([]byte(secretHeader), []byte(s.secretKey)) {
		return &res.AppError{
			Message:    "Invalid IPN secret key",
			Code:       erres.CommonUnauthorized,
			StatusCode: http.StatusUnauthorized,
		}
	}

	// Replay protection: reject notifications older than 5 minutes (SePay
	// recommends this window). Requests without a timestamp are rejected too.
	if req.Timestamp == 0 || time.Since(time.Unix(req.Timestamp, 0)).Abs() > 5*time.Minute {
		return &res.AppError{
			Message:    "IPN timestamp expired",
			Code:       erres.CommonBadRequest,
			StatusCode: http.StatusBadRequest,
		}
	}

	if req.NotificationType != "ORDER_PAID" {
		logs.WithContext(ctx).Info("ignoring SePay IPN with type", "type", req.NotificationType)
		return nil
	}

	if req.Order.OrderStatus != "CAPTURED" {
		logs.WithContext(ctx).Info("ignoring SePay IPN with order status", "status", req.Order.OrderStatus)
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

	var paidAmount float64
	if _, err := fmt.Sscanf(req.Order.OrderAmount, "%f", &paidAmount); err != nil {
		return &res.AppError{
			Message:    "Invalid order_amount in IPN",
			Code:       erres.CommonBadRequest,
			StatusCode: http.StatusBadRequest,
		}
	}

	logs.WithContext(ctx).Info("SePay IPN: ORDER_PAID",
		"invoice", invoiceNumber,
		"amount", req.Order.OrderAmount,
		"method", req.Transaction.PaymentMethod)

	if err := s.orderService.ConfirmPayment(ctx, invoiceNumber, paidAmount, req.Transaction.TransactionID); err != nil {
		logs.WithContext(ctx).Error("failed to mark order as paid",
			"invoice", invoiceNumber, "error", err)
		return err
	}

	return nil
}
