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

	// Users and authentication
	logins = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace:   metrics.Namespace,
		Name:        "logins",
		Help:        "Number of login attempts, labeled by the result (success, invalid_credentials, email_not_verified, password_change_required, mfa_required, error). Attempts that fail proof of work are not counted here.",
		ConstLabels: map[string]string{},
	}, []string{
		"result",
	})
	usersCreated = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: metrics.Namespace,
		Name:      "users_created",
		Help:      "Number of users that have registered.",
	})
	emailsVerified = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: metrics.Namespace,
		Name:      "emails_verified",
		Help:      "Number of email addresses that have been verified.",
	})
	passwordResets = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: metrics.Namespace,
		Name:      "password_resets",
		Help:      "Number of passwords that have been reset using a reset token.",
	})
	totpEnabled = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: metrics.Namespace,
		Name:      "totp_enabled",
		Help:      "Number of times TOTP has been enabled on a login.",
	})
	totpVerifications = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace:   metrics.Namespace,
		Name:        "totp_verifications",
		Help:        "Number of TOTP codes checked during login, labeled by whether the code was valid or invalid.",
		ConstLabels: map[string]string{},
	}, []string{
		"result",
	})

	// Links
	linksCreated = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace:   metrics.Namespace,
		Name:        "links_created",
		Help:        "Number of links created, labeled by the type of link (manual, plaid, lunch_flow).",
		ConstLabels: map[string]string{},
	}, []string{
		"link_type",
	})
	linksDeleted = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace:   metrics.Namespace,
		Name:        "links_deleted",
		Help:        "Number of links deleted, labeled by the type of link (manual, plaid, lunch_flow).",
		ConstLabels: map[string]string{},
	}, []string{
		"link_type",
	})
	plaidLinksCreated = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace:   metrics.Namespace,
		Name:        "plaid_links_created",
		Help:        "Number of Plaid links created, labeled by product. A link with multiple products counts once for each product.",
		ConstLabels: map[string]string{},
	}, []string{
		"product",
	})
	plaidWebhooks = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace:   metrics.Namespace,
		Name:        "plaid_webhooks",
		Help:        "Number of verified webhooks received from Plaid, labeled by the webhook type and code.",
		ConstLabels: map[string]string{},
	}, []string{
		"webhook_type",
		"webhook_code",
	})

	// Spending and transactions
	spendingCreated = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: metrics.Namespace,
		Name:      "spending_created",
		Help:      "Number of spending objects (expenses and goals) created.",
	})
	spendingDeleted = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: metrics.Namespace,
		Name:      "spending_deleted",
		Help:      "Number of spending objects (expenses and goals) deleted.",
	})
	transactionsCreated = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: metrics.Namespace,
		Name:      "transactions_created",
		Help:      "Number of transactions created through the API. Does not include transactions from Plaid, Lunch Flow or file imports.",
	})
	transactionsUpdated = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: metrics.Namespace,
		Name:      "transactions_updated",
		Help:      "Number of transactions updated through the API. Does not include updates from Plaid, Lunch Flow or file imports.",
	})
	transactionsDeleted = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: metrics.Namespace,
		Name:      "transactions_deleted",
		Help:      "Number of transactions deleted through the API. Does not include deletes from Plaid, Lunch Flow or file imports.",
	})

	// File uploads
	fileUploads = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace:   metrics.Namespace,
		Name:        "file_uploads",
		Help:        "Number of files uploaded, labeled by what the file is for and its content type.",
		ConstLabels: map[string]string{},
	}, []string{
		"kind",
		"content_type",
	})
	fileUploadSize = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: metrics.Namespace,
		Name:      "file_upload_size",
		Help:      "Size of uploaded files, in bytes.",
		// 1KB, 10KB, 100KB, 1MB, 10MB, 100MB
		Buckets: prometheus.ExponentialBuckets(1024, 10, 6),
	}, []string{
		"kind",
		"content_type",
	})
)
