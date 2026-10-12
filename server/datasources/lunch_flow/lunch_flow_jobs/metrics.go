package lunch_flow_jobs

import (
	"github.com/monetr/monetr/server/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	lunchFlowSyncs = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace:   metrics.Namespace,
		Name:        "lunch_flow_syncs",
		Help:        "Number of Lunch Flow syncs performed, labeled by whether the sync succeeded or failed. A failed sync may be retried and counted again.",
		ConstLabels: map[string]string{},
	}, []string{
		"result",
	})
)

// Start both results at 0 so prometheus can see the first sync.
func init() {
	for _, result := range []string{
		"success",
		"failure",
	} {
		lunchFlowSyncs.With(prometheus.Labels{
			"result": result,
		})
	}
}
