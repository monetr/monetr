package metrics

import (
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	// Namespace is prefixed onto every metric monetr exposes. Metrics themselves
	// are declared as package level vars in the package that uses them.
	Namespace = "monetr"
)

type Stats struct {
	mux                 *http.ServeMux
	server              *http.Server
	listenAndServerSync sync.Once
}

func NewStats() *Stats {
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
