package secrets

import (
	"context"
	"os"
	"testing"

	"github.com/monetr/monetr/pkg/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	GoogleCredentialsEnv = "TEST_GOOGLE_KMS_CREDENTIALS"
	GoogleKeyName        = "TEST_GOOGLE_KMS_NAME"
)

func MustHaveLegitimateCredentials(t *testing.T) (credentialsFile, keyName string) {
	log := testutils.GetLog(t)
	log.Info("this test requires two environment variables to be set as it requires credentials to talk to Google's API for KMS")
	credentialsFile = os.Getenv(GoogleCredentialsEnv)
	if _, err := os.Stat(credentialsFile); err != nil {
		t.Logf("%s: %s | could not read credentials file", GoogleCredentialsEnv, credentialsFile)
		t.Skip("the google service account credentials file could not be read")
		return
	}

	keyName = os.Getenv(GoogleKeyName)
	if keyName == "" {
		log.WithField(GoogleKeyName, keyName).Fatal("no key name was provided")
		return
	}

	return credentialsFile, keyName
}

func TestNewGoogleKMS(t *testing.T) {
	t.Run("with legitimate credentials", func(t *testing.T) {
		log := testutils.GetLog(t)
		credentialsFile, keyName := MustHaveLegitimateCredentials(t)

		kms, err := NewGoogleKMS(context.Background(), GoogleKMSConfig{
			Log:             log,
			KeyName:         keyName,
			CredentialsFile: &credentialsFile,
		})
		assert.NoError(t, err, "must not return an error when creating the KMS interface")
		assert.NotNil(t, kms, "the KMS interface returned must be valid")
	})
}

func TestGoogleKMS_Encrypt(t *testing.T) {
	t.Run("with legitimate credentials", func(t *testing.T) {
		log := testutils.GetLog(t)
		credentialsFile, keyName := MustHaveLegitimateCredentials(t)

		kms, err := NewGoogleKMS(context.Background(), GoogleKMSConfig{
			Log:             log,
			KeyName:         keyName,
			CredentialsFile: &credentialsFile,
		})
		require.NoError(t, err, "must not return an error when creating the KMS interface")
		require.NotNil(t, kms, "the KMS interface returned must be valid")

		input := []byte("i am a little teapot")
		keyId, version, data, err := kms.Encrypt(context.Background(), input)
		assert.NoError(t, err, "should not return an error when encrypting")
		assert.Equal(t, keyName, keyId, "should have the same key Id as the configuration specified")
		assert.NotEmpty(t, version, "should contain a version")
		assert.NotEmpty(t, data, "some data should have been returned")
		assert.NotEqual(t, input, data, "the returned data should not be the unencrypted input")
	})
}

func TestGoogleKMS_Dencrypt(t *testing.T) {
	t.Run("with legitimate credentials", func(t *testing.T) {
		log := testutils.GetLog(t)
		credentialsFile, keyName := MustHaveLegitimateCredentials(t)

		kms, err := NewGoogleKMS(context.Background(), GoogleKMSConfig{
			Log:             log,
			KeyName:         keyName,
			CredentialsFile: &credentialsFile,
		})
		require.NoError(t, err, "must not return an error when creating the KMS interface")
		require.NotNil(t, kms, "the KMS interface returned must be valid")

		input := []byte("i am a little teapot")
		keyId, version, data, err := kms.Encrypt(context.Background(), input)
		require.NoError(t, err, "should not return an error when encrypting")
		require.Equal(t, keyName, keyId, "should have the same key Id as the configuration specified")
		require.NotEmpty(t, version, "should contain a version")
		require.NotEmpty(t, data, "some data should have been returned")
		require.NotEqual(t, input, data, "the returned data should not be the unencrypted input")

		decrypted, err := kms.Decrypt(context.Background(), keyId, version, data)
		assert.NoError(t, err, "must be able to decrypt the data")
		assert.Equal(t, input, decrypted, "the decrypted value should match the original input")
	})
}
