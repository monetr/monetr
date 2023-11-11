package main

import (
	"crypto/ed25519"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"

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

func loadCertificates(configuration config.Configuration) (ed25519.PublicKey, ed25519.PrivateKey, error) {
	// TODO Add support for both the public and private key being in the same file!
	var publicKey ed25519.PublicKey
	var privateKey ed25519.PrivateKey
	var ok bool

	{ // Parse the public key
		keyBytes, err := ioutil.ReadFile(configuration.Security.PublicKey)
		if err != nil {
			return nil, nil, errors.Wrap(err, "unable to read public key")
		}

		keyBlock, _ := pem.Decode(keyBytes)
		if keyBlock == nil {
			return nil, nil, errors.New("failed to decode PEM block containing public key")
		}

		key, err := x509.ParsePKIXPublicKey(keyBlock.Bytes)
		if err != nil {
			return nil, nil, errors.Wrap(err, "failed to parse public key")
		}

		publicKey, ok = key.(ed25519.PublicKey)
		if !ok {
			return nil, nil, errors.New("provided public key is not an ED25519 public key")
		}
	}

	{ // Parse the private key
		keyBytes, err := ioutil.ReadFile(configuration.Security.PrivateKey)
		if err != nil {
			return nil, nil, errors.Wrap(err, "unable to read private key")
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
	}

	return publicKey, privateKey, nil
}
