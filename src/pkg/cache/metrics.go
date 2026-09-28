package cache

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	CacheHitsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cache_hits_total",
			Help: "Total number of cache hits, partitioned by namespace.",
		},
		[]string{"namespace"},
	)

	CacheMissesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cache_misses_total",
			Help: "Total number of cache misses, partitioned by namespace.",
		},
		[]string{"namespace"},
	)

	CacheGetErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cache_get_errors_total",
			Help: "Total number of cache get errors, partitioned by namespace.",
		},
		[]string{"namespace"},
	)

	CacheSetErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cache_set_errors_total",
			Help: "Total number of cache set errors, partitioned by namespace.",
		},
		[]string{"namespace"},
	)

	CacheInvalidationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cache_invalidations_total",
			Help: "Total number of cache invalidations, partitioned by namespace.",
		},
		[]string{"namespace"},
	)

	CacheGetDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "cache_get_duration_seconds",
			Help:    "Latency of cache gets in seconds, partitioned by namespace.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"namespace"},
	)

	CacheSetDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "cache_set_duration_seconds",
			Help:    "Latency of cache sets in seconds, partitioned by namespace.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"namespace"},
	)
)
