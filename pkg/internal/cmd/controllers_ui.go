//+build ui

package cmd

import (
	"github.com/go-pg/pg/v10"
	"github.com/gomodule/redigo/redis"
	"github.com/monetr/rest-api/pkg/application"
	"github.com/monetr/rest-api/pkg/config"
	"github.com/monetr/rest-api/pkg/controller"
	"github.com/monetr/rest-api/pkg/internal/plaid_helper"
	"github.com/monetr/rest-api/pkg/jobs"
	"github.com/monetr/rest-api/pkg/metrics"
	"github.com/monetr/rest-api/pkg/ui"
	"github.com/sirupsen/logrus"
	stripe_client "github.com/stripe/stripe-go/v72/client"
)

func getControllers(
	log *logrus.Entry,
	configuration config.Configuration,
	db *pg.DB,
	job jobs.JobManager,
	plaidClient plaid_helper.Client,
	stats *metrics.Stats,
	stripeClient *stripe_client.API,
	cache *redis.Pool,
	plaidSecrets secrets.PlaidSecretsProvider,
	basicPaywall billing.BasicPayWall,
) []application.Controller {
	return []application.Controller{
		controller.NewController(
			log,
			configuration,
			db,
			job,
			plaidClient,
			stats,
			stripeClient,
			cache,
			plaidSecrets,
			basicPaywall,
		),
		ui.NewUIController(),
	}
}
