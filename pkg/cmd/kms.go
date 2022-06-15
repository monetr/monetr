package main

import (
	"context"
	"strings"

	"github.com/monetr/monetr/pkg/config"
	"github.com/monetr/monetr/pkg/secrets"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func getKMS(log *logrus.Entry, configuration config.Configuration) (secrets.KeyManagement, error) {
	if !configuration.KeyManagement.Enabled {
		log.Trace("key management is not enabled, it will not be initialized")
		return nil, nil
	}

	log.Trace("setting up key management interface")

	if configuration.KeyManagement.Provider == "" {
		return nil, errors.New("key management is enabled by not provider is configured")
	}

	var kms secrets.KeyManagement
	var err error
	switch strings.ToLower(configuration.KeyManagement.Provider) {
	case "aws":
		kmsConfig := configuration.KeyManagement.AWS
		log.Trace("using AWS KMS")
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
		log.Trace("using Google KMS")
		kms, err = secrets.NewGoogleKMS(context.Background(), secrets.GoogleKMSConfig{
			Log:             log,
			KeyName:         kmsConfig.ResourceName,
			URL:             nil,
			APIKey:          nil,
			CredentialsFile: kmsConfig.CredentialsJSON,
		})
	default:
		return nil, errors.Errorf("invalid kms provider: %s", configuration.KeyManagement.Provider)
	}

	if err != nil {
		log.WithError(err).Fatalf("failed to configure KMS interface")
		return nil, err
	}

	return kms, nil
}
