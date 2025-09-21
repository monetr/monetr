package commands

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"

	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/logging"
	"github.com/monetr/monetr/server/secrets"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func adminKMSCheck(parent *cobra.Command) {
	var arguments struct {
		KMS string
	}

	command := &cobra.Command{
		Use:   "kms:check",
		Short: "Tests the configured KMS provider to make sure data can be encrypted and decrypted",
		RunE: func(cmd *cobra.Command, args []string) error {
			configuration := config.LoadConfiguration()
			if arguments.KMS != "" {
				configuration.KeyManagement.Provider = arguments.KMS
			}

			log := logging.NewLoggerWithConfig(configuration.Logging)

			kms, err := secrets.GetKMS(log, configuration)
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

			keyID, version, result, err := kms.Encrypt(context.Background(), testString)
			if err != nil {
				log.WithError(err).Error("failed to encrypt test string")
				return err
			}

			fmt.Printf("Successfully encrypted test string!\n")
			fmt.Printf("Key ID: %+v\n", keyID)
			fmt.Printf("Version: %+v\n", version)
			fmt.Printf("Result Binary -----------\n")
			fmt.Println()
			fmt.Println(hex.Dump([]byte(result)))
			fmt.Println()
			fmt.Printf("Result Formatted: %s\n", result)
			fmt.Println()
			fmt.Println("Testing decryption with test string...")

			decrypted, err := kms.Decrypt(context.Background(), keyID, version, result)
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

	command.PersistentFlags().StringVar(&arguments.KMS, "kms-provider", "", "Specify the provider to use for KMS testing.")

	parent.AddCommand(command)
}
