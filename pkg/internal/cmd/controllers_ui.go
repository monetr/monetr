//go:build ui
// +build ui

package cmd

import (
	"github.com/go-pg/pg/v10"
	"github.com/gomodule/redigo/redis"
	"github.com/monetr/rest-api/pkg/application"
	"github.com/monetr/rest-api/pkg/billing"
	"github.com/monetr/rest-api/pkg/config"
	"github.com/monetr/rest-api/pkg/controller"
	"github.com/monetr/rest-api/pkg/internal/platypus"
	"github.com/monetr/rest-api/pkg/internal/stripe_helper"
	"github.com/monetr/rest-api/pkg/jobs"
	"github.com/monetr/rest-api/pkg/mail"
	"github.com/monetr/rest-api/pkg/metrics"
	"github.com/monetr/rest-api/pkg/secrets"
	"github.com/monetr/rest-api/pkg/ui"
	"github.com/sirupsen/logrus"
)

func getControllers(
	log *logrus.Entry,
	configuration config.Configuration,
	db *pg.DB,
	job jobs.JobManager,
	plaidClient platypus.Platypus,
	stats *metrics.Stats,
	stripe stripe_helper.Stripe,
	cache *redis.Pool,
	plaidSecrets secrets.PlaidSecretsProvider,
	basicPaywall billing.BasicPayWall,
	smtpCommunication mail.Communication,
) []application.Controller {
	return []application.Controller{
		controller.NewController(
			log,
			configuration,
			db,
			job,
			plaidClient,
			stats,
			stripe,
			cache,
			plaidSecrets,
			basicPaywall,
			smtpCommunication,
		),
		ui.NewUIController(),
	}
}
