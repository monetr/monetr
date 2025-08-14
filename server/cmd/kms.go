package main

import (
	"context"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
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

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var kms secrets.KeyManagement
	var err error
	switch strings.ToLower(configuration.KeyManagement.Provider) {
	case "aws":
		kmsConfig := configuration.KeyManagement.AWS
		log.WithFields(logrus.Fields{
			"keyId": kmsConfig.KeyID,
		}).Trace("using AWS KMS")
		kms, err = secrets.NewAWSKMS(ctx, secrets.AWSKMSConfig{
			Log:       log,
			KeyID:     kmsConfig.KeyID,
			Region:    kmsConfig.Region,
			AccessKey: kmsConfig.AccessKey,
			SecretKey: kmsConfig.SecretKey,
			Endpoint:  kmsConfig.Endpoint,
		})
	case "vault":
		vaultConfig := configuration.KeyManagement.Vault
		log.WithFields(logrus.Fields{
			"keyId": vaultConfig.KeyID,
		}).Trace("using vault transit KMS")
		kms, err = secrets.NewVaultTransit(ctx, secrets.VaultTransitConfig{
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
		kms = secrets.NewPlaintextKMS()
	default:
		return nil, errors.Errorf("invalid kms provider: %s", configuration.KeyManagement.Provider)
	}

	if err != nil {
		log.WithError(err).Fatalf("failed to configure KMS interface")
		return nil, err
	}

	{ // Test the KMS provider
		span := sentry.StartSpan(ctx, "app.bootstrap")
		defer span.Finish()
		span.Sampled = sentry.SampledFalse
		testText := "Hello World!"
		keyId, keyVersion, cipherText, err := kms.Encrypt(span.Context(), testText)
		if err != nil {
			log.WithError(err).Fatalf("failed to test KMS, encryption failed; is everything configured properly?")
			return nil, err
		}
		if len(cipherText) == 0 {
			log.Fatalf("ciphertext returned from KMS test was empty, something is very wrong!")
			return nil, errors.Errorf("ciphertext returned from KMS test was empty, something is very wrong!")
		}

		decrypted, err := kms.Decrypt(span.Context(), keyId, keyVersion, cipherText)
		if err != nil {
			log.WithError(err).Fatalf("failed to test KMS, decryption failed; is everything configured properly?")
			return nil, err
		}

		if testText != decrypted {
			log.Fatalf("failed to test KMS, decrypted value is different from the original!")
			return nil, errors.New("failed to test KMS, decrypted value is different from the original!")
		}
		log.Debug("KMS test succeeded")
	}

	return kms, nil
}
