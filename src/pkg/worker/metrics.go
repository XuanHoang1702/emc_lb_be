package worker

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	OrderCancellationSuccess = promauto.NewCounter(prometheus.CounterOpts{
		Name: "emc_lb_order_cancellation_success_total",
		Help: "The total number of successful automatic order cancellations",
	})
	OrderCancellationFailure = promauto.NewCounter(prometheus.CounterOpts{
		Name: "emc_lb_order_cancellation_failure_total",
		Help: "The total number of failed automatic order cancellations",
	})
	OrderCancellationInventoryReleaseFailure = promauto.NewCounter(prometheus.CounterOpts{
		Name: "emc_lb_order_cancellation_inventory_release_failure_total",
		Help: "The total number of failures to release inventory during cancellation",
	})
)
