package security

import (
	"context"
	"testing"

	"emc_lb/src/internal/service"
	"emc_lb/src/pkg/config"
	"emc_lb/src/pkg/entities"
	"github.com/stretchr/testify/assert"
)

// MockOrderService is needed to mock GetOrderByInvoiceNumber
type mockOrderService struct {
	service.OrderService // embed to satisfy interface
}

func (m *mockOrderService) GetOrderByInvoiceNumber(ctx context.Context, invoice string) (entities.OrderResponse, error) {
	return entities.OrderResponse{
		InvoiceNumber: invoice,
		TotalAmount:   100000,
	}, nil
}

func TestSEC10_CallbackURLValidation(t *testing.T) {
	// Setup config with specific allowed hosts
	cfg := &config.AppConfig{
		Payment: config.PaymentSettings{
			SepayEnv:                  "production",
			SepayMerchantID:           "test_merchant",
			SepaySecretKey:            "test_secret",
			SepaySuccessURL:           "https://shop.emc.com/success",
			SepayErrorURL:             "https://shop.emc.com/error",
			SepayCancelURL:            "https://shop.emc.com/cancel",
			SepayAllowedCallbackHosts: "shop.emc.com",
		},
	}

	paymentSvc := service.NewPaymentService(cfg, &mockOrderService{})
	ctx := context.Background()

	tests := []struct {
		name       string
		successURL string
		errorURL   string
		expectErr  bool
	}{
		{"Valid overriding URLs", "https://shop.emc.com/new-success", "https://shop.emc.com/new-error", false},
		{"Valid default URLs (empty override)", "", "", false},
		{"Invalid success host", "https://evil.example/success", "", true},
		{"Invalid error scheme", "https://shop.emc.com/success", "http://shop.emc.com/error", true},
		{"Embedded credentials", "https://user:pass@shop.emc.com/success", "", true},
		{"Javascript injection", "javascript:alert(1)", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := entities.CheckoutInitRequest{
				OrderInvoiceNumber: "INV123",
				OrderAmount:        100000,
				SuccessURL:         tt.successURL,
				ErrorURL:           tt.errorURL,
			}

			_, err := paymentSvc.InitCheckout(ctx, req)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
