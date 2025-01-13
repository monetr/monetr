package main

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/database"
	"github.com/monetr/monetr/server/logging"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/platypus"
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
	newPlaidCommand(command)

	parent.AddCommand(command)
}

func newPlaidCommand(parent *cobra.Command) {
	command := &cobra.Command{
		Use:   "plaid",
		Short: "Manage Plaid links in monetr",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	newPlaidRefreshTransactionsCommand(command)

	parent.AddCommand(command)
}

func newPlaidRefreshTransactionsCommand(parent *cobra.Command) {
	var linkId string
	command := &cobra.Command{
		Use:   "refresh-transactions",
		Short: "Trigger a transaction refresh for a Plaid link",
		RunE: func(cmd *cobra.Command, args []string) error {
			clock := clock.New()
			configuration := config.LoadConfiguration()

			log := logging.NewLoggerWithConfig(configuration.Logging)
			if configFileName := configuration.GetConfigFileName(); configFileName != "" {
				log.WithField("config", configFileName).Info("config file loaded")
			}

			if linkId == "" {
				log.Fatal("link ID must be specified via --link")
				return cmd.Help()
			}

			db, err := database.GetDatabase(log, configuration, nil)
			if err != nil {
				log.WithError(err).Fatal("failed to setup database")
				return err
			}

			kms, err := getKMS(log, configuration)
			if err != nil {
				log.WithError(err).Fatal("failed to initialize KMS")
				return err
			}

			log.Info("retrieving link from database")
			var link models.Link
			if err := db.Model(&link).
				Relation("PlaidLink").
				Where(`"link"."link_id" = ?`, linkId).
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

			log.Info("triggering transaction refresh")

			if err := client.RefeshTransactions(context.Background()); err != nil {
				log.WithError(err).Fatal("failed to refresh transactions")
				return err
			}

			log.Info("transaction refresh triggered successfully!")

			return nil
		},
	}
	command.PersistentFlags().StringVar(&linkId, "link", "", "Link Id to trigger the Plaid refresh on")

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

	newViewSecretCommand(command)
	newTestKMSCommand(command)
	newMigrateKMSCommand(command)

	parent.AddCommand(command)
}

func newViewSecretCommand(parent *cobra.Command) {
	var itemId string
	var accountId string
	var kmsProvider string

	command := &cobra.Command{
		Use:   "get",
		Short: "Retrieve a secret's value from the data store, meant for debugging purposes only!",
		RunE: func(cmd *cobra.Command, args []string) error {
			configuration := config.LoadConfiguration()
			if kmsProvider != "" {
				configuration.KeyManagement.Provider = kmsProvider
			}

			log := logging.NewLoggerWithConfig(configuration.Logging)

			db, err := database.GetDatabase(log, configuration, nil)
			if err != nil {
				log.WithError(err).Fatalf("failed to initialze database")
				return errors.Wrap(err, "failed to initialize database")
			}

			kms, err := getKMS(log, configuration)
			if err != nil {
				log.WithError(err).Fatal("failed to setup KMS")
				return err
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

			decrypted, err := kms.Decrypt(cmd.Context(), token.KeyID, token.Version, token.AccessToken)
			if err != nil {
				log.WithError(err).Fatal("failed to decrypt secret")
				return err
			}

			fmt.Println(string(decrypted))
			return nil
		},
	}
	command.PersistentFlags().StringVar(&itemId, "item-id", "", "The Plaid Item ID to retrieve the secret for.")
	command.PersistentFlags().StringVar(&accountId, "account-id", "", "The Account ID that the Plaid item belongs to in monetr.")
	command.PersistentFlags().StringVar(&kmsProvider, "kms-provider", "", "Override the KMS provider setting in the config.")

	parent.AddCommand(command)
}

func newTestKMSCommand(parent *cobra.Command) {
	var kmsProvider string

	command := &cobra.Command{
		Use:   "test-kms",
		Short: "Tests the configured KMS provider to make sure data can be encrypted and decrypted",
		RunE: func(cmd *cobra.Command, args []string) error {
			configuration := config.LoadConfiguration()
			if kmsProvider != "" {
				configuration.KeyManagement.Provider = kmsProvider
			}

			log := logging.NewLoggerWithConfig(configuration.Logging)

			kms, err := getKMS(log, configuration)
			if err != nil {
				log.WithError(err).Fatal("failed to setup KMS")
				return err
			}

			if kms == nil {
				log.Warn("no KMS configured")
				return nil
			}

			testString := "Lorem ipsum dolor sit amet"
			fmt.Printf("Testing KMS with test string: %s\n", testString)

			keyId, version, result, err := kms.Encrypt(context.Background(), testString)
			if err != nil {
				log.WithError(err).Error("failed to encrypt test string")
				return err
			}

			fmt.Printf("Successfully encrypted test string!\n")
			fmt.Printf("Key ID: %+v\n", keyId)
			fmt.Printf("Version: %+v\n", version)
			fmt.Printf("Result Binary -----------\n")
			fmt.Println()
			fmt.Println(hex.Dump([]byte(result)))
			fmt.Println()
			fmt.Printf("Result Formatted: %s\n", result)
			fmt.Println()
			fmt.Println("Testing decryption with test string...")

			decrypted, err := kms.Decrypt(context.Background(), keyId, version, result)
			if err != nil {
				log.WithError(err).Error("failed to dencrypt test string")
				return err
			}

			fmt.Printf("Successfully dencrypted test string!\n")

			if !bytes.Equal([]byte(testString), []byte(decrypted)) {
				log.Error("Input string and result do not match!")
				fmt.Println("Input:", testString)
				fmt.Println("Output:", string(decrypted))
				return errors.New("input string and decrytped string do not match")
			}

			fmt.Println("Input and output match! Everything is working!")

			return nil
		},
	}

	command.PersistentFlags().StringVar(&kmsProvider, "provider", "", "Specify the provider to use for KMS testing.")

	parent.AddCommand(command)
}

func newMigrateKMSCommand(parent *cobra.Command) {
	var fromKms string
	var toKms string
	var dryRun bool

	command := &cobra.Command{
		Use:   "migrate-kms",
		Short: "Migrate all stored secrets from one method of encryption to another.",
		Long:  "Migrate all stored secrets from one method of encryption to another. This can be used to go from plaintext secret storage to an encrypted storage setup or vice versa. It can also allow you to easily migrate from one encrypted KMS provider to another. In order to perform the migration, specify the configuration for both KMS providers you require, and specify the new one as the provider in the config. Specify the old one as an argument to this command `--from-provider=`.",
		RunE: func(cmd *cobra.Command, args []string) error {
			configuration := config.LoadConfiguration()
			fromConfiguration := configuration
			toConfiguration := configuration
			fromConfiguration.KeyManagement.Provider = fromKms
			toConfiguration.KeyManagement.Provider = toKms

			log := logging.NewLoggerWithConfig(configuration.Logging)

			kms, err := getKMS(log, toConfiguration)
			if err != nil {
				log.WithError(err).Fatal("failed to setup new KMS")
				return err
			}

			oldKms, err := getKMS(log, fromConfiguration)
			if err != nil {
				log.WithError(err).Fatal("failed to setup old KMS")
				return err
			}

			db, err := database.GetDatabase(log, configuration, nil)
			if err != nil {
				log.WithError(err).Fatalf("failed to initialze database")
				return errors.Wrap(err, "failed to initialize database")
			}

			txn, err := db.Begin()
			if err != nil {
				log.WithError(err).Fatal("failed to begin database transaction")
				return errors.Wrap(err, "failed to being database transaction")
			}

			offset := 0
			for {
				log.WithField("offset", offset).Trace("querying batch of 100 secrets")
				var secrets []models.Secret
				err := txn.Model(&secrets).
					Order(`secret_id ASC`).
					Limit(100).
					Offset(offset).
					Select(&secrets)
				if err != nil {
					log.WithField("offset", offset).
						WithError(err).
						Fatal("failed to retrieve batch of secrets")
					return err
				}

				for i := range secrets {
					secret := secrets[i]
					func() {
						ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
						defer cancel()

						plaintext, err := oldKms.Decrypt(ctx, secret.KeyID, secret.Version, secret.Secret)
						if err != nil {
							log.
								WithFields(logrus.Fields{
									"secretId":  secret.SecretId,
									"accountId": secret.AccountId,
									"keyId":     secret.KeyID,
									"version":   secret.Version,
								}).
								WithError(err).
								Fatal("failed to decrypt secret using old provider")
							panic(err)
						}

						newKeyId, newVersion, newCiphertext, err := kms.Encrypt(ctx, plaintext)
						if err != nil {
							log.
								WithFields(logrus.Fields{
									"secretId":  secret.SecretId,
									"accountId": secret.AccountId,
								}).
								WithError(err).
								Fatal("failed to re-encrypt secret using new provider")
							panic(err)
						}

						if dryRun {
							log.
								WithFields(logrus.Fields{
									"secretId":  secret.SecretId,
									"accountId": secret.AccountId,
									"old": logrus.Fields{
										"keyId":   secret.KeyID,
										"version": secret.Version,
									},
									"new": logrus.Fields{
										"keyId":   newKeyId,
										"version": newVersion,
									},
								}).
								Info("successfully re-encrypted secret with new kms, changes wont be persisted due to dry run")
							return
						}

						secret.KeyID = newKeyId
						secret.Version = newVersion
						secret.Secret = newCiphertext
						_, err = txn.Model(&secret).WherePK().Update(&secret)
						if err != nil {
							log.WithFields(logrus.Fields{
								"secretId":  secret.SecretId,
								"accountId": secret.AccountId,
							}).
								WithError(err).
								Fatal("failed to update secret with rotated ciphertext")
							panic(err)
						}

						log.
							WithFields(logrus.Fields{
								"secretId":  secret.SecretId,
								"accountId": secret.AccountId,
								"old": logrus.Fields{
									"keyId":   secret.KeyID,
									"version": secret.Version,
								},
								"new": logrus.Fields{
									"keyId":   newKeyId,
									"version": newVersion,
								},
							}).
							Info("successfully re-encrypted secret with new kms")
					}()
				}

				if len(secrets) < 100 {
					log.Info("no more secrets to update")
					break
				}

				offset += len(secrets)
			}

			if dryRun {
				log.Info("dry run! changes will not be persisted!")
				txn.Rollback()
				return nil
			}

			log.Info("all changes will now be committed!")
			return txn.Commit()
		},
	}

	command.PersistentFlags().StringVar(&fromKms, "from-provider", "", "Specify the KMS provider you are migrating from. Valid values are: plaintext, aws, google, vault")
	command.PersistentFlags().StringVar(&toKms, "to-provider", "", "Specify the KMS provider you are migrating to. Valid values are: plaintext, aws, google, vault")
	command.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "Don't persist the changes to the database, but still perform all the rotations in memory to ensure they all succeed.")
	command.MarkPersistentFlagRequired("from-provider")
	command.MarkPersistentFlagRequired("to-provider")

	parent.AddCommand(command)
}
