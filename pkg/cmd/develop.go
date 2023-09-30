//go:build development

package main

import (
	"context"

	"github.com/monetr/monetr/pkg/cache"
	"github.com/monetr/monetr/pkg/config"
	"github.com/monetr/monetr/pkg/logging"
	"github.com/monetr/monetr/pkg/models"
	"github.com/monetr/monetr/pkg/platypus"
	"github.com/monetr/monetr/pkg/repository"
	"github.com/monetr/monetr/pkg/secrets"
	"github.com/monetr/monetr/pkg/stripe_helper"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	newDevelopCommand(rootCommand)
}

// newDevelopCommand is just a place where some helpful local dev stuff can be put. Right now this is used to remove all
// plaid links when the development environment is shutdown. This makes sure that we don't have stuff lingering in
// plaid's sandbox.
func newDevelopCommand(parent *cobra.Command) {
	developCommand := &cobra.Command{
		Use:   "development",
		Short: "Development tools for working locally.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	parent.AddCommand(developCommand)

	cleanStripe := &cobra.Command{
		Use:   "clean:stripe",
		Short: "Delete all Stripe customers and subscriptions in the development environment.",
		RunE: func(cmd *cobra.Command, args []string) error {
			configuration := config.LoadConfiguration()

			log := logging.NewLoggerWithConfig(configuration.Logging)
			if configFileName := configuration.GetConfigFileName(); configFileName != "" {
				log.WithField("config", configFileName).Info("config file loaded")
			}

			db, err := getDatabase(log, configuration, nil)
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

	cacheFlush := &cobra.Command{
		Use:   "cache:flush",
		Short: "Flush all data from the Redis cache server.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			configuration := config.LoadConfiguration()

			log := logging.NewLoggerWithConfig(configuration.Logging)
			if configFileName := configuration.GetConfigFileName(); configFileName != "" {
				log.WithField("config", configFileName).Info("config file loaded")
			}

			redisController, err := cache.NewRedisCache(log, configuration.Redis)
			if err != nil {
				log.WithError(err).Fatalf("failed to create redis cache: %+v", err)
				return err
			}
			defer redisController.Close()

			conn, err := redisController.Pool().Dial()
			if err != nil {
				log.WithError(err).Fatalf("failed to retrieve connection from redis pool: %+v", err)
				return err
			}

			if err := conn.Send("FLUSHALL"); err != nil {
				log.WithError(err).Fatalf("failed to flush redis cache: %+v", err)
				return err
			}

			log.Info("done!")
			return nil
		},
	}

	cleanPlaid := &cobra.Command{
		Use:   "clean:plaid",
		Short: "Remove all Plaid links currently configured in the development environment.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			configuration := config.LoadConfiguration()

			log := logging.NewLoggerWithConfig(configuration.Logging)
			if configFileName := configuration.GetConfigFileName(); configFileName != "" {
				log.WithField("config", configFileName).Info("config file loaded")
			}

			db, err := getDatabase(log, configuration, nil)
			if err != nil {
				log.WithError(err).Fatal("failed to setup database")
				return err
			}

			kms, err := getKMS(log, configuration)
			if err != nil {
				log.WithError(err).Fatal("failed to initialize KMS")
				return err
			}

			plaidSecrets := secrets.NewPostgresPlaidSecretsProvider(log, db, kms)

			log.Info("retrieving Plaid links from the database")
			var plaidLinks []models.Link
			db.Model(&plaidLinks).
				Relation("PlaidLink").
				Where(`"link"."link_type" = ?`, models.PlaidLinkType).
				Select(&plaidLinks)
			if len(plaidLinks) == 0 {
				log.Info("no Plaid links to clean up")
				return nil
			}

			log.WithField("count", len(plaidLinks)).Info("found Plaid link(s)")

			plaid := platypus.NewPlaid(log, plaidSecrets, repository.NewPlaidRepository(db), configuration.Plaid)

			for _, link := range plaidLinks {
				accessToken, err := plaidSecrets.GetAccessTokenForPlaidLinkId(context.Background(), link.AccountId, link.PlaidLink.ItemId)
				if err != nil {
					log.WithError(err).Warn("failed to retrieve access token for link")
					continue
				}

				client, err := plaid.NewClient(context.Background(), &link, accessToken, link.PlaidLink.ItemId)
				if err != nil {
					log.WithError(err).Warn("failed to create Plaid client")
					continue
				}

				log.WithField("itemId", link.PlaidLink.ItemId).Info("removing item")
				if err = client.RemoveItem(context.Background()); err != nil {
					log.WithError(err).Warn("failed to remove item")
					continue
				}

				db.Model(&link).Set(`"link_type" = ?`, models.ManualLinkType).Update(&link)
			}

			log.Info("done!")
			return nil
		},
	}

	developCommand.AddCommand(cacheFlush)
	developCommand.AddCommand(cleanStripe)
	developCommand.AddCommand(cleanPlaid)
}
