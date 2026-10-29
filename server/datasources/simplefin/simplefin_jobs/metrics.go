package simplefin_jobs

import (
	"github.com/monetr/monetr/server/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	simpleFinSyncs = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace:   metrics.Namespace,
		Name:        "simplefin_syncs",
		Help:        "Number of SimpleFIN syncs performed, labeled by whether the sync succeeded, failed or was denied by SimpleFIN. Denied is when SimpleFIN wants payment or the token was revoked.",
		ConstLabels: map[string]string{},
	}, []string{
		"result",
	})
)

// Start all the results at 0 so prometheus can see the first sync.
func init() {
	for _, result := range []string{
		"success",
		"failure",
		"denied",
	} {
		simpleFinSyncs.With(prometheus.Labels{
			"result": result,
		})
	}
}
