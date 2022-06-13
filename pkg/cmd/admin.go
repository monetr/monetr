package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/monetr/monetr/pkg/config"
	"github.com/monetr/monetr/pkg/logging"
	"github.com/monetr/monetr/pkg/models"
	"github.com/monetr/monetr/pkg/secrets"
	"github.com/monetr/monetr/pkg/vault_helper"
	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func init() {
	newAdminCommand(rootCommand)
}

func newAdminCommand(parent *cobra.Command) {
	command := &cobra.Command{
		Use:   "admin",
		Short: "General administrative tasks for hosting/maintaining monetr",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	newSecretsCommand(command)

	parent.AddCommand(command)
}

func newSecretsCommand(parent *cobra.Command) {
	command := &cobra.Command{
		Use:   "secrets",
		Short: "Manage secrets within monetr.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	newSecretInformationCommand(command)
	newSecretMigrationCommand(command)

	parent.AddCommand(command)
}

func newSecretInformationCommand(parent *cobra.Command) {
	command := &cobra.Command{
		Use:   "info",
		Short: "Display information about secrets and how they are stored.",
		RunE: func(cmd *cobra.Command, args []string) error {
			configuration := config.LoadConfiguration()
			log := logging.NewLoggerWithConfig(configuration.Logging)

			db, err := getDatabase(log, configuration, nil)
			if err != nil {
				log.WithError(err).Fatalf("failed to initialze database")
				return errors.Wrap(err, "failed to initialize database")
			}

			var vault vault_helper.VaultHelper
			if configuration.Vault.Enabled {
				vault, err = vault_helper.NewVaultHelper(log, vault_helper.Config{
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
					log.WithError(err).Fatalf("failed to initialize vault client")
					return errors.Wrap(err, "failed to initialize vault client")
				}
				defer vault.Close()
			}

			links := make([]models.Link, 0)
			err = db.Model(&links).
				Relation("PlaidLink").
				Join(`LEFT JOIN "plaid_tokens" AS "plaid_token"`).
				JoinOn(`"plaid_link"."item_id" = "plaid_token"."item_id"`).
				Where(`"plaid_token"."access_token" IS NULL`).
				Where(`"plaid_link"."item_id" IS NOT NULL`).
				Select(&links)
			if err != nil {
				log.WithError(err).Fatalf("failed to retrieve links that are missing their plaid token in postgres")
				return errors.Wrap(err, "failed to retrieve plaid links for information")
			}
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{
				"Plaid Link ID (monetr's ID)",
				"Item ID (Plaid's ID)",
				"Products",
				"Webhook URL",
				"Institution ID",
				"Institution Name",
				"Vault Path",
				"Status",
			})

			secretsProvider := secrets.NewVaultPlaidSecretsProvider(log, vault)

			for _, link := range links {
				vaultStatus := "unknown"
				if vault != nil {
					ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
					accessToken, err := secretsProvider.GetAccessTokenForPlaidLinkId(ctx, link.AccountId, link.PlaidLink.ItemId)
					cancel()
					if err != nil {
						vaultStatus = errors.Wrap(err, "failed to read secret from vault").Error()
					} else if accessToken == "" {
						vaultStatus = "blank key"
					} else {
						vaultStatus = "ok"
					}
				}

				table.Append([]string{
					strconv.FormatUint(link.PlaidLink.PlaidLinkID, 10),
					link.PlaidLink.ItemId,
					strings.Join(link.PlaidLink.Products, ", "),
					link.PlaidLink.WebhookUrl,
					link.PlaidLink.InstitutionId,
					link.PlaidLink.InstitutionName,
					fmt.Sprintf("customers/plaid/data/%d/%s", link.AccountId, link.PlaidLink.ItemId),
					vaultStatus,
				})
			}
			table.Render()

			return nil
		},
	}

	parent.AddCommand(command)
}

func newSecretMigrationCommand(parent *cobra.Command) {
	command := &cobra.Command{
		Use:   "vault:migrate",
		Short: "Migrate secrets from Vault to Postgres.",
		Long: "Migrate secrets currently stored in Vault (soon to be deprecated) into Postgres.\n" +
			"If configured, this will also encrypt migrated secrets using the specified Key\n" +
			"Management System.\n\n" +
			"THIS DOES NOT MIGRATE SECRETS BETWEEN KEY MANAGEMENT SYSTEMS!!!\n\n" +
			"Please run:\n\n" +
			"\t$ monetr admin secrets info\n\n" +
			"Before running this command. This will tell you what secrets would be migrated.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()

		},
	}

	parent.AddCommand(command)
}
