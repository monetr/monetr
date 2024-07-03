package main

import (
	"context"
	"strings"
	"time"

	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/secrets"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func getKMS(log *logrus.Entry, configuration config.Configuration) (secrets.KeyManagement, error) {
	log.Trace("setting up key management interface")
	if configuration.KeyManagement.Provider == "" {
		return nil, errors.New("a key management provider must be specified")
	}

	var kms secrets.KeyManagement
	var err error
	switch strings.ToLower(configuration.KeyManagement.Provider) {
	case "aws":
		kmsConfig := configuration.KeyManagement.AWS
		log.WithFields(logrus.Fields{
			"keyId": kmsConfig.KeyID,
		}).Trace("using AWS KMS")
		kms, err = secrets.NewAWSKMS(context.Background(), secrets.AWSKMSConfig{
			Log:       log,
			KeyID:     kmsConfig.KeyID,
			Region:    kmsConfig.Region,
			AccessKey: kmsConfig.AccessKey,
			SecretKey: kmsConfig.SecretKey,
			Endpoint:  kmsConfig.Endpoint,
		})
	case "google":
		kmsConfig := configuration.KeyManagement.Google
		log.WithFields(logrus.Fields{
			"keyName": kmsConfig.ResourceName,
		}).Trace("using Google KMS")
		kms, err = secrets.NewGoogleKMS(context.Background(), secrets.GoogleKMSConfig{
			Log:             log,
			KeyName:         kmsConfig.ResourceName,
			URL:             nil,
			APIKey:          nil,
			CredentialsFile: kmsConfig.CredentialsJSON,
		})
	case "vault":
		vaultConfig := configuration.KeyManagement.Vault
		log.WithFields(logrus.Fields{
			"keyId": vaultConfig.KeyID,
		}).Trace("using vault transit KMS")
		kms, err = secrets.NewVaultTransit(context.Background(), secrets.VaultTransitConfig{
			Log:                log,
			KeyID:              vaultConfig.KeyID,
			Address:            vaultConfig.Endpoint,
			Role:               vaultConfig.Role,
			AuthMethod:         vaultConfig.AuthMethod,
			Token:              vaultConfig.Token,
			TokenFile:          vaultConfig.TokenFile,
			Username:           vaultConfig.Username,
			Password:           vaultConfig.Password,
			Timeout:            15 * time.Second,
			TLSCertificatePath: vaultConfig.TLSCertificatePath,
			TLSKeyPath:         vaultConfig.TLSKeyPath,
			TLSCAPath:          vaultConfig.TLSCAPath,
			InsecureSkipVerify: vaultConfig.InsecureSkipVerify,
			IdleConnTimeout:    15 * time.Second,
		})
	case "plaintext":
		log.Trace("using plaintext KMS, secrets will not be encrypted")
		return secrets.NewPlaintextKMS(), nil
	default:
		return nil, errors.Errorf("invalid kms provider: %s", configuration.KeyManagement.Provider)
	}

	if err != nil {
		log.WithError(err).Fatalf("failed to configure KMS interface")
		return nil, err
	}

	return kms, nil
}
