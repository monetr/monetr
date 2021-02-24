package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type Stats struct {
	JobsEnqueued  *prometheus.CounterVec
	JobsProcessed *prometheus.CounterVec
	JobsFailed    *prometheus.CounterVec
	Queries       *prometheus.CounterVec
	HTTPRequests  *prometheus.CounterVec
}

func NewStats() *Stats {
	return &Stats{
		JobsEnqueued: promauto.NewCounterVec(prometheus.CounterOpts{
			Namespace:   "harder",
			Name:        "jobs_enqueued",
			Help:        "Number of jobs that have been enqueued for background work.",
			ConstLabels: map[string]string{},
		}, []string{
			"job_name",
		}),
		JobsProcessed: promauto.NewCounterVec(prometheus.CounterOpts{
			Namespace:   "harder",
			Name:        "jobs_processed",
			Help:        "Number of jobs that have been processed, processed jobs are jobs that have succeeded or failed.",
			ConstLabels: map[string]string{},
		}, []string{
			"job_name",
		}),
		JobsFailed: promauto.NewCounterVec(prometheus.CounterOpts{
			Namespace:   "harder",
			Name:        "jobs_failed",
			Help:        "Number of jobs that have failed.",
			ConstLabels: map[string]string{},
		}, []string{
			"job_name",
		}),
		Queries: promauto.NewCounterVec(prometheus.CounterOpts{
			Namespace:   "harder",
			Name:        "queries",
			Help:        "Number of SQL queries issued. This includes BEGIN, COMMIT and ROLLBACK. Essentially each PostgreSQL round trip.",
			ConstLabels: map[string]string{},
		}, []string{}),
		HTTPRequests: promauto.NewCounterVec(prometheus.CounterOpts{
			Namespace:   "harder",
			Name:        "http_requests",
			Help:        "Number of HTTP requests received.",
			ConstLabels: map[string]string{},
		}, []string{
			"path",
		}),
	}
}

func (s *Stats) JobEnqueued(name string) {
	s.JobsEnqueued.With(prometheus.Labels{
		"job_name": name,
	}).Inc()
}
