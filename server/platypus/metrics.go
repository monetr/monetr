package platypus

import (
	"github.com/monetr/monetr/server/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	plaidApiRequests = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace:   metrics.Namespace,
		Name:        "plaid_api_requests",
		Help:        "Number of requests made to Plaid's API, labeled by the endpoint and the HTTP status code. The status is error if no response was received at all.",
		ConstLabels: map[string]string{},
	}, []string{
		"endpoint",
		"status",
	})
	plaidApiRequestTime = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: metrics.Namespace,
		Name:      "plaid_api_request_time",
		Help:      "Time it takes for a request to Plaid's API to be completed, in whole milliseconds.",
		Buckets:   []float64{50, 100, 250, 500, 1000, 2500, 5000, 10000, 30000},
	}, []string{
		"endpoint",
	})
	plaidApiErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace:   metrics.Namespace,
		Name:        "plaid_api_errors",
		Help:        "Number of errors returned by Plaid's API, labeled by the endpoint and Plaid's error code.",
		ConstLabels: map[string]string{},
	}, []string{
		"endpoint",
		"error_code",
	})
)

// Start every Plaid endpoint we call at 0 for successful requests. Other status
// codes and error codes come from Plaid so those will still start at 1. If a
// new Plaid API call gets added it should be added here too.
func init() {
	for _, endpoint := range []string{
		"/accounts/get",
		"/institutions/get_by_id",
		"/item/public_token/exchange",
		"/item/remove",
		"/item/webhook/update",
		"/link/token/create",
		"/transactions/get",
		"/transactions/refresh",
		"/transactions/sync",
		"/webhook_verification_key/get",
	} {
		plaidApiRequests.With(prometheus.Labels{
			"endpoint": endpoint,
			"status":   "200",
		})
		plaidApiRequestTime.With(prometheus.Labels{
			"endpoint": endpoint,
		})
	}
}
