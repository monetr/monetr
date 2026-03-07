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

func adminPlaidRefresh(parent *cobra.Command) {
	var arguments struct {
		LinkID string
	}
	command := &cobra.Command{
		Use:   "plaid:refresh",
		Short: "Trigger a transaction refresh for a Plaid link",
		RunE: func(cmd *cobra.Command, args []string) error {
			clock := clock.New()
			configuration := config.LoadConfiguration()

			log := logging.NewLoggerWithConfig(configuration.Logging)
			if configFileName := configuration.GetConfigFileName(); configFileName != "" {
				log.Info("config file loaded", "config", configFileName)
			}

			if arguments.LinkID == "" {
				log.Error("link ID must be specified via --link")
				return cmd.Help()
			}

			db, err := database.GetDatabase(log, configuration, nil)
			if err != nil {
				log.Error("failed to setup database", "err", err)
				return err
			}

			kms, err := secrets.GetKMS(log, configuration)
			if err != nil {
				log.Error("failed to initialize KMS", "err", err)
				return err
			}

			log.Info("retrieving link from database")
			var link models.Link
			if err := db.Model(&link).
				Relation("PlaidLink").
				Where(`"link"."link_id" = ?`, arguments.LinkID).
				Limit(1).
				Select(&link); err != nil {
				log.Error("failed to retrieve link specified", "err", err)
				return err
			}

			if link.PlaidLink == nil {
				log.Error("link does not have a plaid link!")
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
				log.Warn("failed to create Plaid client", "err", err)
				return err
			}

			log.Info("triggering transaction refresh")

			if err := client.RefeshTransactions(context.Background()); err != nil {
				log.Error("failed to refresh transactions", "err", err)
				return err
			}

			log.Info("transaction refresh triggered successfully!")

			return nil
		},
	}

	command.PersistentFlags().StringVar(&arguments.LinkID, "link", "", "Link Id to trigger the Plaid refresh on")

	parent.AddCommand(command)
}
