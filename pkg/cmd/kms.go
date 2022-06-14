package main

import (
	"context"

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

	{ // Assert that the configuration only has a single provider setup.
		count := 0
		for _, provider := range []interface{}{
			configuration.KeyManagement.AWS,
			configuration.KeyManagement.Google,
		} {
			if provider != nil {
				count++
			}
		}

		switch count {
		case 0:
			return nil, errors.New("key management is enabled by not provider is configured")
		case 1:
			break
		default:
			return nil, errors.New("you can only have one key management provider configured at a time")
		}
	}

	var kms secrets.KeyManagement
	var err error
	if kmsConfig := configuration.KeyManagement.AWS; kmsConfig != nil {
		log.Trace("using AWS KMS")
		kms, err = secrets.NewAWSKMS(context.Background(), secrets.AWSKMSConfig{
			Log:       log,
			KeyID:     kmsConfig.KeyID,
			Region:    kmsConfig.Region,
			AccessKey: kmsConfig.AccessKey,
			SecretKey: kmsConfig.SecretKey,
			Endpoint:  kmsConfig.Endpoint,
		})
	} else if kmsConfig := configuration.KeyManagement.Google; kmsConfig != nil {
		log.Trace("using Google KMS")
		kms, err = secrets.NewGoogleKMS(context.Background(), secrets.GoogleKMSConfig{
			Log:             log,
			KeyName:         kmsConfig.ResourceName,
			URL:             nil,
			APIKey:          nil,
			CredentialsFile: kmsConfig.CredentialsJSON,
		})
	}

	if err != nil {
		log.WithError(err).Fatalf("failed to configure KMS interface")
		return nil, err
	}

	return kms, nil
}
