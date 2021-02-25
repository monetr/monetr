package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"strconv"
	"time"
)

type Stats struct {
	JobsEnqueued     *prometheus.CounterVec
	JobsProcessed    *prometheus.CounterVec
	JobsFailed       *prometheus.CounterVec
	JobRunTime       *prometheus.HistogramVec
	Queries          *prometheus.CounterVec
	HTTPRequests     *prometheus.CounterVec
	HTTPResponseType *prometheus.HistogramVec
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
		JobRunTime: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "harder",
			Name:      "job_run_time",
			Help:      "The amount of time it takes for jobs to run.",
			Buckets:   []float64{1, 100, 1000, 10000},
		}, []string{
			"job_name",
			"account_id",
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

func (s *Stats) JobFinished(name string, accountId uint64, start time.Time) {
	s.JobRunTime.With(prometheus.Labels{
		"job_name":   name,
		"account_id": strconv.FormatUint(accountId, 10),
	}).Observe(float64(time.Since(start).Milliseconds()))
	s.JobsProcessed.With(prometheus.Labels{
		"job_name": name,
	}).Inc()
}
