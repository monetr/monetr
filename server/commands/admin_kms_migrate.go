package commands

import (
	"context"
	"time"

	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/database"
	"github.com/monetr/monetr/server/logging"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/secrets"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func adminKMSMigrate(parent *cobra.Command) {
	var arguments struct {
		FromKMS string
		ToKMS   string
		DryRun  bool
	}

	command := &cobra.Command{
		Use:   "migrate-kms",
		Short: "Migrate all stored secrets from one method of encryption to another.",
		Long:  "Migrate all stored secrets from one method of encryption to another. This can be used to go from plaintext secret storage to an encrypted storage setup or vice versa. It can also allow you to easily migrate from one encrypted KMS provider to another. In order to perform the migration, specify the configuration for both KMS providers you require, and specify the new one as the provider in the config. Specify the old one as an argument to this command `--from-provider=`.",
		RunE: func(cmd *cobra.Command, args []string) error {
			configuration := config.LoadConfiguration()
			fromConfiguration := configuration
			toConfiguration := configuration
			fromConfiguration.KeyManagement.Provider = arguments.FromKMS
			toConfiguration.KeyManagement.Provider = arguments.ToKMS

			log := logging.NewLoggerWithConfig(configuration.Logging)

			kms, err := secrets.GetKMS(log, toConfiguration)
			if err != nil {
				log.WithError(err).Fatal("failed to setup new KMS")
				return err
			}

			oldKms, err := secrets.GetKMS(log, fromConfiguration)
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

						if arguments.DryRun {
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

			if arguments.DryRun {
				log.Info("dry run! changes will not be persisted!")
				txn.Rollback()
				return nil
			}

			log.Info("all changes will now be committed!")
			return txn.Commit()
		},
	}

	command.PersistentFlags().StringVar(&arguments.FromKMS, "from-provider", "", "Specify the KMS provider you are migrating from. Valid values are: plaintext, aws, google, vault")
	command.PersistentFlags().StringVar(&arguments.ToKMS, "to-provider", "", "Specify the KMS provider you are migrating to. Valid values are: plaintext, aws, google, vault")
	command.PersistentFlags().BoolVar(&arguments.DryRun, "dry-run", false, "Don't persist the changes to the database, but still perform all the rotations in memory to ensure they all succeed.")

	command.MarkPersistentFlagRequired("from-provider")
	command.MarkPersistentFlagRequired("to-provider")

	parent.AddCommand(command)
}
