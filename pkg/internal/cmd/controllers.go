//+build !ui

package cmd

import (
	"github.com/go-pg/pg/v10"
	"github.com/monetrapp/rest-api/pkg/application"
	"github.com/monetrapp/rest-api/pkg/config"
	"github.com/monetrapp/rest-api/pkg/controller"
	"github.com/monetrapp/rest-api/pkg/jobs"
	"github.com/monetrapp/rest-api/pkg/metrics"
	"github.com/plaid/plaid-go/plaid"
	stripe_client "github.com/stripe/stripe-go/v72/client"
)

func getControllers(
	configuration config.Configuration,
	db *pg.DB,
	job jobs.JobManager,
	plaidClient *plaid.Client,
	stats *metrics.Stats,
	stripeClient *stripe_client.API,
) []application.Controller {
	return []application.Controller{
		controller.NewController(configuration, db, job, plaidClient, stats, stripeClient),
	}
}
