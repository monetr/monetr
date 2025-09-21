//go:build development

package commands

import (
	"context"

	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/database"
	"github.com/monetr/monetr/server/logging"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/stripe_helper"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func developmentCleanStripe(parent *cobra.Command) {
	command := &cobra.Command{
		Use:   "clean:stripe",
		Short: "Delete all Stripe customers and subscriptions in the development environment.",
		RunE: func(cmd *cobra.Command, args []string) error {
			configuration := config.LoadConfiguration()

			log := logging.NewLoggerWithConfig(configuration.Logging)
			if configFileName := configuration.GetConfigFileName(); configFileName != "" {
				log.WithField("config", configFileName).Info("config file loaded")
			}

			db, err := database.GetDatabase(log, configuration, nil)
			if err != nil {
				log.WithError(err).Fatal("failed to setup database")
				return err
			}

			log.Info("retrieving accounts with stripe details from the database")
			var stripeItems []models.Account
			db.Model(&stripeItems).
				WhereOr(`"account"."stripe_customer_id" IS NOT NULL`).
				WhereOr(`"account"."stripe_subscription_id" IS NOT NULL`).
				Select(&stripeItems)

			if len(stripeItems) == 0 {
				log.Info("no Stripe customers or subscriptions to clean up")
				return nil
			}

			log.WithField("count", len(stripeItems)).Info("found Stripe item(s)")

			// TODO Remove the items from stripe!
			stripe := stripe_helper.NewStripeHelper(
				log,
				configuration.Stripe.APIKey,
			)

			for _, item := range stripeItems {
				itemLog := log.WithFields(logrus.Fields{
					"stripeCustomerId": item.StripeCustomerId,
				})
				if item.StripeSubscriptionId != nil {
					itemLog = itemLog.WithField("stripeSubscriptionId", item.StripeSubscriptionId)
					itemLog.Info("removing subscription")

					if err := stripe.CancelSubscription(context.Background(), *item.StripeSubscriptionId); err != nil {
						itemLog.WithError(err).Warn("failed to cancel subscription")
					}
				}

				if item.StripeCustomerId != nil {
					itemLog.Info("removing customer")

					if err := stripe.RemoveCustomer(context.Background(), *item.StripeCustomerId); err != nil {
						itemLog.WithError(err).Warn("failed to remove stripe customer")
					}
				}
			}

			return nil
		},
	}

	parent.AddCommand(command)
}
