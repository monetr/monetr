package powchallenge

import (
	"github.com/monetr/monetr/server/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	powIssued = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace:   metrics.Namespace,
		Name:        "pow_issued",
		Help:        "Number of proof of work challenges that have been issued to clients.",
		ConstLabels: map[string]string{},
	}, []string{})
	powVerified = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace:   metrics.Namespace,
		Name:        "pow_verified",
		Help:        "Number of proof of work challenge verifications, labeled by the result of the verification.",
		ConstLabels: map[string]string{},
	}, []string{
		"result",
	})
	powVerifyTime = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: metrics.Namespace,
		Name:      "pow_verify_time",
		Help:      "Time it takes to verify a proof of work challenge, in whole milliseconds.",
		Buckets:   []float64{1, 50, 100, 500, 1000, 10000},
	}, []string{
		"result",
	})
)
