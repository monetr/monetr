package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/monetr/monetr/server/config"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	rootCommand = &cobra.Command{
		Use:   "monetr",
		Short: "monetr's budgeting application",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
)

func init() {
	rootCommand.PersistentFlags().StringVarP(&config.LogLevel, "log-level", "L", "info", "Specify the log level to use, allowed values: trace, debug, info, warn, error, fatal")
	rootCommand.PersistentFlags().StringVarP(&config.FilePath, "config", "c", "", "Specify the config file to use.")
	viper.BindPFlag("Logging.Level", rootCommand.PersistentFlags().Lookup("log-level"))
	viper.BindPFlag("configFile", rootCommand.PersistentFlags().Lookup("config"))
	newVersionCommand(rootCommand)
	newNoticesCommand(rootCommand)
}

func loadCertificates(configuration config.Configuration, generateCertificates bool) (ed25519.PublicKey, ed25519.PrivateKey, error) {
	// TODO Add support for both the public and private key being in the same file!
	var publicKey ed25519.PublicKey
	var privateKey ed25519.PrivateKey
	var ok bool

	{ // Parse the private key
		keyBytes, err := ioutil.ReadFile(configuration.Security.PrivateKey)
		if os.IsNotExist(err) {
			directory, err := filepath.Abs(path.Dir(configuration.Security.PrivateKey))
			if err != nil {
				return nil, nil, errors.Wrap(err, "public key directory is not valid")
			}
			if err := os.MkdirAll(directory, 0755); err != nil {
				return nil, nil, errors.Wrap(err, "failed to create directory for certificates")
			}

			publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
			if err != nil {
				return nil, nil, errors.Wrap(err, "failed to generate certificate")
			}

			privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
			if err != nil {
				return nil, nil, errors.Wrap(err, "failed to marshal private key")
			}
			privateKeyBlock := &pem.Block{
				Type:  "PRIVATE KEY",
				Bytes: privateKeyBytes,
			}
			privateKeyPem := pem.EncodeToMemory(privateKeyBlock)

			if err := os.WriteFile(configuration.Security.PrivateKey, privateKeyPem, 0644); err != nil {
				return nil, nil, errors.Wrap(err, "failed to write private key")
			}

			return publicKey, privateKey, nil
		} else if err != nil {
			return nil, nil, errors.Wrap(err, "unable to read public key")
		}

		keyBlock, _ := pem.Decode(keyBytes)
		if keyBlock == nil {
			return nil, nil, errors.New("failed to decode PEM block containing private key")
		}

		key, err := x509.ParsePKCS8PrivateKey(keyBlock.Bytes)
		if err != nil {
			return nil, nil, errors.Wrap(err, "failed to parse private key")
		}

		privateKey, ok = key.(ed25519.PrivateKey)
		if !ok {
			return nil, nil, errors.New("provided private key is not an ED25519 private key")
		}

		publicKey, ok = privateKey.Public().(ed25519.PublicKey)
		if !ok {
			return nil, nil, errors.New("provided public key is not an ED25519 public key")
		}

		return publicKey, privateKey, nil
	}
}
