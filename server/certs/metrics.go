package certs

import (
	"github.com/monetr/monetr/server/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	certificateRotations = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace:   metrics.Namespace,
		Name:        "certificate_rotations",
		Help:        "Number of times a certificate or key file changed on disk and was reloaded.",
		ConstLabels: map[string]string{},
	}, []string{
		"path",
	})
	certificateReloadFailures = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace:   metrics.Namespace,
		Name:        "certificate_reload_failures",
		Help:        "Number of times a certificate or key file failed to load. The previously loaded certificates are kept when this happens.",
		ConstLabels: map[string]string{},
	}, []string{
		"path",
	})
	certificateNotAfter = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace:   metrics.Namespace,
		Name:        "certificate_not_after",
		Help:        "Unix timestamp in seconds of when the currently loaded certificate expires. For a CA file with multiple certificates this is the one that expires first.",
		ConstLabels: map[string]string{},
	}, []string{
		"path",
	})
)
