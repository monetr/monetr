package logging

import (
	"github.com/monetr/monetr/server/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	queries = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace:   metrics.Namespace,
		Name:        "queries",
		Help:        "Number of SQL queries issued. This excludes BEGIN, COMMIT and ROLLBACK. Essentially each PostgreSQL round trip.",
		ConstLabels: map[string]string{},
	}, []string{
		"stmt",
	})
	queryTime = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: metrics.Namespace,
		Name:      "query_time",
		Help:      "Time it takes for a SQL query to be completed, in whole milliseconds. This excludes BEGIN, COMMIT and ROLLBACK.",
		Buckets:   []float64{1, 5, 10, 50, 100, 500, 1000},
	}, []string{
		"stmt",
	})
)
