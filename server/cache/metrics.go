package cache

import (
	"time"

	"github.com/monetr/monetr/server/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	cacheOperations = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace:   metrics.Namespace,
		Name:        "cache_operations",
		Help:        "Number of cache operations, labeled by the operation and the result. Gets are a hit, miss or error.",
		ConstLabels: map[string]string{},
	}, []string{
		"operation",
		"result",
	})
	cacheOperationTime = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: metrics.Namespace,
		Name:      "cache_operation_time",
		Help:      "Time it takes for a cache operation to be completed.",
		Buckets:   []float64{0.5, 1, 5, 10, 50, 100, 500},
	}, []string{
		"operation",
		"result",
	})
)

func recordOperation(operation, result string, start time.Time) {
	labels := prometheus.Labels{
		"operation": operation,
		"result":    result,
	}
	cacheOperations.With(labels).Inc()
	// Cache operations are usually well under a millisecond, so keep the fraction
	// instead of using Milliseconds() like the other metrics do. Otherwise
	// everything would just be 0.
	cacheOperationTime.With(labels).Observe(float64(time.Since(start)) / float64(time.Millisecond))
}
