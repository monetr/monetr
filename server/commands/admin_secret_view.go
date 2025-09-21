package commands

import (
	"fmt"

	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/database"
	"github.com/monetr/monetr/server/logging"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/secrets"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func adminSecretView(parent *cobra.Command) {
	var arguments struct {
		ItemID string
		KMS    string
	}
	command := &cobra.Command{
		Use:   "secret:view",
		Short: "Retrieve a secret's value from the data store, meant for debugging purposes only!",
		RunE: func(cmd *cobra.Command, args []string) error {
			configuration := config.LoadConfiguration()
			if arguments.KMS != "" {
				configuration.KeyManagement.Provider = arguments.KMS
			}

			log := logging.NewLoggerWithConfig(configuration.Logging)

			db, err := database.GetDatabase(log, configuration, nil)
			if err != nil {
				log.WithError(err).Fatalf("failed to initialze database")
				return errors.Wrap(err, "failed to initialize database")
			}

			kms, err := secrets.GetKMS(log, configuration)
			if err != nil {
				log.WithError(err).Fatal("failed to setup KMS")
				return err
			}

			var token models.PlaidToken
			err = db.Model(&token).
				Where(`"item_id" = ?`, arguments.KMS).
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

	command.PersistentFlags().StringVar(&arguments.ItemID, "item-id", "", "The Plaid Item ID to retrieve the secret for.")
	command.PersistentFlags().StringVar(&arguments.KMS, "kms-provider", "", "Override the KMS provider setting in the config.")

	command.MarkPersistentFlagRequired("item-id")

	parent.AddCommand(command)
}
