//+build !ui

package cmd

import (
	"github.com/go-pg/pg/v10"
	"github.com/gomodule/redigo/redis"
	"github.com/monetrapp/rest-api/pkg/application"
	"github.com/monetrapp/rest-api/pkg/config"
	"github.com/monetrapp/rest-api/pkg/controller"
	"github.com/monetrapp/rest-api/pkg/internal/plaid_helper"
	"github.com/monetrapp/rest-api/pkg/jobs"
	"github.com/monetrapp/rest-api/pkg/metrics"
	"github.com/monetrapp/rest-api/pkg/secrets"
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
		),
	}
}
