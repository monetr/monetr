package secrets

import (
	"context"
	"os"
	"testing"

	"github.com/monetr/monetr/pkg/internal/testutils"
	"github.com/stretchr/testify/assert"
)

const (
	GoogleCredentialsEnv = "TEST_GOOGLE_KMS_CREDENTIALS"
	GoogleKeyName        = "TEST_GOOGLE_KMS_NAME"
)

func TestNewGoogleKMS(t *testing.T) {
	t.Run("with service account credentials", func(t *testing.T) {
		log := testutils.GetLog(t)
		log.Info("this test requires two environment variables to be set as it requires credentials to talk to Google's API for KMS")

		credentialsFile := os.Getenv(GoogleCredentialsEnv)
		if _, err := os.Stat(credentialsFile); err != nil {
			log.WithError(err).WithField(GoogleCredentialsEnv, credentialsFile).Error("could not read credentials file")
			t.Skip("the google service account credentials file could not be read")
			return
		}

		keyName := os.Getenv(GoogleKeyName)
		if keyName == "" {
			log.WithField(GoogleKeyName, keyName).Fatal("no key name was provided")
			return
		}

		kms, err := NewGoogleKMS(context.Background(), GoogleKMSConfig{
			Log:             log,
			KeyName:         keyName,
			CredentialsFile: &credentialsFile,
		})
		assert.NoError(t, err, "must not return an error when creating the KMS interface")
		assert.NotNil(t, kms, "the KMS interface returned must be valid")
	})
}
