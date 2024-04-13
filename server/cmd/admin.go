package main

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"

	"github.com/benbjohnson/clock"
	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/logging"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/platypus"
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
	var linkId uint64
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

			if linkId == 0 {
				log.Fatal("link ID must be specified via --link")
				return cmd.Help()
			}

			db, err := getDatabase(log, configuration, nil)
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
	command.PersistentFlags().Uint64Var(&linkId, "link", 0, "Link Id to trigger the Plaid refresh on")

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

	parent.AddCommand(command)
}

func newViewSecretCommand(parent *cobra.Command) {
	var itemId string
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
			fmt.Printf("Key ID: %s\n", keyId)
			fmt.Printf("Version: %s\n", version)
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

	parent.AddCommand(command)
}
