package controller

import (
	"github.com/monetr/monetr/server/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	httpRequests = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace:   metrics.Namespace,
		Name:        "http_requests",
		Help:        "Number of HTTP requests received.",
		ConstLabels: map[string]string{},
	}, []string{
		"path",
		"method",
		"status",
	})
	httpResponseTime = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: metrics.Namespace,
		Name:      "http_response_time",
		Help:      "Time it takes for an HTTP request to be completed, in whole milliseconds.",
		Buckets:   []float64{1, 50, 100, 500, 1000, 10000},
	}, []string{
		"path",
		"method",
		"status",
	})
)
