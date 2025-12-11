package commands

import (
	"context"

	"github.com/benbjohnson/clock"
	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/database"
	"github.com/monetr/monetr/server/logging"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/platypus"
	"github.com/monetr/monetr/server/secrets"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func adminPlaidUpdateWebhook(parent *cobra.Command) {
	var arguments struct {
		LinkID string
	}

	command := &cobra.Command{
		Use:   "plaid:update-webhook",
		Short: "Update the Plaid webhook URL for a Plaid link",
		RunE: func(cmd *cobra.Command, args []string) error {
			clock := clock.New()
			configuration := config.LoadConfiguration()

			log := logging.NewLoggerWithConfig(configuration.Logging)
			if configFileName := configuration.GetConfigFileName(); configFileName != "" {
				log.WithField("config", configFileName).Info("config file loaded")
			}

			if arguments.LinkID == "" {
				log.Fatal("link ID must be specified via --link")
				return cmd.Help()
			}

			db, err := database.GetDatabase(log, configuration, nil)
			if err != nil {
				log.WithError(err).Fatal("failed to setup database")
				return err
			}

			kms, err := secrets.GetKMS(log, configuration)
			if err != nil {
				log.WithError(err).Fatal("failed to initialize KMS")
				return err
			}

			log.Info("retrieving link from database")
			var link models.Link
			if err := db.Model(&link).
				Relation("PlaidLink").
				Where(`"link"."link_id" = ?`, arguments.LinkID).
				Limit(1).
				Select(&link); err != nil {
				log.WithError(err).Fatal("failed to retrieve link specified")
				return err
			}

			if link.PlaidLink == nil {
				log.Fatal("link does not have a plaid link!")
				return errors.New("link is not a valid plaid link")
			}

			plaid := platypus.NewPlaid(
				log,
				clock,
				kms,
				db,
				configuration.Plaid,
			)

			client, err := plaid.NewClientFromLink(
				context.Background(),
				link.AccountId,
				link.LinkId,
			)
			if err != nil {
				log.WithError(err).Warn("failed to create Plaid client")
				return err
			}

			if err := client.UpdateWebhook(cmd.Context()); err != nil {
				log.WithError(err).Fatal("failed to update plaid webhook url")
				return err
			}

			log.Info("plaid webhook url updated successfully!")

			return nil
		},
	}

	command.PersistentFlags().StringVar(&arguments.LinkID, "link", "", "Link Id to update the webhook URL of in Plaid")

	parent.AddCommand(command)
}
