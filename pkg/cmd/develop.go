//go:build development

package main

import (
	"context"

	"github.com/monetr/monetr/pkg/config"
	"github.com/monetr/monetr/pkg/logging"
	"github.com/monetr/monetr/pkg/models"
	"github.com/monetr/monetr/pkg/platypus"
	"github.com/monetr/monetr/pkg/repository"
	"github.com/monetr/monetr/pkg/secrets"
	"github.com/monetr/monetr/pkg/vault_helper"
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

	cleanPlaid := &cobra.Command{
		Use:   "clean:plaid",
		Short: "Remove all Plaid links currently configured in the development environment.",
		RunE: func(cmd *cobra.Command, args []string) error {
			configuration := config.LoadConfiguration()

			log := logging.NewLoggerWithConfig(configuration.Logging)
			if configFileName := configuration.GetConfigFileName(); configFileName != "" {
				log.WithField("config", configFileName).Info("config file loaded")
			}

			var vault vault_helper.VaultHelper
			if configuration.Vault.Enabled {
				log.Debug("vault is enabled for secret storage")
				client, err := vault_helper.NewVaultHelper(log, vault_helper.Config{
					Address:         configuration.Vault.Address,
					Role:            configuration.Vault.Role,
					Auth:            configuration.Vault.Auth,
					Token:           configuration.Vault.Token,
					TokenFile:       configuration.Vault.TokenFile,
					Timeout:         configuration.Vault.Timeout,
					IdleConnTimeout: configuration.Vault.IdleConnTimeout,
					Username:        configuration.Vault.Username,
					Password:        configuration.Vault.Password,
				})
				if err != nil {
					log.WithError(err).Fatalf("failed to create vault helper")
					return err
				}

				vault = client
			}

			db, err := getDatabase(log, configuration, nil)
			if err != nil {
				log.WithError(err).Fatal("failed to setup database")
				return err
			}

			var plaidSecrets secrets.PlaidSecretsProvider
			if configuration.Vault.Enabled {
				log.Debugf("secrets will be stored in vault")
				plaidSecrets = secrets.NewVaultPlaidSecretsProvider(log, vault)
			} else {
				log.Debugf("secrets will be stored in postgres")
				plaidSecrets = secrets.NewPostgresPlaidSecretsProvider(log, db, nil)
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

	developCommand.AddCommand(cleanPlaid)
}
