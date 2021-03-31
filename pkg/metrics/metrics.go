package metrics

import (
	"github.com/kataras/iris/v12/context"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Stats struct {
	mux                 *http.ServeMux
	server              *http.Server
	listenAndServerSync sync.Once
	JobsEnqueued        *prometheus.CounterVec
	JobsProcessed       *prometheus.CounterVec
	JobsFailed          *prometheus.CounterVec
	JobRunTime          *prometheus.HistogramVec
	Queries             *prometheus.CounterVec
	QueryTime           *prometheus.HistogramVec
	HTTPRequests        *prometheus.CounterVec
	HTTPResponseTime    *prometheus.HistogramVec
}

func NewStats() *Stats {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	return &Stats{
		mux: mux,
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
			Buckets:   []float64{1, 50, 100, 500, 1000, 10000},
		}, []string{
			"job_name",
			"account_id",
		}),
		Queries: promauto.NewCounterVec(prometheus.CounterOpts{
			Namespace:   "harder",
			Name:        "queries",
			Help:        "Number of SQL queries issued. This excludes BEGIN, COMMIT and ROLLBACK. Essentially each PostgreSQL round trip.",
			ConstLabels: map[string]string{},
		}, []string{
			"stmt",
		}),
		HTTPRequests: promauto.NewCounterVec(prometheus.CounterOpts{
			Namespace:   "harder",
			Name:        "http_requests",
			Help:        "Number of HTTP requests received.",
			ConstLabels: map[string]string{},
		}, []string{
			"path",
			"method",
			"status",
		}),
		HTTPResponseTime: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "harder",
			Name:      "http_response_time",
			Help:      "Time it takes for an HTTP request to be completed.",
			Buckets:   []float64{1, 50, 100, 500, 1000, 10000},
		}, []string{
			"path",
			"method",
			"status",
		}),
	}
}

func (s *Stats) Listen(address string) {
	s.listenAndServerSync.Do(func() {
		s.server = &http.Server{
			Addr:    address,
			Handler: s.mux,
		}
		go s.server.ListenAndServe()
	})
}

func (s *Stats) Close() error {
	if s.server != nil {
		return s.server.Close()
	}

	return nil
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

func (s *Stats) FinishedRequest(ctx *context.Context, responseTime time.Duration) {
	s.HTTPResponseTime.With(prometheus.Labels{
		"path":   strings.TrimPrefix(ctx.RouteName(), ctx.Method()),
		"method": ctx.Method(),
		"status": strconv.FormatInt(int64(ctx.GetStatusCode()), 10),
	}).Observe(float64(responseTime.Milliseconds()))
}
