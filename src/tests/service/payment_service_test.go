package service_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"emc_lb/src/internal/service"
	"emc_lb/src/pkg/config"
	"emc_lb/src/pkg/entities"
	"emc_lb/src/pkg/res"
)

func testPaymentConfig(secretKey string) *config.AppConfig {
	return &config.AppConfig{
		Payment: config.PaymentSettings{
			SepayEnv:        "sandbox",
			SepayMerchantID: "test-merchant",
			SepaySecretKey:  secretKey,
			SepaySuccessURL: "http://localhost/success",
			SepayErrorURL:   "http://localhost/error",
			SepayCancelURL:  "http://localhost/cancel",
		},
	}
}

// stubOrderSvc for payment testing
type stubPaymentOrderSvc struct {
	markAsPaidFn func(invoiceNumber string, amount float64) error
}

func (s *stubPaymentOrderSvc) CreateOrder(_ context.Context, _ string, _ entities.CreateOrderRequest) ([]entities.OrderResponse, error) {
	return nil, nil
}

func (s *stubPaymentOrderSvc) ListOrders(_ context.Context, _ string) ([]entities.OrderResponse, error) {
	return nil, nil
}

func (s *stubPaymentOrderSvc) GetOrder(_ context.Context, _ string) (entities.OrderResponse, error) {
	return entities.OrderResponse{}, nil
}

func (s *stubPaymentOrderSvc) GetOrderByInvoiceNumber(_ context.Context, invoiceNumber string) (entities.OrderResponse, error) {
	return entities.OrderResponse{TotalAmount: 500000}, nil
}

func (s *stubPaymentOrderSvc) MarkAsPaidByInvoice(_ context.Context, invoiceNumber string, amount float64) error {
	return nil
}

func (s *stubPaymentOrderSvc) ConfirmPayment(_ context.Context, invoiceNumber string, amount float64, transactionID string) error {
	if s.markAsPaidFn != nil {
		return s.markAsPaidFn(invoiceNumber, amount)
	}
	return nil
}

func (s *stubPaymentOrderSvc) UpdateOrderStatus(_ context.Context, _ string, _ string) error {
	return nil
}
func (s *stubPaymentOrderSvc) CreateOrderFromCheckout(ctx context.Context, userID string, req entities.CheckoutRequest) ([]entities.OrderResponse, error) {
	return nil, nil
}
func (s *stubPaymentOrderSvc) CancelOrder(ctx context.Context, userID string, orderID string) error {
	return nil
}
func (s *stubPaymentOrderSvc) ExpireOrder(ctx context.Context, orderID string) error {
	return nil
}

func TestProcessIPN_OrderPaid(t *testing.T) {
	var markedInvoice string
	orderSvc := &stubPaymentOrderSvc{
		markAsPaidFn: func(inv string, amount float64) error {
			markedInvoice = inv
			return nil
		},
	}

	svc := service.NewPaymentService(testPaymentConfig("test-secret"), orderSvc)

	err := svc.ProcessIPN(context.Background(), entities.SePayIPNRequest{
		Timestamp:        time.Now().Unix(),
		NotificationType: "ORDER_PAID",
		Order: entities.SePayIPNOrder{
			OrderStatus:        "CAPTURED",
			OrderInvoiceNumber: "INV-12345",
			OrderAmount:        "500000",
		},
		Transaction: entities.SePayIPNTx{
			PaymentMethod: "BANK_TRANSFER",
		},
	}, "test-secret")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if markedInvoice != "INV-12345" {
		t.Fatalf("expected invoice INV-12345 to be marked as paid, got %s", markedInvoice)
	}
}

func TestProcessIPN_InvalidSecret(t *testing.T) {
	t.Setenv("SEPAY_SECRET_KEY", "correct-secret")

	svc := service.NewPaymentService(testPaymentConfig("test-secret"), &stubPaymentOrderSvc{})

	err := svc.ProcessIPN(context.Background(), entities.SePayIPNRequest{
		Timestamp:        time.Now().Unix(),
		NotificationType: "ORDER_PAID",
		Order: entities.SePayIPNOrder{
			OrderStatus:        "CAPTURED",
			OrderInvoiceNumber: "INV-12345",
		},
	}, "wrong-secret")

	if err == nil {
		t.Fatal("expected error for invalid secret key")
	}
	appErr, ok := err.(*res.AppError)
	if !ok {
		t.Fatal("expected AppError")
	}
	if appErr.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", appErr.StatusCode)
	}
}

func TestProcessIPN_ExpiredTimestamp(t *testing.T) {
	t.Setenv("SEPAY_SECRET_KEY", "test-secret")

	svc := service.NewPaymentService(testPaymentConfig("test-secret"), &stubPaymentOrderSvc{})

	err := svc.ProcessIPN(context.Background(), entities.SePayIPNRequest{
		Timestamp:        time.Now().Add(-10 * time.Minute).Unix(),
		NotificationType: "ORDER_PAID",
		Order: entities.SePayIPNOrder{
			OrderStatus:        "CAPTURED",
			OrderInvoiceNumber: "INV-12345",
		},
	}, "test-secret")

	if err == nil {
		t.Fatal("expected error for expired timestamp")
	}
	appErr, ok := err.(*res.AppError)
	if !ok {
		t.Fatal("expected AppError")
	}
	if appErr.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", appErr.StatusCode)
	}
}

func TestProcessIPN_IgnoreNonOrderPaid(t *testing.T) {
	t.Setenv("SEPAY_SECRET_KEY", "test-secret")

	svc := service.NewPaymentService(testPaymentConfig("test-secret"), &stubPaymentOrderSvc{})

	err := svc.ProcessIPN(context.Background(), entities.SePayIPNRequest{
		Timestamp:        time.Now().Unix(),
		NotificationType: "ORDER_CREATED",
	}, "test-secret")
	if err != nil {
		t.Fatalf("expected no error for non ORDER_PAID notification, got %v", err)
	}
}

func TestProcessIPN_IgnoreNonCaptured(t *testing.T) {
	t.Setenv("SEPAY_SECRET_KEY", "test-secret")

	svc := service.NewPaymentService(testPaymentConfig("test-secret"), &stubPaymentOrderSvc{})

	err := svc.ProcessIPN(context.Background(), entities.SePayIPNRequest{
		Timestamp:        time.Now().Unix(),
		NotificationType: "ORDER_PAID",
		Order: entities.SePayIPNOrder{
			OrderStatus: "PENDING",
		},
	}, "test-secret")
	if err != nil {
		t.Fatalf("expected no error for non CAPTURED status, got %v", err)
	}
}

func TestProcessIPN_MissingInvoice(t *testing.T) {
	t.Setenv("SEPAY_SECRET_KEY", "test-secret")

	svc := service.NewPaymentService(testPaymentConfig("test-secret"), &stubPaymentOrderSvc{})

	err := svc.ProcessIPN(context.Background(), entities.SePayIPNRequest{
		Timestamp:        time.Now().Unix(),
		NotificationType: "ORDER_PAID",
		Order: entities.SePayIPNOrder{
			OrderStatus:        "CAPTURED",
			OrderInvoiceNumber: "",
		},
	}, "test-secret")
	if err == nil {
		t.Fatal("expected error for missing invoice number")
	}
	appErr, ok := err.(*res.AppError)
	if !ok {
		t.Fatal("expected AppError")
	}
	if appErr.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", appErr.StatusCode)
	}
}

func TestInitCheckout_MissingConfig(t *testing.T) {
	// Pass config with empty payment settings to trigger "not configured" error
	svc := service.NewPaymentService(&config.AppConfig{}, &stubPaymentOrderSvc{})

	_, err := svc.InitCheckout(context.Background(), entities.CheckoutInitRequest{
		OrderAmount:        500000,
		OrderInvoiceNumber: "INV-1",
	})
	if err == nil {
		t.Fatal("expected error when payment gateway not configured")
	}
}
