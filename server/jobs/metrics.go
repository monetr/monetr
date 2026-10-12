package jobs

import (
	"github.com/monetr/monetr/server/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	jobsStalled = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace:   metrics.Namespace,
		Name:        "jobs_stalled",
		Help:        "Number of stalled jobs that were recovered, labeled by whether they were retried or marked as failed because they had no attempts left.",
		ConstLabels: map[string]string{},
	}, []string{
		"result",
	})
)
