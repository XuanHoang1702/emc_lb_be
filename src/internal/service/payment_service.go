package service

import (
	"context"
	"log"
	"net/http"

	"emc_lb/src/pkg/entities"
	erres "emc_lb/src/pkg/errors"
	"emc_lb/src/pkg/res"
	"emc_lb/src/pkg/utils"

	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"
)

type PaymentService interface {
	InitCheckout(ctx context.Context, req entities.CheckoutInitRequest) (*entities.CheckoutInitResponse, error)
	ProcessWebhook(ctx context.Context, req entities.SePayWebhookRequest, authHeader string) error
}

type paymentService struct {
	merchantID   string
	secretKey    string
	orderService OrderService
}

func NewPaymentService(orderService OrderService) PaymentService {
	return &paymentService{
		merchantID:   utils.GetEnv("CLIENT_KEY", ""),
		secretKey:    utils.GetEnv("SECRET_KEY", ""),
		orderService: orderService,
	}
}

func (s *paymentService) InitCheckout(ctx context.Context, req entities.CheckoutInitRequest) (*entities.CheckoutInitResponse, error) {
	// Build map of signable fields manually because the SDK has a bug skipping "env"
	fields := map[string]string{
		"merchant":             s.merchantID,
		"env":                  "sandbox", // explicitly add env
		"operation":            "PURCHASE",
		"order_amount":         fmt.Sprintf("%.0f", req.OrderAmount), // format without trailing zeros
		"currency":             "VND",
		"order_invoice_number": req.OrderInvoiceNumber,
		"order_description":    req.OrderDescription,
	}

	// Canonical order required by SePay
	signFieldOrder := []string{
		"merchant",
		"env",
		"operation",
		"payment_method",
		"order_amount",
		"currency",
		"order_invoice_number",
		"order_description",
		"customer_id",
		"agreement_id",
		"agreement_name",
		"agreement_type",
		"agreement_payment_frequency",
		"agreement_amount_per_payment",
		"success_url",
		"error_url",
		"cancel_url",
		"order_id",
	}

	var parts []string
	for _, key := range signFieldOrder {
		val, ok := fields[key]
		if !ok {
			continue
		}
		parts = append(parts, key+"="+val)
	}

	mac := hmac.New(sha256.New, []byte(s.secretKey))
	mac.Write([]byte(strings.Join(parts, ",")))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	fields["signature"] = signature

	return &entities.CheckoutInitResponse{
		CheckoutURL: "https://pay-sandbox.sepay.vn/v1/checkout/init",
		FormValues:  fields,
	}, nil
}

func (s *paymentService) ProcessWebhook(ctx context.Context, req entities.SePayWebhookRequest, authHeader string) error {
	// API KEY or Bearer token validation
	expectedToken := "Apikey " + s.secretKey
	if authHeader != expectedToken && authHeader != "Bearer "+s.secretKey {
		return &res.AppError{
			Message:    "Unauthorized webhook signature",
			Code:       erres.UserUnauthorized,
			StatusCode: http.StatusUnauthorized,
		}
	}

	log.Printf("Received payment via SePay webhook: ID=%d, Amount=%f, Code=%s", req.ID, req.TransferAmount, req.Code)

	// SePay sends the transfer content in req.Code (which could be the invoice number like INV-1234)
	invoiceNumber := strings.TrimSpace(req.Code)
	if err := s.orderService.MarkAsPaidByInvoice(ctx, invoiceNumber); err != nil {
		log.Printf("Failed to mark order as paid: %v", err)
		return err
	}

	return nil
}
