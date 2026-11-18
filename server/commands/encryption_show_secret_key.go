package commands

import (
	"fmt"
	"strings"

	"github.com/benbjohnson/clock"
	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/database"
	"github.com/monetr/monetr/server/encryption"
	"github.com/monetr/monetr/server/logging"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func encryptionShowSecretKey(parent *cobra.Command) {
	var arguments struct {
		SecretFile string
	}

	command := &cobra.Command{
		Use:   "show-secret-key",
		Short: "Prints the current secretbox key, decrypted with the configured key.",
		Long: strings.Join([]string{
			"Prints the current secretbox key, decrypted with the configured key.",
			"With a KMS this recovers the secretbox key, with an older secretbox key",
			"this returns the current one. The key is printed to stdout unless",
			"--secret-file is provided.",
		}, " "),
		RunE: func(cmd *cobra.Command, args []string) error {
			clock := clock.New()
			configuration := config.LoadConfiguration()
			// This only reads keys, never migrate the database from here.
			configuration.PostgreSQL.Migrate = false
			log := logging.NewLoggerWithConfig(configuration.Logging)

			db, err := database.GetDatabase(log, configuration)
			if err != nil {
				log.Error("failed to setup database connection", "err", err)
				return err
			}
			defer db.Close()

			secret, err := encryption.CurrentSecretboxKey(
				cmd.Context(),
				log,
				clock,
				configuration,
				db,
			)
			if errors.Is(err, encryption.ErrEncryptionNotInitialized) {
				return errors.New("encryption has not been set up, run monetr encryption init first")
			} else if err != nil {
				return err
			}

			if arguments.SecretFile != "" {
				if err := encryption.WriteSecretboxKeyFile(
					arguments.SecretFile,
					secret,
				); err != nil {
					return errors.Wrapf(err, "failed to write secretbox key to path: %q", arguments.SecretFile)
				}
				log.InfoContext(cmd.Context(),
					"secretbox key has been written to disk",
					"path", arguments.SecretFile,
				)
				return nil
			}

			// Logs go to stderr, so stdout only ever has the key on it and can be
			// redirected or piped.
			_, err = fmt.Fprintln(cmd.OutOrStdout(), secret)
			return err
		},
	}

	command.PersistentFlags().StringVarP(&arguments.SecretFile, "secret-file", "f", "", "Write the secretbox key to the specified file on the disk instead of printing it, this file must not already exist.")

	parent.AddCommand(command)
}
