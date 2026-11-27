package commands

import (
	"context"

	"github.com/benbjohnson/clock"
	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/database"
	"github.com/monetr/monetr/server/encryption"
	"github.com/monetr/monetr/server/logging"
	"github.com/monetr/monetr/server/models"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func encryptionInit(parent *cobra.Command) {
	var arguments struct {
		MigrateDatabase bool
		SecretFile      string
		PrintSecret     bool
	}

	command := &cobra.Command{
		Use:   "init",
		Short: "Initializes monetr's secret key, root key, and key encryption key.",
		RunE: func(cmd *cobra.Command, args []string) error {
			clock := clock.New()
			configuration := config.LoadConfiguration()
			log := logging.NewLoggerWithConfig(configuration.Logging)

			// Check before touching the database. With a KMS the secret can be
			// recovered later through the KMS, so discarding it is fine. With
			// secretbox the secret is the only way to unseal, so it has to go
			// somewhere.
			switch models.RootKeyKind(configuration.Encryption.Kind) {
			case models.RootKeyKindSecretBox:
				if arguments.SecretFile == "" && !arguments.PrintSecret {
					return errors.New("encryption kind is secretbox, so --secret-file or --print-secret is required, otherwise the generated secret would be lost")
				}
			case models.RootKeyKindAWS, models.RootKeyKindOpenBao:
			default:
				return errors.Errorf("unknown encryption kind %q, must be secretbox, aws, or openbao", configuration.Encryption.Kind)
			}

			db, err := database.GetDatabase(log, configuration)
			if err != nil {
				log.Error("failed to setup database connection", "err", err)
				return err
			}
			defer db.Close()

			return encryption.Initialize(
				cmd.Context(),
				log,
				clock,
				configuration.Encryption,
				db,
				func(ctx context.Context, secret string) error {
					if arguments.SecretFile != "" {
						if err := encryption.WriteSecretboxKeyFile(
							arguments.SecretFile,
							secret,
						); err != nil {
							return errors.Wrapf(err, "failed to write secretbox key to path: %q", arguments.SecretFile)
						}
					}

					if arguments.PrintSecret {
						// TODO This goes through the logger, so with a log level above info
						// the secret is never printed, and with only --print-secret it would
						// be lost. It also ends up wherever logs are shipped. Print it to
						// stdout instead (cmd.OutOrStdout()).
						log.InfoContext(ctx,
							"Generated secretbox key for monetr, please store this key safely!",
							"secret", secret,
						)
					}

					return nil
				},
			)
		},
	}

	command.PersistentFlags().BoolVarP(&arguments.MigrateDatabase, "migrate", "m", false, "Automatically run database migrations on startup. Defaults to: false")
	command.PersistentFlags().BoolVarP(&arguments.PrintSecret, "print-secret", "p", false, "Prints the generated secret that monetr can use to unseal its encrypted data. Store the secret safely. Defaults to: false")
	command.PersistentFlags().StringVarP(&arguments.SecretFile, "secret-file", "f", "", "Write the generated secret to the specified file on the disk, this file must not already exist.")

	viper.BindPFlag("PostgreSQL.Migrate", command.PersistentFlags().Lookup("migrate"))

	parent.AddCommand(command)
}
