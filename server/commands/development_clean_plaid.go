//go:build development

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
	"github.com/spf13/cobra"
)

func developmentCleanPlaid(parent *cobra.Command) {
	command := &cobra.Command{
		Use:   "clean:plaid",
		Short: "Remove all Plaid links currently configured in the development environment.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			clock := clock.New()
			configuration := config.LoadConfiguration()

			log := logging.NewLoggerWithConfig(configuration.Logging)
			if configFileName := configuration.GetConfigFileName(); configFileName != "" {
				log.Info("config file loaded", "config", configFileName)
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

			log.Info("found Plaid link(s)", "count", len(plaidLinks))

			plaid := platypus.NewPlaid(
				log,
				clock,
				kms,
				db,
				configuration.Plaid,
			)

			for _, link := range plaidLinks {
				client, err := plaid.NewClientFromLink(context.Background(), link.AccountId, link.LinkId)
				if err != nil {
					log.Warn("failed to create Plaid client", "err", err)
					continue
				}

				log.Info("removing item", "itemId", link.PlaidLink.PlaidId)
				if err = client.RemoveItem(context.Background()); err != nil {
					log.Warn("failed to remove item", "err", err)
					continue
				}

				db.Model(&link).Set(`"link_type" = ?`, models.ManualLinkType).Update(&link)
			}

			log.Info("done!")
			return nil
		},
	}

	parent.AddCommand(command)
}
