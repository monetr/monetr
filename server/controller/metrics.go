package controller

import (
	"github.com/monetr/monetr/server/consts"
	"github.com/monetr/monetr/server/metrics"
	. "github.com/monetr/monetr/server/models"
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

// Create the metrics for every label value we know about ahead of time. A
// counter that only shows up on its first increment starts at 1, prometheus
// can't see that first increase so the event just gets lost. HTTP metrics are
// left out since seeding every route and status would be a ton of series.
func init() {
	for _, result := range []string{
		"success",
		"invalid_credentials",
		"email_not_verified",
		"password_change_required",
		"mfa_required",
		"error",
	} {
		logins.With(prometheus.Labels{
			"result": result,
		})
	}

	for _, result := range []string{
		"valid",
		"invalid",
	} {
		totpVerifications.With(prometheus.Labels{
			"result": result,
		})
	}

	for _, linkType := range []LinkType{
		ManualLinkType,
		PlaidLinkType,
		LunchFlowLinkType,
	} {
		labels := prometheus.Labels{
			"link_type": string(linkType),
		}
		linksCreated.With(labels)
		linksDeleted.With(labels)
	}

	for _, product := range consts.PlaidProductStrings() {
		plaidLinksCreated.With(prometheus.Labels{
			"product": product,
		})
	}

	// These are the webhooks we actually handle in processWebhook, anything else
	// Plaid sends us will still start at 1.
	for webhookType, webhookCodes := range map[string][]string{
		"TRANSACTIONS": {
			"SYNC_UPDATES_AVAILABLE",
			"INITIAL_UPDATE",
			"HISTORICAL_UPDATE",
			"RECURRING_TRANSACTIONS_UPDATE",
		},
		"ITEM": {
			"LOGIN_REPAIRED",
			"ERROR",
			"PENDING_EXPIRATION",
			"USER_PERMISSION_REVOKED",
			"USER_ACCOUNT_REVOKED",
			"WEBHOOK_UPDATE_ACKNOWLEDGED",
			"NEW_ACCOUNTS_AVAILABLE",
		},
	} {
		for _, webhookCode := range webhookCodes {
			plaidWebhooks.With(prometheus.Labels{
				"webhook_type": webhookType,
				"webhook_code": webhookCode,
			})
		}
	}

	// Transaction uploads only accept OFX files and imports only accept CSV.
	for kind, contentType := range map[string]ContentType{
		TransactionUpload{}.FileKind(): IntuitQFXContentType,
		TransactionImport{}.FileKind(): TextCSVContentType,
	} {
		labels := prometheus.Labels{
			"kind":         kind,
			"content_type": string(contentType),
		}
		fileUploads.With(labels)
		fileUploadSize.With(labels)
	}
}
