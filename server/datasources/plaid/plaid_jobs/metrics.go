package plaid_jobs

import (
	"github.com/monetr/monetr/server/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	plaidSyncs = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace:   metrics.Namespace,
		Name:        "plaid_syncs",
		Help:        "Number of Plaid syncs performed, labeled by whether the sync succeeded or failed. A failed sync may be retried and counted again.",
		ConstLabels: map[string]string{},
	}, []string{
		"result",
	})
)
