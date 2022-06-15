package main

import (
	"context"
	"encoding/hex"
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
	"github.com/sirupsen/logrus"
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
	newViewSecretCommand(command)

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

			var kms secrets.KeyManagement
			if configuration.KeyManagement.Enabled {
				if kmsConfig := configuration.KeyManagement.AWS; kmsConfig != nil {
					kms, err = secrets.NewAWSKMS(cmd.Context(), secrets.AWSKMSConfig{
						Log:       log,
						KeyID:     kmsConfig.KeyID,
						Region:    kmsConfig.Region,
						AccessKey: kmsConfig.AccessKey,
						SecretKey: kmsConfig.SecretKey,
						Endpoint:  kmsConfig.Endpoint,
					})
					if err != nil {
						log.WithError(err).Fatalf("failed to init AWS KMS client")
						return err
					}
				}
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
				"Re-Encryption Status",
			})

			secretsProvider := secrets.NewVaultPlaidSecretsProvider(log, vault)

			for _, link := range links {
				vaultStatus := "unknown"
				var accessToken string
				if vault != nil {
					ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
					accessToken, err = secretsProvider.GetAccessTokenForPlaidLinkId(ctx, link.AccountId, link.PlaidLink.ItemId)
					cancel()
					if err != nil {
						vaultStatus = errors.Wrap(err, "failed to read secret from vault").Error()
					} else if accessToken == "" {
						vaultStatus = "blank key"
					} else {
						vaultStatus = "ok"
					}
				}

				reEncryptionStatus := "not re-encrypted"
				if accessToken == "" {
					reEncryptionStatus = "no secret to re-encrypt"
				} else if kms != nil {
					keyId, _, _, err := kms.Encrypt(cmd.Context(), []byte(accessToken))
					if err != nil {
						reEncryptionStatus = errors.Wrap(err, "failed to re-encrypt secret").Error()
					} else {
						reEncryptionStatus = fmt.Sprintf("successfully re-encrypted using: %s", keyId)
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
					reEncryptionStatus,
				})
			}
			table.Render()

			return nil
		},
	}

	parent.AddCommand(command)
}

func newViewSecretCommand(parent *cobra.Command) {
	var itemId string

	command := &cobra.Command{
		Use:   "get",
		Short: "Retrieve a secret's value from the data store, meant for debugging purposes only!",
		RunE: func(cmd *cobra.Command, args []string) error {
			configuration := config.LoadConfiguration()
			log := logging.NewLoggerWithConfig(configuration.Logging)

			db, err := getDatabase(log, configuration, nil)
			if err != nil {
				log.WithError(err).Fatalf("failed to initialze database")
				return errors.Wrap(err, "failed to initialize database")
			}

			kms, err := getKMS(log, configuration)
			if err != nil {
				log.WithError(err).Fatal("failed to setup KMS")
				return err
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

				log.Fatalf("cannot read secrets from vault at this time")
				return nil
			}

			var token models.PlaidToken
			err = db.Model(&token).
				Where(`"item_id" = ?`, itemId).
				Limit(1).
				Select(&token)
			if err != nil {
				log.WithError(err).Fatalf("failed to retrieve secret")
				return errors.Wrap(err, "failed to retrieve secret")
			}

			if token.KeyID == nil {
				fmt.Println(token.AccessToken)
				return nil
			}

			version := ""
			if token.Version != nil && *token.Version != "" {
				version = *token.Version
			}
			decoded, err := hex.DecodeString(token.AccessToken)
			if err != nil {
				log.WithError(err).Fatal("failed to decode secret")
				return err
			}
			decrypted, err := kms.Decrypt(cmd.Context(), *token.KeyID, version, decoded)
			if err != nil {
				log.WithError(err).Fatal("failed to decrypt secret")
				return err
			}


			fmt.Println(string(decrypted))
			return nil
		},
	}
	command.PersistentFlags().StringVar(&itemId, "item-id", "", "The Plaid Item ID to retrieve the secret for.")

	parent.AddCommand(command)
}

func newSecretMigrationCommand(parent *cobra.Command) {
	var dryRun bool

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
			configuration := config.LoadConfiguration()
			log := logging.NewLoggerWithConfig(configuration.Logging)

			db, err := getDatabase(log, configuration, nil)
			if err != nil {
				log.WithError(err).Fatalf("failed to initialze database")
				return errors.Wrap(err, "failed to initialize database")
			}

			kms, err := getKMS(log, configuration)
			if err != nil {
				log.WithError(err).Fatal("failed to setup KMS")
				return err
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
				"Institution ID",
				"Institution Name",
				"Vault Path",
				"Status",
			})

			secretsProvider := secrets.NewVaultPlaidSecretsProvider(log, vault)

			if vault == nil {
				log.Fatalf("must have vault enabled to perform vault migration")
				return errors.New("must have vault enabled")
			}

			txn, err := db.Begin()
			if err != nil {
				log.WithError(err).Fatalf("failed to begin transaction")
			}

			for _, link := range links {
				itemLog := log.WithFields(logrus.Fields{
					"accountId": link.AccountId,
					"itemId":    link.PlaidLink.ItemId,
					"linkId":    link.LinkId,
				})
				accessToken, err := secretsProvider.GetAccessTokenForPlaidLinkId(context.Background(), link.AccountId, link.PlaidLink.ItemId)
				if err != nil {
					if dryRun {
						itemLog.WithError(err).Error("failed to migrate secret")
						table.Append([]string{
							strconv.FormatUint(link.PlaidLink.PlaidLinkID, 10),
							link.PlaidLink.ItemId,
							link.PlaidLink.InstitutionId,
							link.PlaidLink.InstitutionName,
							fmt.Sprintf("customers/plaid/data/%d/%s", link.AccountId, link.PlaidLink.ItemId),
							"failed to migrate secret",
						})
						continue
					}
					itemLog.WithError(err).Fatalf("failed to migrate secret")
					return err
				}

				item := models.PlaidToken{
					ItemId:      link.PlaidLink.ItemId,
					AccountId:   link.AccountId,
					KeyID:       nil,
					Version:     nil,
					AccessToken: "",
				}

				if kms == nil {
					itemLog.Info("no KMS provided, secret will be stored in plain text")
					item.AccessToken = accessToken
				} else {
					itemLog.Info("using provided KMS to re-encrypt secret")
					keyId, version, encrypted, err := kms.Encrypt(context.Background(), []byte(accessToken))
					if err != nil {
						if dryRun {
							itemLog.WithError(err).Error("failed to re-encrypt secret")
							table.Append([]string{
								strconv.FormatUint(link.PlaidLink.PlaidLinkID, 10),
								link.PlaidLink.ItemId,
								link.PlaidLink.InstitutionId,
								link.PlaidLink.InstitutionName,
								fmt.Sprintf("customers/plaid/data/%d/%s", link.AccountId, link.PlaidLink.ItemId),
								"failed to re-encrypt secret",
							})
							continue
						}

						itemLog.WithError(err).Fatalf("failed to re-encrypt secret")
						return err
					}

					item.KeyID = &keyId
					if version != "" {
						item.Version = &version
					}
					item.AccessToken = hex.EncodeToString(encrypted)
				}

				status := "successful"

				result, err := txn.Model(&item).Insert(&item)
				if err != nil {
					if dryRun {
						log.WithError(err).Error("failed to store migrated secret")
						status = "failed"
					} else {
						log.WithError(err).Fatalf("failed to store migrated secret")
						return err
					}
				} else if result.RowsAffected() != 1 {
					status = "failed - unknown"
				}

				if dryRun {
					status += " (dry run)"
				}

				table.Append([]string{
					strconv.FormatUint(link.PlaidLink.PlaidLinkID, 10),
					link.PlaidLink.ItemId,
					link.PlaidLink.InstitutionId,
					link.PlaidLink.InstitutionName,
					fmt.Sprintf("customers/plaid/data/%d/%s", link.AccountId, link.PlaidLink.ItemId),
					status,
				})
			}
			if dryRun {
				if err = txn.Rollback(); err != nil {
					log.WithError(err).Fatalf("failed to rollback changes")
					return err
				}
			} else {
				if err = txn.Commit(); err != nil {
					log.WithError(err).Fatalf("failed to commit changes")
					return err
				}
			}

			table.Render()

			return nil
		},
	}
	command.PersistentFlags().BoolVar(&dryRun, "dry-run", true, "Test the changes that would be made without making them.")

	parent.AddCommand(command)
}
