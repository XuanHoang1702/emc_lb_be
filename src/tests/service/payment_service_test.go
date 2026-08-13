package service_test

import (
	"context"
	"net/http"
	"testing"

	"emc_lb/src/internal/service"
	"emc_lb/src/pkg/entities"
	"emc_lb/src/pkg/res"
)

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
	if s.markAsPaidFn != nil {
		return s.markAsPaidFn(invoiceNumber, amount)
	}
	return nil
}

func (s *stubPaymentOrderSvc) UpdateOrderStatus(_ context.Context, _ string, _ string) error {
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

	svc := service.NewPaymentService(orderSvc)

	err := svc.ProcessIPN(context.Background(), entities.SePayIPNRequest{
		NotificationType: "ORDER_PAID",
		Order: entities.SePayIPNOrder{
			OrderStatus:        "CAPTURED",
			OrderInvoiceNumber: "INV-12345",
			OrderAmount:        "500000",
		},
		Transaction: entities.SePayIPNTx{
			PaymentMethod: "BANK_TRANSFER",
		},
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if markedInvoice != "INV-12345" {
		t.Fatalf("expected invoice INV-12345 to be marked as paid, got %s", markedInvoice)
	}
}

func TestProcessIPN_IgnoreNonOrderPaid(t *testing.T) {
	svc := service.NewPaymentService(&stubPaymentOrderSvc{})

	err := svc.ProcessIPN(context.Background(), entities.SePayIPNRequest{
		NotificationType: "ORDER_CREATED",
	})
	if err != nil {
		t.Fatalf("expected no error for non ORDER_PAID notification, got %v", err)
	}
}

func TestProcessIPN_IgnoreNonCaptured(t *testing.T) {
	svc := service.NewPaymentService(&stubPaymentOrderSvc{})

	err := svc.ProcessIPN(context.Background(), entities.SePayIPNRequest{
		NotificationType: "ORDER_PAID",
		Order: entities.SePayIPNOrder{
			OrderStatus: "PENDING",
		},
	})
	if err != nil {
		t.Fatalf("expected no error for non CAPTURED status, got %v", err)
	}
}

func TestProcessIPN_MissingInvoice(t *testing.T) {
	svc := service.NewPaymentService(&stubPaymentOrderSvc{})

	err := svc.ProcessIPN(context.Background(), entities.SePayIPNRequest{
		NotificationType: "ORDER_PAID",
		Order: entities.SePayIPNOrder{
			OrderStatus:        "CAPTURED",
			OrderInvoiceNumber: "",
		},
	})
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
	// NewPaymentService reads from env; with no env set, merchant/secret are empty
	svc := service.NewPaymentService(&stubPaymentOrderSvc{})

	_, err := svc.InitCheckout(context.Background(), entities.CheckoutInitRequest{
		OrderAmount:        500000,
		OrderInvoiceNumber: "INV-1",
	})
	if err == nil {
		t.Fatal("expected error when payment gateway not configured")
	}
}
