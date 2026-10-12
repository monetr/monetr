package metrics

import (
	"net/http"
	"sync"
	"time"

	"github.com/monetr/monetr/server/build"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	// Namespace is prefixed onto every metric monetr exposes. Metrics themselves
	// are declared as package level vars in the package that uses them.
	Namespace = "monetr"
)

var (
	buildInfo = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace:   Namespace,
		Name:        "build_info",
		Help:        "Always 1, the labels have the version of monetr that is running.",
		ConstLabels: map[string]string{},
	}, []string{
		"version",
		"revision",
		"build_type",
	})
)

type Stats struct {
	mux                 *http.ServeMux
	server              *http.Server
	listenAndServerSync sync.Once
}

func NewStats() *Stats {
	// The build variables are set in main, not at init, so this can't be a const
	// label on the metric itself.
	buildInfo.With(prometheus.Labels{
		"version":    build.Release,
		"revision":   build.Revision,
		"build_type": build.BuildType,
	}).Set(1)

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	return &Stats{
		mux: mux,
	}
}

func (s *Stats) Listen(address string) {
	s.listenAndServerSync.Do(func() {
		s.server = &http.Server{
			Addr:              address,
			Handler:           s.mux,
			ReadTimeout:       30 * time.Second,
			ReadHeaderTimeout: 30 * time.Second,
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
