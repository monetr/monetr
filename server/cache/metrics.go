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
		Help:      "Time it takes for a cache operation to be completed, in milliseconds. Keeps fractions of a millisecond.",
		Buckets:   []float64{0.5, 1, 5, 10, 50, 100, 500},
	}, []string{
		"operation",
		"result",
	})
)

// Create every operation and result we can record up front so they start at 0,
// otherwise prometheus can't see the first one.
func init() {
	for operation, results := range map[string][]string{
		"get": {
			"hit",
			"miss",
			"error",
		},
		"set": {
			"ok",
			"error",
		},
		"delete": {
			"ok",
			"error",
		},
		"compare_and_swap": {
			"swapped",
			"not_swapped",
			"error",
		},
	} {
		for _, result := range results {
			labels := prometheus.Labels{
				"operation": operation,
				"result":    result,
			}
			cacheOperations.With(labels)
			cacheOperationTime.With(labels)
		}
	}
}

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
