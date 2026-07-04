package service

import (
	"context"
	"log"
	"os"

	"emc_lb/src/pkg/entities"

	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"
)

type PaymentService interface {
	InitCheckout(ctx context.Context, req entities.CheckoutInitRequest) (*entities.CheckoutInitResponse, error)
	ProcessWebhook(ctx context.Context, req entities.SePayWebhookRequest) error
}

type paymentService struct {
	merchantID string
	secretKey  string
}

func NewPaymentService() PaymentService {
	return &paymentService{
		merchantID: os.Getenv("CLIENT_KEY"),
		secretKey:  os.Getenv("SECREC_KEY"),
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

func (s *paymentService) ProcessWebhook(ctx context.Context, req entities.SePayWebhookRequest) error {
	// TODO: Log the webhook or update Order/Transaction status in database
	log.Printf("Received payment via SePay webhook: ID=%d, Amount=%f, Code=%s", req.ID, req.TransferAmount, req.Code)

	return nil
}
