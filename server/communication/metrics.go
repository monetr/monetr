package communication

import (
	"github.com/monetr/monetr/server/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	emailsSent = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace:   metrics.Namespace,
		Name:        "emails_sent",
		Help:        "Number of emails sent successfully, labeled by the email template.",
		ConstLabels: map[string]string{},
	}, []string{
		"template",
	})
	emailFailures = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace:   metrics.Namespace,
		Name:        "email_failures",
		Help:        "Number of emails that failed to send, labeled by the email template.",
		ConstLabels: map[string]string{},
	}, []string{
		"template",
	})
)

// Create the metrics for every template up front so they start at 0, otherwise
// prometheus can't see the first email that gets sent.
func init() {
	for _, template := range []string{
		VerifyEmailParams{}.Template(),
		PasswordResetParams{}.Template(),
		PasswordChangedParams{}.Template(),
		PlaidDisconnectedParams{}.Template(),
		TrialAboutToExpireParams{}.Template(),
		ApiKeyCreatedParams{}.Template(),
	} {
		labels := prometheus.Labels{
			"template": template,
		}
		emailsSent.With(labels)
		emailFailures.With(labels)
	}
}
