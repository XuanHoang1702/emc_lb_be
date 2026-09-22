package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests processed, partitioned by status code, method and route.",
		},
		[]string{"method", "route", "status"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Latency of HTTP requests in seconds, partitioned by method and route.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "route"},
	)

	httpResponseSize = promauto.NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "http_response_size_bytes",
			Help: "Size of HTTP responses in bytes, partitioned by method, route and status.",
		},
		[]string{"method", "route", "status"},
	)
)

// PrometheusMiddleware collects core HTTP metrics for Prometheus.
// It uses ctx.FullPath() to prevent high cardinality issues with dynamic IDs in URLs.
func PrometheusMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()

		// Process request
		ctx.Next()

		// Calculate metrics after request is processed
		duration := time.Since(start).Seconds()
		status := strconv.Itoa(ctx.Writer.Status())
		method := ctx.Request.Method

		// Use route template to avoid high cardinality (e.g., /api/v1/users/:id)
		route := ctx.FullPath()
		if route == "" {
			// If route is empty, it means no registered route was matched (e.g., 404 Not Found)
			route = "404_not_found"
		}

		size := float64(ctx.Writer.Size())
		if size < 0 {
			size = 0
		}

		// Record metrics
		httpRequestsTotal.WithLabelValues(method, route, status).Inc()
		httpRequestDuration.WithLabelValues(method, route).Observe(duration)
		httpResponseSize.WithLabelValues(method, route, status).Observe(size)
	}
}
