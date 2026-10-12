package queue

import (
	"github.com/monetr/monetr/server/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	jobsEnqueued = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace:   metrics.Namespace,
		Name:        "jobs_enqueued",
		Help:        "Number of jobs that have been enqueued for background work.",
		ConstLabels: map[string]string{},
	}, []string{
		"job_name",
	})
	jobsProcessed = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace:   metrics.Namespace,
		Name:        "jobs_processed",
		Help:        "Number of jobs that have been processed, processed jobs are jobs that have succeeded or failed.",
		ConstLabels: map[string]string{},
	}, []string{
		"job_name",
	})
	jobsFailed = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace:   metrics.Namespace,
		Name:        "jobs_failed",
		Help:        "Number of jobs that have failed.",
		ConstLabels: map[string]string{},
	}, []string{
		"job_name",
	})
	jobRunTime = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: metrics.Namespace,
		Name:      "job_run_time",
		Help:      "The amount of time it takes for jobs to run, in whole milliseconds.",
		Buckets:   []float64{1, 50, 100, 500, 1000, 10000},
	}, []string{
		"job_name",
	})
	jobWaitTime = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: metrics.Namespace,
		Name:      "job_wait_time",
		Help:      "Time between when a job was due to run and when a worker started it, in whole milliseconds. Jobs are scheduled to the second so this is only accurate to about a second.",
		Buckets:   []float64{1000, 5000, 10000, 30000, 60000, 300000, 900000},
	}, []string{
		"job_name",
	})
)
