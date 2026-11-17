package commands

import (
	"context"

	"github.com/benbjohnson/clock"
	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/database"
	"github.com/monetr/monetr/server/encryption"
	"github.com/monetr/monetr/server/logging"
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
					if arguments.PrintSecret {
						log.InfoContext(ctx,
							"Generated secretbox key for monetr, please store this key safely!",
							"secret", secret,
						)
					}

					if arguments.SecretFile != "" {
						// TODO Write file to disk with the correct permissions, also
						// validate the file path is valid?
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
