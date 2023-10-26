package main

import (
	"github.com/go-pg/pg/v10"
	"github.com/gomodule/redigo/redis"
	"github.com/monetr/monetr/server/application"
	"github.com/monetr/monetr/server/background"
	"github.com/monetr/monetr/server/billing"
	"github.com/monetr/monetr/server/communication"
	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/controller"
	"github.com/monetr/monetr/server/metrics"
	"github.com/monetr/monetr/server/platypus"
	"github.com/monetr/monetr/server/secrets"
	"github.com/monetr/monetr/server/stripe_helper"
	"github.com/monetr/monetr/server/ui"
	"github.com/sirupsen/logrus"
)

func getControllers(
	log *logrus.Entry,
	configuration config.Configuration,
	db *pg.DB,
	backgroundJobs *background.BackgroundJobs,
	plaidClient platypus.Platypus,
	stats *metrics.Stats,
	stripe stripe_helper.Stripe,
	cache *redis.Pool,
	plaidSecrets secrets.PlaidSecretsProvider,
	basicPaywall billing.BasicPayWall,
	email communication.EmailCommunication,
) []application.Controller {
	return []application.Controller{
		controller.NewController(
			log,
			configuration,
			db,
			backgroundJobs,
			plaidClient,
			stats,
			stripe,
			cache,
			plaidSecrets,
			basicPaywall,
			email,
		),
		ui.NewUIController(log, configuration),
	}
}
