package secrets

import (
	"context"
	"strings"
	"time"

	"log/slog"

	"github.com/getsentry/sentry-go"
	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/logging"
	"github.com/pkg/errors"
)

func GetKMS(
	log *slog.Logger,
	configuration config.Configuration,
) (KeyManagement, error) {
	log.Log(context.Background(), logging.LevelTrace, "setting up key management interface")
	if configuration.KeyManagement.Provider == "" {
		return nil, errors.New("a key management provider must be specified")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var kms KeyManagement
	var err error
	switch strings.ToLower(configuration.KeyManagement.Provider) {
	case "aws":
		kmsConfig := configuration.KeyManagement.AWS
		log.Log(context.Background(), logging.LevelTrace, "using AWS KMS", "keyId", kmsConfig.KeyID)
		kms, err = NewAWSKMS(ctx, AWSKMSConfig{
			Log:       log,
			KeyID:     kmsConfig.KeyID,
			Region:    kmsConfig.Region,
			AccessKey: kmsConfig.AccessKey,
			SecretKey: kmsConfig.SecretKey,
			Endpoint:  kmsConfig.Endpoint,
		})
	case "vault":
		vaultConfig := configuration.KeyManagement.Vault
		log.Log(context.Background(), logging.LevelTrace, "using vault transit KMS", "keyId", vaultConfig.KeyID)
		log.WarnContext(context.Background(), "vault transit KMS is going to be deprecated in a future release, see guide for migrating to another KMS provider")
		kms, err = NewVaultTransit(ctx, VaultTransitConfig{
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
	case "openbao":
		openbaoConfig := configuration.KeyManagement.OpenBao
		log.Log(context.Background(), logging.LevelTrace, "using openbao transit KMS", "keyId", openbaoConfig.KeyID)
		kms, err = NewOpenBaoTransit(ctx, OpenBaoTransitConfig{
			Log:                log,
			KeyID:              openbaoConfig.KeyID,
			Address:            openbaoConfig.Endpoint,
			Role:               openbaoConfig.Role,
			AuthMethod:         openbaoConfig.AuthMethod,
			Token:              openbaoConfig.Token,
			TokenFile:          openbaoConfig.TokenFile,
			Username:           openbaoConfig.Username,
			Password:           openbaoConfig.Password,
			Timeout:            15 * time.Second,
			TLSCertificatePath: openbaoConfig.TLSCertificatePath,
			TLSKeyPath:         openbaoConfig.TLSKeyPath,
			TLSCAPath:          openbaoConfig.TLSCAPath,
			InsecureSkipVerify: openbaoConfig.InsecureSkipVerify,
			IdleConnTimeout:    15 * time.Second,
		})
	case "plaintext":
		log.Log(context.Background(), logging.LevelTrace, "using plaintext KMS, secrets will not be encrypted")
		kms = NewPlaintextKMS()
	default:
		return nil, errors.Errorf("invalid kms provider: %s", configuration.KeyManagement.Provider)
	}

	if err != nil {
		log.ErrorContext(context.Background(), "failed to configure KMS interface", "err", err)
		return nil, err
	}

	{ // Test the KMS provider
		span := sentry.StartSpan(ctx, "app.bootstrap")
		defer span.Finish()
		span.Sampled = sentry.SampledFalse
		testText := "Hello World!"
		keyID, keyVersion, cipherText, err := kms.Encrypt(span.Context(), testText)
		if err != nil {
			log.ErrorContext(context.Background(), "failed to test KMS, encryption failed; is everything configured properly?", "err", err)
			return nil, err
		}
		if cipherText == "" {
			log.ErrorContext(context.Background(), "ciphertext returned from KMS test was empty, something is very wrong!")
			return nil, errors.Errorf("ciphertext returned from KMS test was empty, something is very wrong!")
		}

		decrypted, err := kms.Decrypt(span.Context(), keyID, keyVersion, cipherText)
		if err != nil {
			log.ErrorContext(context.Background(), "failed to test KMS, decryption failed; is everything configured properly?", "err", err)
			return nil, err
		}

		if testText != decrypted {
			log.ErrorContext(context.Background(), "failed to test KMS, decrypted value is different from the original!")
			return nil, errors.New("failed to test KMS, decrypted value is different from the original!")
		}
		log.DebugContext(context.Background(), "KMS test succeeded")
	}

	return kms, nil
}
