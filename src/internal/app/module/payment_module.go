package module

import (
	"emc_lb/src/internal/handler"
	route "emc_lb/src/internal/routes"
	"emc_lb/src/internal/service"
)

type PaymentModule struct {
	routes route.Route
}

func NewPaymentModule() *PaymentModule {
	// If repository is needed in the future, inject it here
	paymentService := service.NewPaymentService()
	paymentHandler := handler.NewPaymentHandler(paymentService)
	paymentRoute := route.NewPaymentRoute(paymentHandler)

	return &PaymentModule{
		routes: paymentRoute,
	}
}

func (m *PaymentModule) Routes() route.Route {
	return m.routes
}
