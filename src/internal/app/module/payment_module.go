package module

import (
	"emc_lb/src/internal/handler"
	route "emc_lb/src/internal/routes"
	"emc_lb/src/internal/service"
	"emc_lb/src/pkg/config"
)

type PaymentModule struct {
	routes route.Route
}

func NewPaymentModule(cfg *config.AppConfig, orderService service.OrderService) *PaymentModule {
	paymentService := service.NewPaymentService(cfg, orderService)
	paymentHandler := handler.NewPaymentHandler(paymentService)
	paymentRoute := route.NewPaymentRoute(paymentHandler)

	return &PaymentModule{
		routes: paymentRoute,
	}
}

func (m *PaymentModule) Routes() route.Route {
	return m.routes
}
